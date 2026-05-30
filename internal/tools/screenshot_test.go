package tools

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScreenshotTool_Info(t *testing.T) {
	tool := &ScreenshotTool{}
	info := tool.Info()
	if info.Name != "screenshot" {
		t.Errorf("Name = %q, want 'screenshot'", info.Name)
	}
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
	if info.Parameters == nil {
		t.Error("Parameters should not be nil")
	}
	props, ok := info.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("Parameters should have properties")
	}
	if _, ok := props["path"]; !ok {
		t.Error("Expected 'path' parameter")
	}
}

func TestScreenshotTool_RequiresApproval(t *testing.T) {
	tool := &ScreenshotTool{}
	if !tool.RequiresApproval() {
		t.Error("screenshot should require approval")
	}
}

func TestScreenshotTool_InvalidArgs(t *testing.T) {
	tool := &ScreenshotTool{}
	result, err := tool.Run(context.Background(), `{invalid json}`)
	if err != nil {
		t.Fatalf("Run should not return error for invalid args, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid JSON args")
	}
}

func TestScreenshotTool_NonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping on macOS")
	}

	tool := &ScreenshotTool{}
	result, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error on non-macOS")
	}
	if !strings.Contains(result.Content, "only available on macOS") {
		t.Errorf("Expected macOS-only error, got: %s", result.Content)
	}
}

func TestScreenshotTool_UnsafePathValidatedBeforePlatform(t *testing.T) {
	parent := t.TempDir()
	project := filepath.Join(parent, "project")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("chdir project: %v", err)
	}

	tool := &ScreenshotTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+filepath.Join(outside, "shot.png")+`"}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected unsafe path error")
	}
	if !strings.Contains(result.Content, "unsafe path") {
		t.Fatalf("expected unsafe path message, got: %s", result.Content)
	}
}
