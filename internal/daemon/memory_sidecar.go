package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
)

const (
	MemoryProviderDisabled = "disabled"
	MemoryProviderLocal    = "local"
	MemoryProviderDegraded = "degraded"
)

const (
	MemoryReasonNoLocalMemory = "no_local_memory"
	MemoryReasonBundleInvalid = "bundle_invalid"
	MemoryReasonQueryEmpty    = "query_empty"
)

const (
	MemoryRecallOutcomeMatched = "matched"
	MemoryRecallOutcomeNoData  = "no_data"
	MemoryRecallOutcomeError   = "error"
)

type MemoryBundleStatus struct {
	Version   string     `json:"version,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Dir       string     `json:"dir,omitempty"`
}

type MemorySidecarStatus struct {
	Provider          string              `json:"provider"`
	Ready             bool                `json:"ready"`
	Reason            string              `json:"reason,omitempty"`
	Bundle            *MemoryBundleStatus `json:"bundle,omitempty"`
	FallbackAvailable bool                `json:"fallback_available"`
	LocalFacts        int                 `json:"local_facts"`
}

type MemoryRecallRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type MemoryRecallMatch struct {
	Category string  `json:"category"`
	Text     string  `json:"text"`
	Entry    string  `json:"entry"`
	Line     int     `json:"line"`
	Score    float64 `json:"score"`
}

type MemoryRecallResult struct {
	Attempted       bool                `json:"attempted"`
	Provider        string              `json:"provider"`
	Outcome         string              `json:"outcome"`
	Reason          string              `json:"reason,omitempty"`
	ContextInjected bool                `json:"context_injected"`
	Results         []MemoryRecallMatch `json:"results"`
	Bundle          *MemoryBundleStatus `json:"bundle,omitempty"`
}

type memoryBundleManifest struct {
	BundleVersion   string `json:"bundle_version"`
	BundleCreatedAt string `json:"bundle_created_at"`
	Version         string `json:"version"`
	CreatedAt       string `json:"created_at"`
}

func (s *Server) memorySidecarStatus() MemorySidecarStatus {
	memoryDir := s.memoryDir()
	if memoryDir == "" {
		return MemorySidecarStatus{Provider: MemoryProviderDisabled, Reason: MemoryReasonNoLocalMemory}
	}
	facts := s.localMemoryFacts()
	bundle, bundleErr := readCurrentMemoryBundle(memoryDir)
	if bundleErr != nil {
		return MemorySidecarStatus{
			Provider:          MemoryProviderDegraded,
			Ready:             false,
			Reason:            MemoryReasonBundleInvalid,
			FallbackAvailable: len(facts) > 0,
			LocalFacts:        len(facts),
		}
	}
	if bundle != nil || len(facts) > 0 {
		return MemorySidecarStatus{
			Provider:          MemoryProviderLocal,
			Ready:             true,
			Bundle:            bundle,
			FallbackAvailable: len(facts) > 0,
			LocalFacts:        len(facts),
		}
	}
	return MemorySidecarStatus{Provider: MemoryProviderDisabled, Reason: MemoryReasonNoLocalMemory}
}

func (s *Server) recallMemory(query string) MemoryRecallResult {
	query = strings.TrimSpace(query)
	status := s.memorySidecarStatus()
	result := MemoryRecallResult{
		Attempted: true,
		Provider:  status.Provider,
		Bundle:    status.Bundle,
		Results:   []MemoryRecallMatch{},
	}
	if query == "" {
		result.Outcome = MemoryRecallOutcomeNoData
		result.Reason = MemoryReasonQueryEmpty
		return result
	}
	facts := s.localMemoryFacts()
	if len(facts) == 0 {
		result.Outcome = MemoryRecallOutcomeNoData
		result.Reason = MemoryReasonNoLocalMemory
		return result
	}
	matches := rankMemoryFacts(query, facts)
	if len(matches) == 0 {
		result.Outcome = MemoryRecallOutcomeNoData
		result.Reason = MemoryRecallOutcomeNoData
		return result
	}
	result.Outcome = MemoryRecallOutcomeMatched
	result.Results = matches
	return result
}

func (s *Server) localMemoryFacts() []memoryFactView {
	view, err := s.buildMemoryView()
	if err != nil {
		return nil
	}
	return append([]memoryFactView(nil), view.Facts...)
}

func readCurrentMemoryBundle(memoryDir string) (*MemoryBundleStatus, error) {
	current := filepath.Join(memoryDir, "current")
	target, err := os.Readlink(current)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(memoryDir, target)
	}
	cleanMemory := filepath.Clean(memoryDir)
	cleanTarget := filepath.Clean(target)
	if !strings.HasPrefix(cleanTarget, cleanMemory+string(filepath.Separator)) {
		return nil, os.ErrPermission
	}
	data, err := os.ReadFile(filepath.Join(cleanTarget, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var manifest memoryBundleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	version := firstNonEmpty(manifest.BundleVersion, manifest.Version)
	created := firstNonEmpty(manifest.BundleCreatedAt, manifest.CreatedAt)
	var createdAt *time.Time
	if created != "" {
		if parsed, err := time.Parse(time.RFC3339, created); err == nil {
			createdAt = &parsed
		}
	}
	return &MemoryBundleStatus{Version: version, CreatedAt: createdAt, Dir: cleanTarget}, nil
}

func rankMemoryFacts(query string, facts []memoryFactView) []MemoryRecallMatch {
	terms := memoryRecallTerms(query)
	matches := make([]MemoryRecallMatch, 0)
	for _, fact := range facts {
		score := memoryFactScore(terms, fact)
		if score <= 0 {
			continue
		}
		matches = append(matches, MemoryRecallMatch{
			Category: fact.Category,
			Text:     fact.Text,
			Entry:    fact.Entry,
			Line:     fact.Line,
			Score:    score,
		})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].Line < matches[j].Line
	})
	if len(matches) > 5 {
		matches = matches[:5]
	}
	return matches
}

func memoryRecallTerms(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	out := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.Trim(word, ".,;:!?()[]{}\"'")
		if len(word) < 3 {
			continue
		}
		out = append(out, word)
	}
	return out
}

func memoryFactScore(terms []string, fact memoryFactView) float64 {
	if len(terms) == 0 {
		return 0
	}
	haystack := strings.ToLower(fact.Category + " " + fact.Subject + " " + fact.Text)
	score := 0.0
	for _, term := range terms {
		if strings.Contains(haystack, term) {
			score += 1
		}
	}
	return score
}

func privateMemoryBlock(result MemoryRecallResult) string {
	if result.Outcome != MemoryRecallOutcomeMatched || len(result.Results) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<private_memory>\n")
	for _, match := range result.Results {
		sb.WriteString("- [")
		sb.WriteString(match.Category)
		sb.WriteString("] ")
		sb.WriteString(match.Text)
		sb.WriteString("\n")
	}
	sb.WriteString("</private_memory>")
	return sb.String()
}

func memoryPreflightEventData(result MemoryRecallResult) map[string]any {
	return map[string]any{
		"attempted":        result.Attempted,
		"provider":         result.Provider,
		"outcome":          result.Outcome,
		"reason":           result.Reason,
		"results_count":    len(result.Results),
		"context_injected": result.ContextInjected,
		"bundle_ready":     result.Bundle != nil,
	}
}

type daemonMemoryPreflightProvider struct {
	server *Server
	runID  string
}

func (s *Server) memoryPreflightProvider(runID string) agent.MemoryPreflightProvider {
	return daemonMemoryPreflightProvider{server: s, runID: runID}
}

func (p daemonMemoryPreflightProvider) PreflightMemory(_ context.Context, query string) (agent.MemoryPreflightResult, error) {
	result := p.server.recallMemory(query)
	block := privateMemoryBlock(result)
	return agent.MemoryPreflightResult{
		Block:           block,
		Attempted:       result.Attempted,
		Provider:        result.Provider,
		Outcome:         result.Outcome,
		Reason:          result.Reason,
		ResultsCount:    len(result.Results),
		ContextInjected: block != "",
	}, nil
}
