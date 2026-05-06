package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ctxwin "github.com/starclaw/starclaw/internal/context"
)

func TestMemoryAppendTool_Info(t *testing.T) {
	tool := &MemoryAppendTool{}
	info := tool.Info()
	if info.Name != "memory_append" {
		t.Errorf("Name = %q, want 'memory_append'", info.Name)
	}
	if len(info.Required) != 1 || info.Required[0] != "content" {
		t.Error("Required should be ['content']")
	}
}

func TestMemoryAppendTool_EmptyContent(t *testing.T) {
	tool := &MemoryAppendTool{}
	result, _ := tool.Run(context.Background(), `{"content":""}`)
	if !result.IsError {
		t.Error("Empty content should return error")
	}
}

func TestMemoryAppendTool_NoMemoryDir(t *testing.T) {
	tool := &MemoryAppendTool{}
	result, _ := tool.Run(context.Background(), `{"content":"- test"}`)
	if !result.IsError {
		t.Error("Missing memory dir should return error")
	}
	if !strings.Contains(result.Content, "not available") {
		t.Errorf("Expected 'not available' message, got: %s", result.Content)
	}
}

func TestMemoryAppendTool_Success(t *testing.T) {
	dir := t.TempDir()
	tool := &MemoryAppendTool{}

	ctx := ctxwin.WithMemoryDir(context.Background(), dir)
	result, err := tool.Run(ctx, `{"content":"- Important: user prefers tabs over spaces"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Errorf("Expected success, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "next session") {
		t.Errorf("Expected 'next session' message, got: %s", result.Content)
	}

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		t.Fatalf("Read MEMORY.md failed: %v", err)
	}
	if !strings.Contains(string(data), "tabs over spaces") {
		t.Errorf("MEMORY.md should contain the entry: %s", string(data))
	}
}

func TestMemoryAppendTool_RequiresApproval(t *testing.T) {
	tool := &MemoryAppendTool{}
	if tool.RequiresApproval() {
		t.Error("memory_append should not require approval")
	}
}
