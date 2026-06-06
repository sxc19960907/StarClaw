package daemon

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	ctxmem "github.com/starclaw/starclaw/internal/context"
)

type memoryEntryView struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	Primary  bool      `json:"primary"`
}

type memoryFactView struct {
	Category string `json:"category"`
	Text     string `json:"text"`
	Entry    string `json:"entry"`
	Line     int    `json:"line"`
	Subject  string `json:"subject,omitempty"`
}

type memoryWarningView struct {
	Type     string   `json:"type"`
	Subject  string   `json:"subject,omitempty"`
	Message  string   `json:"message"`
	Lines    []int    `json:"lines,omitempty"`
	Entries  []string `json:"entries,omitempty"`
	Category string   `json:"category,omitempty"`
}

type memoryView struct {
	MemoryDir  string              `json:"memory_dir"`
	Entries    []memoryEntryView   `json:"entries"`
	Content    string              `json:"content,omitempty"`
	Categories map[string]int      `json:"categories,omitempty"`
	Facts      []memoryFactView    `json:"facts,omitempty"`
	Warnings   []memoryWarningView `json:"warnings,omitempty"`
}

type memoryAppendRequest struct {
	Content string `json:"content"`
}

func (s *Server) handleGetMemory(w http.ResponseWriter, r *http.Request) {
	view, err := s.buildMemoryView()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleAppendMemory(w http.ResponseWriter, r *http.Request) {
	var req memoryAppendRequest
	if !decodeBody(w, r, &req) {
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	memoryDir := s.memoryDir()
	if memoryDir == "" {
		writeError(w, http.StatusInternalServerError, "memory directory not configured")
		return
	}
	if err := ctxmem.BoundedAppend(memoryDir, content); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	view, err := s.buildMemoryView()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, `/\`) || !strings.HasSuffix(name, ".md") {
		writeError(w, http.StatusBadRequest, "invalid memory entry name")
		return
	}
	memoryDir := s.memoryDir()
	if memoryDir == "" {
		writeError(w, http.StatusInternalServerError, "memory directory not configured")
		return
	}
	target := filepath.Join(memoryDir, name)
	if filepath.Clean(target) != filepath.Join(filepath.Clean(memoryDir), name) {
		writeError(w, http.StatusBadRequest, "invalid memory entry name")
		return
	}
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "memory entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	view, err := s.buildMemoryView()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) buildMemoryView() (memoryView, error) {
	memoryDir := s.memoryDir()
	view := memoryView{MemoryDir: memoryDir, Entries: []memoryEntryView{}}
	if memoryDir == "" {
		return view, nil
	}
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return view, nil
		}
		return view, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		name := entry.Name()
		if name == "MEMORY.md.lock" {
			continue
		}
		view.Entries = append(view.Entries, memoryEntryView{
			Name:     name,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Primary:  name == "MEMORY.md",
		})
	}
	sort.Slice(view.Entries, func(i, j int) bool {
		if view.Entries[i].Primary != view.Entries[j].Primary {
			return view.Entries[i].Primary
		}
		return view.Entries[i].Name < view.Entries[j].Name
	})
	if data, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md")); err == nil {
		view.Content = string(data)
		view.Facts, view.Categories, view.Warnings = analyzeMemoryTaxonomy("MEMORY.md", view.Content)
	}
	return view, nil
}

func (s *Server) memoryDir() string {
	if s.deps == nil || s.deps.StarclawDir == "" {
		return ""
	}
	return filepath.Join(s.deps.StarclawDir, "memory")
}

var memoryCategoryAliases = map[string]string{
	"preference":   "preferences",
	"preferences":  "preferences",
	"decision":     "decisions",
	"decisions":    "decisions",
	"command":      "commands",
	"commands":     "commands",
	"architecture": "architecture",
	"arch":         "architecture",
	"person":       "people",
	"people":       "people",
	"risk":         "risks",
	"risks":        "risks",
}

var memoryHeadingRE = regexp.MustCompile(`^#+\s+(.+?)\s*$`)
var memoryBracketRE = regexp.MustCompile(`^\s*[-*]\s*\[([A-Za-z _-]+)\]\s*(.+)$`)
var memoryColonRE = regexp.MustCompile(`^\s*[-*]\s*([A-Za-z _-]+):\s*(.+)$`)
var memoryBulletRE = regexp.MustCompile(`^\s*[-*]\s+(.+)$`)

func analyzeMemoryTaxonomy(entryName, content string) ([]memoryFactView, map[string]int, []memoryWarningView) {
	facts := parseMemoryFacts(entryName, content)
	categories := map[string]int{}
	for _, fact := range facts {
		categories[fact.Category]++
	}
	return facts, categories, memoryWarnings(facts)
}

func parseMemoryFacts(entryName, content string) []memoryFactView {
	var facts []memoryFactView
	currentCategory := "uncategorized"
	lines := strings.Split(content, "\n")
	for idx, raw := range lines {
		lineNo := idx + 1
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if match := memoryHeadingRE.FindStringSubmatch(line); match != nil {
			if category := normalizeMemoryCategory(match[1]); category != "" {
				currentCategory = category
			}
			continue
		}
		category := currentCategory
		text := ""
		if match := memoryBracketRE.FindStringSubmatch(line); match != nil {
			if normalized := normalizeMemoryCategory(match[1]); normalized != "" {
				category = normalized
			}
			text = strings.TrimSpace(match[2])
		} else if match := memoryColonRE.FindStringSubmatch(line); match != nil {
			if normalized := normalizeMemoryCategory(match[1]); normalized != "" {
				category = normalized
				text = strings.TrimSpace(match[2])
			} else {
				text = strings.TrimSpace(match[1] + ": " + match[2])
			}
		} else if match := memoryBulletRE.FindStringSubmatch(line); match != nil {
			text = strings.TrimSpace(match[1])
		}
		if text == "" {
			continue
		}
		facts = append(facts, memoryFactView{
			Category: category,
			Text:     text,
			Entry:    entryName,
			Line:     lineNo,
			Subject:  memorySubject(text),
		})
	}
	return facts
}

func normalizeMemoryCategory(value string) string {
	key := strings.ToLower(strings.TrimSpace(value))
	key = strings.TrimSuffix(key, ":")
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	if category, ok := memoryCategoryAliases[key]; ok {
		return category
	}
	if strings.HasSuffix(key, "s") {
		if category, ok := memoryCategoryAliases[strings.TrimSuffix(key, "s")]; ok {
			return category
		}
	}
	return ""
}

func memoryWarnings(facts []memoryFactView) []memoryWarningView {
	var warnings []memoryWarningView
	byText := map[string][]memoryFactView{}
	bySubject := map[string][]memoryFactView{}
	for _, fact := range facts {
		normalized := normalizeMemoryFactText(fact.Text)
		if normalized != "" {
			byText[normalized] = append(byText[normalized], fact)
		}
		if fact.Subject != "" {
			bySubject[fact.Subject] = append(bySubject[fact.Subject], fact)
		}
	}
	for _, group := range byText {
		if len(group) < 2 {
			continue
		}
		warnings = append(warnings, memoryWarningView{
			Type:     "duplicate",
			Message:  fmt.Sprintf("Duplicate memory fact appears %d times.", len(group)),
			Lines:    memoryFactLines(group),
			Entries:  memoryFactEntries(group),
			Category: group[0].Category,
		})
	}
	for subject, group := range bySubject {
		if len(group) < 2 || len(uniqueMemoryTexts(group)) < 2 {
			continue
		}
		warnings = append(warnings, memoryWarningView{
			Type:    "conflict",
			Subject: subject,
			Message: fmt.Sprintf("Multiple memory facts mention %q with different wording.", subject),
			Lines:   memoryFactLines(group),
			Entries: memoryFactEntries(group),
		})
	}
	sort.Slice(warnings, func(i, j int) bool {
		if warnings[i].Type != warnings[j].Type {
			return warnings[i].Type < warnings[j].Type
		}
		return warnings[i].Message < warnings[j].Message
	})
	return warnings
}

func normalizeMemoryFactText(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(text))), " ")
}

func memorySubject(text string) string {
	lower := normalizeMemoryFactText(text)
	for _, sep := range []string{":", " is ", " should ", " uses ", " use ", " prefers ", " prefer "} {
		if idx := strings.Index(lower, sep); idx > 0 {
			subject := strings.TrimSpace(lower[:idx])
			if len(subject) >= 3 {
				return subject
			}
		}
	}
	words := strings.Fields(lower)
	if len(words) >= 3 {
		return strings.Join(words[:3], " ")
	}
	return lower
}

func memoryFactLines(facts []memoryFactView) []int {
	lines := make([]int, 0, len(facts))
	for _, fact := range facts {
		lines = append(lines, fact.Line)
	}
	sort.Ints(lines)
	return lines
}

func memoryFactEntries(facts []memoryFactView) []string {
	seen := map[string]bool{}
	var entries []string
	for _, fact := range facts {
		if fact.Entry == "" || seen[fact.Entry] {
			continue
		}
		seen[fact.Entry] = true
		entries = append(entries, fact.Entry)
	}
	sort.Strings(entries)
	return entries
}

func uniqueMemoryTexts(facts []memoryFactView) map[string]bool {
	out := map[string]bool{}
	for _, fact := range facts {
		out[normalizeMemoryFactText(fact.Text)] = true
	}
	return out
}
