package tools

import (
	"context"
	"encoding/json"
	"os"
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
	action := props["action"].(map[string]any)
	if !strings.Contains(action["description"].(string), "snapshot") {
		t.Errorf("action description should mention snapshot, got %q", action["description"])
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

func TestBrowserToolMarksLeaseOnRun(t *testing.T) {
	tool := &BrowserTool{}
	ctx := WithBrowserUseLease(context.Background())
	lease := BrowserUseLeaseFrom(ctx)
	result, err := tool.Run(ctx, `{invalid json}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Fatal("invalid JSON should still return a tool error")
	}
	if got := BrowserOwnerActiveCount(tool); got != 1 {
		t.Fatalf("owner count = %d, want 1", got)
	}
	lease.ReleaseOnly()
	if got := BrowserOwnerActiveCount(tool); got != 0 {
		t.Fatalf("owner count after release = %d, want 0", got)
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
	if os.Getenv("CI") != "" {
		t.Skip("skipping in CI because opening a real browser requires a desktop environment")
	}
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
	if !tool.IsReadOnlyCall(`{"action":"status"}`) {
		t.Error("status should be read-only")
	}
	if !tool.IsReadOnlyCall(`{"action":"snapshot"}`) {
		t.Error("snapshot should be read-only")
	}
	if tool.IsReadOnlyCall(`{"action":"navigate","url":"https://example.com"}`) {
		t.Error("navigate should not be read-only")
	}
	if tool.IsReadOnlyCall(`invalid`) {
		t.Error("invalid JSON should not be read-only")
	}
}

func TestBrowserTool_Status(t *testing.T) {
	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{"action":"status"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if result.IsError {
		t.Fatalf("Unexpected error: %s", result.Content)
	}
	var status browserStatus
	if err := json.Unmarshal([]byte(result.Content), &status); err != nil {
		t.Fatalf("status should be JSON: %v\n%s", err, result.Content)
	}
	if status.Platform != runtime.GOOS {
		t.Fatalf("Platform = %q, want %q", status.Platform, runtime.GOOS)
	}
	if len(status.Actions) != len(browserActions) {
		t.Fatalf("Actions = %#v, want %#v", status.Actions, browserActions)
	}
	if len(status.SupportedBrowsers) == 0 {
		t.Fatal("SupportedBrowsers should not be empty")
	}
}

func TestBrowserTool_SnapshotNonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("non-darwin behavior")
	}
	tool := &BrowserTool{}
	result, err := tool.Run(context.Background(), `{"action":"snapshot"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if result.IsError {
		t.Fatalf("snapshot unsupported response should be structured non-error, got: %s", result.Content)
	}
	var snap browserSnapshot
	if err := json.Unmarshal([]byte(result.Content), &snap); err != nil {
		t.Fatalf("snapshot should be JSON: %v\n%s", err, result.Content)
	}
	if snap.Supported {
		t.Fatal("Snapshot should not be supported on non-macOS")
	}
	if snap.Source != "unsupported_platform" {
		t.Fatalf("Source = %q, want unsupported_platform", snap.Source)
	}
}

func TestParseBrowserSnapshotOutput(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want browserSnapshot
	}{
		{
			name: "browser",
			out:  "browser\x1fGoogle Chrome\x1fTitle\x1fhttps://example.com\x1fchrome",
			want: browserSnapshot{Supported: true, Platform: "darwin", Browser: "Google Chrome", Title: "Title", URL: "https://example.com", Source: "chrome"},
		},
		{
			name: "window",
			out:  "window\x1fFinder\x1fDesktop\x1ffrontmost_window",
			want: browserSnapshot{Supported: true, Platform: "darwin", FrontmostApp: "Finder", WindowTitle: "Desktop", Source: "frontmost_window"},
		},
		{
			name: "app",
			out:  "app\x1fFinder\x1ffrontmost_app\x1fNo window title available",
			want: browserSnapshot{Supported: false, Platform: "darwin", FrontmostApp: "Finder", Source: "frontmost_app", Message: "No window title available"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := parseBrowserSnapshotOutput(tt.out, "darwin")
			if got != tt.want {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
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
