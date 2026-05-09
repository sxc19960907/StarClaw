package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRulesManager(t *testing.T) {
	m := NewRulesManager("/tmp/starclaw")
	if m == nil {
		t.Fatal("NewRulesManager returned nil")
	}
	if m.starclawDir != "/tmp/starclaw" {
		t.Errorf("expected starclawDir %q, got %q", "/tmp/starclaw", m.starclawDir)
	}
}

func TestLoadRules_NoRules(t *testing.T) {
	dir := t.TempDir()
	m := NewRulesManager(dir)
	rules := m.LoadRules("nonexistent-agent")
	if rules != nil {
		t.Errorf("expected nil for nonexistent agent, got %v", rules)
	}
}

func TestSaveAndLoadRules(t *testing.T) {
	dir := t.TempDir()
	m := NewRulesManager(dir)

	err := m.SaveRule("test-agent", "Always be polite.")
	if err != nil {
		t.Fatalf("SaveRule failed: %v", err)
	}

	rules := m.LoadRules("test-agent")
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0] != "Always be polite." {
		t.Errorf("unexpected rule content: %q", rules[0])
	}
}

func TestSaveRule_MultipleRules(t *testing.T) {
	dir := t.TempDir()
	m := NewRulesManager(dir)

	contents := []string{
		"Rule one: be nice.",
		"Rule two: be helpful.",
		"Rule three: be safe.",
	}

	for _, c := range contents {
		if err := m.SaveRule("multi-agent", c); err != nil {
			t.Fatalf("SaveRule failed: %v", err)
		}
	}

	rules := m.LoadRules("multi-agent")
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	for i, c := range contents {
		if rules[i] != c {
			t.Errorf("rule %d: expected %q, got %q", i, c, rules[i])
		}
	}
}

func TestSaveRule_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	m := NewRulesManager(dir)

	err := m.SaveRule("empty-agent", "")
	if err != nil {
		t.Fatalf("SaveRule with empty content should not error: %v", err)
	}

	rules := m.LoadRules("empty-agent")
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty content, got %d", len(rules))
	}
}

func TestLoadRules_AlphabeticalOrder(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "rules", "ordered-agent")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write files out of alphabetical order.
	if err := os.WriteFile(filepath.Join(rulesDir, "zzz.md"), []byte("last"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "aaa.md"), []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "mmm.md"), []byte("middle"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewRulesManager(dir)
	rules := m.LoadRules("ordered-agent")
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	if rules[0] != "first" {
		t.Errorf("expected first rule 'first', got %q", rules[0])
	}
	if rules[1] != "middle" {
		t.Errorf("expected second rule 'middle', got %q", rules[1])
	}
	if rules[2] != "last" {
		t.Errorf("expected third rule 'last', got %q", rules[2])
	}
}

func TestSaveRule_AutoIncrementFilenames(t *testing.T) {
	dir := t.TempDir()
	m := NewRulesManager(dir)

	// Save rules — they should get auto-incremented filenames.
	if err := m.SaveRule("increment-agent", "first rule"); err != nil {
		t.Fatal(err)
	}
	if err := m.SaveRule("increment-agent", "second rule"); err != nil {
		t.Fatal(err)
	}
	if err := m.SaveRule("increment-agent", "third rule"); err != nil {
		t.Fatal(err)
	}

	rulesDir := filepath.Join(dir, "rules", "increment-agent")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 files, got %d", len(entries))
	}

	// Verify filenames are rule-1.md, rule-2.md, rule-3.md.
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	if !names["rule-1.md"] {
		t.Error("expected rule-1.md")
	}
	if !names["rule-2.md"] {
		t.Error("expected rule-2.md")
	}
	if !names["rule-3.md"] {
		t.Error("expected rule-3.md")
	}
}

func TestLoadRules_IgnoresDirectories(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "rules", "no-dir-agent")
	if err := os.MkdirAll(filepath.Join(rulesDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rulesDir, "rule.md"), []byte("actual rule"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewRulesManager(dir)
	rules := m.LoadRules("no-dir-agent")
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule (subdir ignored), got %d", len(rules))
	}
	if rules[0] != "actual rule" {
		t.Errorf("unexpected rule: %q", rules[0])
	}
}

func TestSaveRule_DifferentAgents(t *testing.T) {
	dir := t.TempDir()
	m := NewRulesManager(dir)

	if err := m.SaveRule("agent-a", "rule for a"); err != nil {
		t.Fatal(err)
	}
	if err := m.SaveRule("agent-b", "rule for b"); err != nil {
		t.Fatal(err)
	}

	rulesA := m.LoadRules("agent-a")
	rulesB := m.LoadRules("agent-b")

	if len(rulesA) != 1 || rulesA[0] != "rule for a" {
		t.Errorf("unexpected rules for agent-a: %v", rulesA)
	}
	if len(rulesB) != 1 || rulesB[0] != "rule for b" {
		t.Errorf("unexpected rules for agent-b: %v", rulesB)
	}
}
