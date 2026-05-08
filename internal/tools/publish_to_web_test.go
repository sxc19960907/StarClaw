package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/config"
)

func TestPublishToWebTool_Info(t *testing.T) {
	tool := NewPublishToWebTool()
	info := tool.Info()
	if info.Name != "publish_to_web" {
		t.Errorf("Name = %q, want 'publish_to_web'", info.Name)
	}
	if info.Description == "" {
		t.Error("Info should have a non-empty Description")
	}
	if info.Parameters == nil {
		t.Error("Info should have Parameters")
	}
	if len(info.Required) != 1 || info.Required[0] != "path" {
		t.Errorf("Required = %v, want ['path']", info.Required)
	}
}

func TestPublishToWebTool_RequiresApproval(t *testing.T) {
	tool := NewPublishToWebTool()
	if tool.RequiresApproval() {
		t.Error("publish_to_web should not require approval")
	}
}

func TestPublishToWebTool_Run_InvalidJSON(t *testing.T) {
	tool := NewPublishToWebTool()
	result, err := tool.Run(context.Background(), "not-json")
	if err != nil {
		t.Fatalf("Run should not return error for invalid JSON: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result for invalid JSON")
	}
	if !strings.Contains(result.Content, "validation error") {
		t.Error("Expected validation error message")
	}
}

func TestPublishToWebTool_Run_MissingPath(t *testing.T) {
	tool := NewPublishToWebTool()
	result, err := tool.Run(context.Background(), `{"purpose": "testing"}`)
	if err != nil {
		t.Fatalf("Run should not return error for empty path: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for empty path")
	}
}

func TestPublishToWebTool_Run_FileNotFound(t *testing.T) {
	tool := NewPublishToWebTool()
	result, err := tool.Run(context.Background(), `{"path": "/nonexistent/file.txt", "purpose": "testing"}`)
	if err != nil {
		t.Fatalf("Run should not return error for nonexistent file: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for nonexistent file")
	}
	if !strings.Contains(result.Content, "file not found") {
		t.Errorf("Expected 'file not found' error, got: %s", result.Content)
	}
}

func TestPublishToWebTool_Run_HappyPath(t *testing.T) {
	// Create a temporary file to publish
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "test.html")
	srcContent := "<html><body><h1>Hello World</h1></body></html>"
	if err := os.WriteFile(src, []byte(srcContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewPublishToWebTool()
	result, err := tool.Run(context.Background(), `{"path": "`+src+`", "purpose": "testing publish functionality"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("Run returned error: %s", result.Content)
	}

	// Verify the response contains a URL and local path
	if !strings.Contains(result.Content, "http://localhost:7533/web/") {
		t.Errorf("Expected URL in result, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Published") {
		t.Errorf("Expected 'Published' in result, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Local path:") {
		t.Errorf("Expected 'Local path:' in result, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "testing publish functionality") {
		t.Errorf("Expected purpose in result, got: %s", result.Content)
	}

	// Verify the file was actually copied
	starclawDir := config.StarclawDir()
	webDir := filepath.Join(starclawDir, "web")
	entries, err := os.ReadDir(webDir)
	if err != nil {
		t.Fatalf("failed to read web directory: %v", err)
	}

	var found bool
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pubFiles, err := os.ReadDir(filepath.Join(webDir, entry.Name()))
		if err != nil {
			continue
		}
		for _, f := range pubFiles {
			if f.Name() == "test.html" && !f.IsDir() {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("published file was not found in ~/.starclaw/web/ directory")
	}

	// Cleanup: remove published files
	os.RemoveAll(webDir)
}

func TestPublishToWebTool_Run_DirectoryPath(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewPublishToWebTool()
	result, err := tool.Run(context.Background(), `{"path": "`+tmpDir+`", "purpose": "testing directory rejection"}`)
	if err != nil {
		t.Fatalf("Run should not return error for directory: %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for directory path")
	}
	if !strings.Contains(result.Content, "directory") {
		t.Errorf("Expected 'directory' in error message, got: %s", result.Content)
	}
}

func TestPublishToWebTool_Run_WithPurpose(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "report.md")
	if err := os.WriteFile(src, []byte("# Test Report"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewPublishToWebTool()
	result, err := tool.Run(context.Background(), `{"path": "`+src+`", "purpose": "sharing test report with reviewer"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("Run returned error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "sharing test report with reviewer") {
		t.Errorf("Expected purpose in result, got: %s", result.Content)
	}

	// Cleanup
	starclawDir := config.StarclawDir()
	os.RemoveAll(filepath.Join(starclawDir, "web"))
}

func TestPublishToWebTool_Run_WithoutPurpose(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "data.json")
	if err := os.WriteFile(src, []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewPublishToWebTool()
	result, err := tool.Run(context.Background(), `{"path": "`+src+`"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("Run returned error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Published") {
		t.Errorf("Expected 'Published' in result, got: %s", result.Content)
	}

	// Cleanup
	starclawDir := config.StarclawDir()
	os.RemoveAll(filepath.Join(starclawDir, "web"))
}
