package instructions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadInstructions_Hierarchy(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(globalDir, "instructions.md"), []byte("global instructions"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(globalDir, "rules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "rules", "alpha.md"), []byte("rule alpha"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "rules", "beta.md"), []byte("rule beta"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(projectDir, "instructions.md"), []byte("project instructions"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "rules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "rules", "gamma.md"), []byte("rule gamma"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "instructions.local.md"), []byte("local overrides"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := LoadInstructions(globalDir, projectDir, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "global instructions") {
		t.Error("should contain global instructions")
	}
	if !strings.Contains(result, "rule alpha") {
		t.Error("should contain rule alpha")
	}
	if !strings.Contains(result, "local overrides") {
		t.Error("should contain local overrides")
	}
}

func TestLoadInstructions_Empty(t *testing.T) {
	result, err := LoadInstructions("", "", 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty, got: %q", result)
	}
}

func TestLoadInstructions_MissingFiles(t *testing.T) {
	dir := t.TempDir()
	result, err := LoadInstructions(dir, dir, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty for no files, got: %q", result)
	}
}

func TestLoadInstructions_Dedup(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(globalDir, "instructions.md"), []byte("shared line\nglobal only"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "instructions.md"), []byte("shared line\nproject only"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := LoadInstructions(globalDir, projectDir, 10000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Count(result, "shared line") != 1 {
		t.Errorf("shared line should appear once, got:\n%s", result)
	}
	if !strings.Contains(result, "global only") {
		t.Error("should contain global only")
	}
	if !strings.Contains(result, "project only") {
		t.Error("should contain project only")
	}
}

func TestLoadInstructions_Truncation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "instructions.md"), []byte(strings.Repeat("x", 5000)), 0644); err != nil {
		t.Fatal(err)
	}

	// Budget 500 tokens = 2000 chars
	result, err := LoadInstructions(dir, "", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "[Instructions truncated]") {
		t.Error("should have truncation message")
	}
}

func TestLoadInstructions_NonMDSkipped(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "rules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rules", "valid.md"), []byte("valid"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rules", "skip.txt"), []byte("skip"), 0644); err != nil {
		t.Fatal(err)
	}

	result, _ := LoadInstructions(dir, "", 10000)
	if !strings.Contains(result, "valid") {
		t.Error("should contain .md file")
	}
	if strings.Contains(result, "skip") {
		t.Error("should skip non-.md file")
	}
}

func TestLoadInstructions_SourceComments(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "instructions.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	result, _ := LoadInstructions(dir, "", 10000)
	if !strings.Contains(result, "<!-- from:") {
		t.Error("should include source comment")
	}
	if !strings.Contains(result, "instructions.md") {
		t.Error("should include file path in source comment")
	}
}

func TestLoadInstructions_RulesAlphabetical(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "rules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rules", "charlie.md"), []byte("charlie"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rules", "alice.md"), []byte("alice"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "rules", "bob.md"), []byte("bob"), 0644); err != nil {
		t.Fatal(err)
	}

	result, _ := LoadInstructions(dir, "", 10000)
	ai := strings.Index(result, "alice")
	bi := strings.Index(result, "bob")
	ci := strings.Index(result, "charlie")
	if ai >= bi || bi >= ci {
		t.Errorf("rules should be alphabetical")
	}
}

func TestLoadMemory_Exists(t *testing.T) {
	dir := t.TempDir()
	memDir := filepath.Join(dir, "memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := strings.Repeat("line\n", 10)
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := LoadMemory(dir, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Count(result, "line") != 5 {
		t.Errorf("expected 5 lines, got %d", strings.Count(result, "line"))
	}
}

func TestLoadMemory_Missing(t *testing.T) {
	result, err := LoadMemory(t.TempDir(), 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Error("should return empty for missing file")
	}
}

func TestLoadMemory_EmptyDir(t *testing.T) {
	result, err := LoadMemory("", 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Error("should return empty")
	}
}

func TestAnnotateStaleness(t *testing.T) {
	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"recent", "## Note (2026-03-30)", "[2 days ago]"},
		{"today", "## Note (2026-04-01)", "[today]"},
		{"yesterday", "## Note (2026-03-31)", "[yesterday]"},
		{"no date", "## Just a heading", "## Just a heading"},
		{"non-heading", "text with (2026-03-30)", "text with (2026-03-30)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := annotateStaleness(tt.input, now)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("expected to contain %q, got %q", tt.contains, got)
			}
		})
	}
}
