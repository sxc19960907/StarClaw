package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilePreviewTool_Info(t *testing.T) {
	tool := &FilePreviewTool{}
	info := tool.Info()
	if info.Name != "file_preview" {
		t.Errorf("Name = %q, want 'file_preview'", info.Name)
	}
	if !tool.IsReadOnlyCall("") {
		t.Error("file_preview should be read-only")
	}
}

func TestFilePreviewTool_BasicPreview(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	path := filepath.Join(dir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tool := &FilePreviewTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "line1") {
		t.Errorf("expected content to contain 'line1', got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "line5") {
		t.Errorf("expected content to contain 'line5', got: %s", result.Content)
	}
}

func TestFilePreviewTool_LinesLimit(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	path := filepath.Join(dir, "many.txt")
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tool := &FilePreviewTool{}
	// Only show first 2 lines
	result, err := tool.Run(context.Background(), `{"path":"`+path+`","lines":2}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "line1") {
		t.Errorf("expected content to contain 'line1', got: %s", result.Content)
	}
	if strings.Contains(result.Content, "line3") {
		t.Errorf("expected content NOT to contain 'line3', got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "truncated") {
		t.Error("expected truncated message")
	}
}

func TestFilePreviewTool_MissingFile(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	tool := &FilePreviewTool{}
	result, err := tool.Run(context.Background(), `{"path":"missing.txt"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing file")
	}
}

func TestFilePreviewTool_EmptyPath(t *testing.T) {
	tool := &FilePreviewTool{}
	result, err := tool.Run(context.Background(), `{"path":""}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for empty path")
	}
}

func TestFilePreviewTool_RequiresApproval(t *testing.T) {
	tool := &FilePreviewTool{}
	if !tool.RequiresApproval() {
		t.Error("file_preview should require approval")
	}
}

func TestFilePreviewTool_InvalidJSON(t *testing.T) {
	tool := &FilePreviewTool{}
	result, err := tool.Run(context.Background(), `{invalid}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid JSON")
	}
}
