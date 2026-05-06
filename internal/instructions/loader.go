// Package instructions loads hierarchical instruction files for the agent.
//
// Priority order (highest wins in deduplication):
//
//	Global:  ~/.starclaw/instructions.md → rules/*.md
//	Project: .starclaw/instructions.md → rules/*.md → instructions.local.md
package instructions

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// LoadInstructions reads all instruction files and returns combined content.
// starclawDir: global config dir (~/.starclaw).
// projectDir: project-level dir (.starclaw relative to CWD), can be empty.
// maxTokens: approximate token budget (1 token ~ 4 chars).
func LoadInstructions(starclawDir, projectDir string, maxTokens int) (string, error) {
	type source struct {
		path     string
		priority int
	}

	var sources []source
	priority := 0

	if starclawDir != "" {
		sources = append(sources, source{filepath.Join(starclawDir, "instructions.md"), priority})
		priority++
		for _, rf := range sortedMDFiles(filepath.Join(starclawDir, "rules")) {
			sources = append(sources, source{rf, priority})
			priority++
		}
	}

	if projectDir != "" {
		sources = append(sources, source{filepath.Join(projectDir, "instructions.md"), priority})
		priority++
		for _, rf := range sortedMDFiles(filepath.Join(projectDir, "rules")) {
			sources = append(sources, source{rf, priority})
			priority++
		}
		sources = append(sources, source{filepath.Join(projectDir, "instructions.local.md"), priority})
		priority++
	}

	type fileContent struct {
		path     string
		lines    []string
		priority int
	}

	var loaded []fileContent
	for _, src := range sources {
		data, err := readMDFile(src.path)
		if err != nil {
			continue
		}
		loaded = append(loaded, fileContent{
			path:     src.path,
			lines:    strings.Split(data, "\n"),
			priority: src.priority,
		})
	}

	// Deduplicate: process highest priority first to claim lines.
	seenLines := make(map[string]struct{})
	for i := len(loaded) - 1; i >= 0; i-- {
		fc := &loaded[i]
		deduped := make([]string, 0, len(fc.lines))
		for _, line := range fc.lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				deduped = append(deduped, line)
				continue
			}
			if _, exists := seenLines[trimmed]; !exists {
				seenLines[trimmed] = struct{}{}
				deduped = append(deduped, line)
			}
		}
		fc.lines = deduped
	}

	// Build output in load order (lowest priority first).
	maxChars := maxTokens * 4
	var parts []string
	for _, fc := range loaded {
		content := strings.Join(fc.lines, "\n")
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		part := fmt.Sprintf("<!-- from: %s -->\n%s", fc.path, content)
		parts = append(parts, part)
	}

	result := strings.Join(parts, "\n\n")
	if len(result) > maxChars {
		result = result[:maxChars]
		result += "\n[Instructions truncated]"
	}

	return result, nil
}

// LoadMemory reads MEMORY.md from starclawDir/memory/.
// Returns the first maxLines lines. Empty string if the file doesn't exist.
func LoadMemory(starclawDir string, maxLines int) (string, error) {
	if starclawDir == "" {
		return "", nil
	}
	return LoadMemoryFrom(filepath.Join(starclawDir, "memory"), maxLines)
}

// LoadMemoryFrom reads MEMORY.md from the given directory.
// Returns the first maxLines lines. Expands inline markdown links to local .md files.
func LoadMemoryFrom(dir string, maxLines int) (string, error) {
	if dir == "" {
		return "", nil
	}
	path := filepath.Join(dir, "MEMORY.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	result := strings.Join(lines, "\n")
	return annotateStaleness(result, time.Now()), nil
}

// sortedMDFiles returns .md files in dir, sorted alphabetically.
func sortedMDFiles(dir string) []string {
	matches, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil || len(matches) == 0 {
		return nil
	}
	sort.Strings(matches)
	return matches
}

// readMDFile reads a .md file, returning an error for non-.md or invalid UTF-8.
func readMDFile(path string) (string, error) {
	if filepath.Ext(path) != ".md" {
		return "", fmt.Errorf("not a .md file: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("invalid UTF-8: %s", path)
	}
	return string(data), nil
}

// annotateStaleness adds "[N days ago]" annotations to dated headings in memory.
func annotateStaleness(content string, now time.Time) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Match: ## Heading (YYYY-MM-DD) or ## Heading (YYYY-MM-DD HH:MM)
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		open := strings.Index(trimmed, "(")
		close := strings.Index(trimmed, ")")
		if open < 0 || close < 0 || close <= open+10 {
			continue
		}
		dateStr := trimmed[open+1 : close]
		if len(dateStr) >= 10 {
			if t, err := time.Parse("2006-01-02", dateStr[:10]); err == nil {
				days := int(now.Sub(t).Hours() / 24)
				switch {
				case days == 0:
					lines[i] = line + " [today]"
				case days == 1:
					lines[i] = line + " [yesterday]"
				case days > 0:
					lines[i] = line + fmt.Sprintf(" [%d days ago]", days)
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}
