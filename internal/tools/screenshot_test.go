package tools

import (
	"context"
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
