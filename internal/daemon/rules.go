package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RulesManager manages agent rules stored as files on disk under
// <starclawDir>/rules/<agentName>/.
type RulesManager struct {
	starclawDir string
}

// NewRulesManager creates a RulesManager that reads and writes rules
// under the given starclaw directory.
func NewRulesManager(starclawDir string) *RulesManager {
	return &RulesManager{starclawDir: starclawDir}
}

// LoadRules returns all rules for the given agent name.  Each rule is a
// single file's content read from <starclawDir>/rules/<agentName>/.  Files
// are loaded in alphabetical order.  Returns an empty slice if no rules
// exist or the directory cannot be read.
func (m *RulesManager) LoadRules(agentName string) []string {
	rulesDir := filepath.Join(m.starclawDir, "rules", agentName)

	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var rules []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(rulesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(data))
		if content != "" {
			rules = append(rules, content)
		}
	}

	return rules
}

// SaveRule saves a rule for the given agent name.  The rule is written
// to a new file inside <starclawDir>/rules/<agentName>/.  If content is
// empty the call is a no-op.
func (m *RulesManager) SaveRule(agentName, content string) error {
	if content == "" {
		return nil
	}

	rulesDir := filepath.Join(m.starclawDir, "rules", agentName)
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("create rules directory: %w", err)
	}

	// Find the next available rule filename.
	nextID := 1
	entries, err := os.ReadDir(rulesDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasPrefix(name, "rule-") && strings.HasSuffix(name, ".md") {
				var id int
				if _, err := fmt.Sscanf(name, "rule-%d.md", &id); err == nil {
					if id >= nextID {
						nextID = id + 1
					}
				}
			}
		}
	}

	filename := fmt.Sprintf("rule-%d.md", nextID)
	path := filepath.Join(rulesDir, filename)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write rule file: %w", err)
	}

	return nil
}
