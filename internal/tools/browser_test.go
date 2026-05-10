package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestBrowserTool_Info(t *testing.T) {
	tool := &BrowserTool{}
	info := tool.Info()
	if info.Name != "browser" {
		t.Errorf("Name = %q, want 'browser'", info.Name)
	}
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
	if info.Parameters == nil {
		t.Error("Parameters should not be nil")
	}
	if len(info.Required) != 1 || info.Required[0] != "action" {
		t.Error("Expected required parameter 'action'")
	}
	props, ok := info.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("Parameters should have properties")
	}
	for _, key := range []string{"action", "url"} {
		if _, ok := props[key]; !ok {
			t.Errorf("Expected '%s' parameter", key)
		}
	}
}

func TestBrowserTool_RequiresApproval(t *testing.T) {
	tool := &BrowserTool{}
	if tool.RequiresApproval() {
		t.Error("browser RequiresApproval should return false (checked per-call)")
	}
}

func TestBrowserTool_InvalidArgs(t *testing.T) {
	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{invalid json}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid JSON args")
	}
}

func TestBrowserTool_UnknownAction(t *testing.T) {
	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{"action":"unknown"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for unknown action")
	}
	if !strings.Contains(result.Content, "unknown action") {
		t.Errorf("Expected unknown action error, got: %s", result.Content)
	}
}

func TestBrowserTool_MissingAction(t *testing.T) {
	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing action")
	}
	if !strings.Contains(result.Content, "missing required parameter") {
		t.Errorf("Expected missing parameter error, got: %s", result.Content)
	}
}

func TestBrowserTool_NavigateNoURL(t *testing.T) {
	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{"action":"navigate"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for navigate without url")
	}
	if !strings.Contains(result.Content, "requires 'url' parameter") {
		t.Errorf("Expected URL required error, got: %s", result.Content)
	}
}

func TestBrowserTool_Navigate(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping on macOS — opens real browser window")
	}
	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{"action":"navigate","url":"https://example.com"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("Unexpected error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Opened URL") && !strings.Contains(result.Content, "example.com") {
		t.Errorf("Expected navigation success, got: %s", result.Content)
	}
}

func TestBrowserTool_IsReadOnlyCall(t *testing.T) {
	tool := &BrowserTool{}
	if !tool.IsReadOnlyCall(`{"action":"get_title"}`) {
		t.Error("get_title should be read-only")
	}
	if tool.IsReadOnlyCall(`{"action":"navigate","url":"https://example.com"}`) {
		t.Error("navigate should not be read-only")
	}
	if tool.IsReadOnlyCall(`invalid`) {
		t.Error("invalid JSON should not be read-only")
	}
}

func TestBrowserTool_GetTitleNonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping on macOS")
	}

	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{"action":"get_title"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for get_title on non-macOS")
	}
	if !strings.Contains(result.Content, "only available on macOS") {
		t.Errorf("Expected macOS-only error, got: %s", result.Content)
	}
}
