package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Candidate identifies one local session that should be considered for sync.
type Candidate struct {
	Dir       string
	AgentName string
	SessionID string
	UpdatedAt time.Time
	Source    string
}

// ScannerDeps carries filesystem roots for candidate discovery.
type ScannerDeps struct {
	StarclawDir string
}

type sessionDir struct {
	Dir       string
	AgentName string
}

// DiscoverCandidates enumerates StarClaw session JSON files and returns local
// candidates newer than marker.LastSyncAt plus due transient retries.
func DiscoverCandidates(ctx context.Context, deps ScannerDeps, cfg Config, marker Marker, now time.Time) ([]Candidate, int, error) {
	dirs, err := discoverSessionDirs(deps.StarclawDir)
	if err != nil {
		return nil, 0, err
	}

	byID := map[string]Candidate{}
	skipped := 0
	for _, sd := range dirs {
		if err := ctx.Err(); err != nil {
			return nil, skipped, err
		}
		entries, err := os.ReadDir(sd.Dir)
		if err != nil {
			skipped++
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			path := filepath.Join(sd.Dir, entry.Name())
			c, err := candidateFromFile(path, sd)
			if err != nil {
				skipped++
				continue
			}
			if !c.UpdatedAt.After(marker.LastSyncAt) {
				continue
			}
			if fe, failed := marker.Failed[c.SessionID]; failed && fe.Category == CategoryPermanent {
				if !c.UpdatedAt.After(fe.LastObservedUpdatedAt) {
					continue
				}
			}
			putFreshest(byID, c)
		}
	}

	for id, fe := range marker.Failed {
		if fe.Category != CategoryTransient || fe.NextAttemptAt == nil || now.Before(*fe.NextAttemptAt) {
			continue
		}
		if _, exists := byID[id]; exists {
			continue
		}
		if c, found := locateSession(dirs, id); found {
			putFreshest(byID, c)
		} else {
			byID[id] = Candidate{SessionID: id, UpdatedAt: fe.LastAttemptAt}
		}
	}

	out := filterCandidates(byID, cfg)
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].SessionID < out[j].SessionID
		}
		return out[i].UpdatedAt.Before(out[j].UpdatedAt)
	})
	return out, skipped, nil
}

func discoverSessionDirs(starclawDir string) ([]sessionDir, error) {
	if strings.TrimSpace(starclawDir) == "" {
		return nil, fmt.Errorf("sync: starclaw dir is required")
	}
	root := filepath.Join(starclawDir, "sessions")
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat sessions dir: %w", err)
	}

	dirs := []sessionDir{{Dir: root}}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read sessions dir: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirs = append(dirs, sessionDir{
			Dir:       filepath.Join(root, entry.Name()),
			AgentName: entry.Name(),
		})
	}
	return dirs, nil
}

func candidateFromFile(path string, sd sessionDir) (Candidate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Candidate{}, fmt.Errorf("read session candidate: %w", err)
	}
	var probe struct {
		ID        string    `json:"id"`
		UpdatedAt time.Time `json:"updated_at"`
		Source    string    `json:"source"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return Candidate{}, fmt.Errorf("parse session candidate: %w", err)
	}
	if probe.UpdatedAt.IsZero() {
		return Candidate{}, fmt.Errorf("parse session candidate: updated_at is required")
	}
	id := probe.ID
	if id == "" {
		id = strings.TrimSuffix(filepath.Base(path), ".json")
	}
	return Candidate{
		Dir:       sd.Dir,
		AgentName: sd.AgentName,
		SessionID: id,
		UpdatedAt: probe.UpdatedAt,
		Source:    probe.Source,
	}, nil
}

func putFreshest(byID map[string]Candidate, c Candidate) {
	if c.SessionID == "" {
		return
	}
	existing, ok := byID[c.SessionID]
	if !ok || c.UpdatedAt.After(existing.UpdatedAt) {
		byID[c.SessionID] = c
	}
}

func locateSession(dirs []sessionDir, id string) (Candidate, bool) {
	for _, sd := range dirs {
		path := filepath.Join(sd.Dir, id+".json")
		c, err := candidateFromFile(path, sd)
		if err == nil {
			return c, true
		}
	}
	return Candidate{}, false
}

func filterCandidates(byID map[string]Candidate, cfg Config) []Candidate {
	excludeAgents := makeStringSet(cfg.ExcludeAgents)
	excludeSources := makeStringSet(cfg.ExcludeSources)
	out := make([]Candidate, 0, len(byID))
	for _, c := range byID {
		agentKey := c.AgentName
		if agentKey == "" {
			agentKey = "default"
		}
		if excludeAgents[agentKey] {
			continue
		}
		sourceKey := c.Source
		if sourceKey == "" {
			sourceKey = "local"
		}
		if excludeSources[sourceKey] {
			continue
		}
		out = append(out, c)
	}
	return out
}

func makeStringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
