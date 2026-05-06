package context

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBoundedAppend(t *testing.T) {
	dir := t.TempDir()
	content := "- New fact: user prefers Go over Python"

	err := BoundedAppend(dir, content)
	if err != nil {
		t.Fatalf("BoundedAppend failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		t.Fatalf("Read MEMORY.md failed: %v", err)
	}
	if !strings.Contains(string(data), content) {
		t.Errorf("MEMORY.md should contain appended content: %s", string(data))
	}
}

func TestBoundedAppend_Overflow(t *testing.T) {
	dir := t.TempDir()

	// Write 151 lines to trigger overflow
	var sb strings.Builder
	for i := 0; i < 151; i++ {
		sb.WriteString("- line ")
		sb.WriteRune(rune('0' + i%10))
		sb.WriteString("\n")
	}
	content := sb.String()

	err := BoundedAppend(dir, content)
	if err != nil {
		t.Fatalf("BoundedAppend failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		t.Fatalf("Read MEMORY.md failed: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, "See [auto-") {
		t.Errorf("Overflow should add pointer line, got: %s", text)
	}

	// Should have created detail file
	files, _ := filepath.Glob(filepath.Join(dir, "auto-*.md"))
	if len(files) < 1 {
		t.Error("Overflow should create auto-*.md detail file")
	}
}

func TestBoundedAppend_EmptyDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new-subdir")
	err := BoundedAppend(dir, "- test entry")
	if err != nil {
		t.Fatalf("BoundedAppend with new dir should succeed: %v", err)
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		content  string
		expected int
	}{
		{"", 0},
		{"line1", 1},
		{"line1\nline2", 2},
		{"line1\nline2\n", 2},
		{"\n\n\n", 3},
	}
	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			if got := countLines([]byte(tt.content)); got != tt.expected {
				t.Errorf("countLines(%q) = %d, want %d", tt.content, got, tt.expected)
			}
		})
	}
}

func TestSplitMemory(t *testing.T) {
	content := `# User Memory

This is my personal context.

## Auto-persisted (2026-01-01)
- Fact 1
- Fact 2

## Auto-consolidated (2026-02-01)
- Merged fact

# More user notes
User stuff here.`

	user, auto := splitMemory(content)
	if !strings.Contains(user, "User Memory") {
		t.Error("user content should contain User Memory")
	}
	if !strings.Contains(user, "More user notes") {
		t.Error("user content should contain More user notes")
	}
	if strings.Contains(user, "Auto-persisted") {
		t.Error("user content should NOT contain Auto-persisted section")
	}
	if !strings.Contains(auto, "Fact 1") {
		t.Error("auto content should contain Fact 1")
	}
	if !strings.Contains(auto, "Merged fact") {
		t.Error("auto content should contain Merged fact")
	}
}

func TestWithMemoryDir(t *testing.T) {
	ctx := context.Background()
	ctx = WithMemoryDir(ctx, "/tmp/memory")

	dir := MemoryDirFromContext(ctx)
	if dir != "/tmp/memory" {
		t.Errorf("MemoryDirFromContext = %q, want /tmp/memory", dir)
	}

	// Empty context
	emptyCtx := context.Background()
	if MemoryDirFromContext(emptyCtx) != "" {
		t.Error("Empty context should return empty string")
	}
}

func TestConsolidateMemory_NotEnoughFiles(t *testing.T) {
	dir := t.TempDir()
	// Only 2 auto files — below threshold
	for i := 0; i < 2; i++ {
		f := filepath.Join(dir, "auto-2026-01-01-xxxxxx.md")
		os.WriteFile(f, []byte("- fact"), 0644)
	}

	err := ConsolidateMemory(context.Background(), nil, dir)
	if err != nil {
		t.Errorf("Should skip silently when below threshold: %v", err)
	}
}

func TestConsolidateMemory_NilCompleter(t *testing.T) {
	dir := t.TempDir()
	// 12 files but nil completer — should skip
	for i := 0; i < 12; i++ {
		f := filepath.Join(dir, "auto-2026-01-01-xxxxx"+string(rune('a'+i))+".md")
		os.WriteFile(f, []byte("- fact"), 0644)
	}

	err := ConsolidateMemory(context.Background(), nil, dir)
	if err != nil {
		t.Errorf("Should skip with nil completer: %v", err)
	}
}
