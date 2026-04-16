package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

func TestFileReadTool(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := &FileReadTool{}

	// Test Info
	info := tool.Info()
	if info.Name != "file_read" {
		t.Errorf("Expected name 'file_read', got '%s'", info.Name)
	}

	// Test successful read
	result, err := tool.Run(context.Background(), `{"path": "`+testFile+`"}`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Line 1") {
		t.Errorf("Expected content to contain 'Line 1', got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "5 | Line 5") {
		t.Errorf("Expected line 5, got: %s", result.Content)
	}

	// Test with offset
	result, _ = tool.Run(context.Background(), `{"path": "`+testFile+`", "offset": 2}`)
	if strings.Contains(result.Content, "Line 1") {
		t.Error("With offset=2, should not contain Line 1")
	}
	if !strings.Contains(result.Content, "3 | Line 3") {
		t.Error("With offset=2, should contain Line 3")
	}

	// Test with limit
	result, _ = tool.Run(context.Background(), `{"path": "`+testFile+`", "limit": 2}`)
	lines := strings.Count(result.Content, "\n")
	if lines != 2 {
		t.Errorf("With limit=2, expected 2 lines, got %d", lines)
	}

	// Test non-existent file
	result, _ = tool.Run(context.Background(), `{"path": "/nonexistent/file.txt"}`)
	if !result.IsError {
		t.Error("Expected error for non-existent file")
	}
	if result.ErrorCategory != agent.ErrCategoryValidation {
		t.Errorf("Expected validation error, got %s", result.ErrorCategory)
	}
}

func TestFileReadTool_IsSafeArgs(t *testing.T) {
	tool := &FileReadTool{}

	// Test safe args (under CWD)
	if !tool.IsSafeArgs(`{"path": "./file.txt"}`) {
		t.Error("Expected ./file.txt to be safe")
	}

	// Test unsafe args (absolute path outside CWD)
	if tool.IsSafeArgs(`{"path": "/etc/passwd"}`) {
		t.Error("Expected /etc/passwd to be unsafe")
	}
}
