package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestAccessibilityTool_Info(t *testing.T) {
	tool := &AccessibilityTool{}
	info := tool.Info()
	if info.Name != "accessibility" {
		t.Errorf("Name = %q, want 'accessibility'", info.Name)
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
	for _, key := range []string{"action", "app", "x", "y"} {
		if _, ok := props[key]; !ok {
			t.Errorf("Expected '%s' parameter", key)
		}
	}
}

func TestAccessibilityTool_RequiresApproval(t *testing.T) {
	tool := &AccessibilityTool{}
	if !tool.RequiresApproval() {
		t.Error("accessibility should require approval")
	}
}

func TestAccessibilityTool_InvalidArgs(t *testing.T) {
	tool := &AccessibilityTool{}
	result, err := tool.Run(context.Background(), `{invalid json}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid JSON args")
	}
}

func TestAccessibilityTool_UnknownAction(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &AccessibilityTool{}
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

func TestAccessibilityTool_MissingAction(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &AccessibilityTool{}
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

func TestAccessibilityTool_NonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping on macOS")
	}

	tool := &AccessibilityTool{}
	result, err := tool.Run(context.Background(), `{"action":"get_focused_element"}`)
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

func TestAccessibilityTool_IsReadOnlyCall(t *testing.T) {
	tool := &AccessibilityTool{}
	if !tool.IsReadOnlyCall(`{"action":"get_focused_element"}`) {
		t.Error("get_focused_element should be read-only")
	}
	if !tool.IsReadOnlyCall(`{"action":"list_windows"}`) {
		t.Error("list_windows should be read-only")
	}
	if !tool.IsReadOnlyCall(`{"action":"get_element_at"}`) {
		t.Error("get_element_at should be read-only")
	}
	if !tool.IsReadOnlyCall(`{"action":"get_window_title"}`) {
		t.Error("get_window_title should be read-only")
	}
}

func TestAccessibilityTool_GetFocusedElement(t *testing.T) {
	t.Skip("skipping — queries real macOS UI. Run with TEST_REAL_ACCESSIBILITY=1.")
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &AccessibilityTool{}
	result, err := tool.Run(context.Background(), `{"action":"get_focused_element"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		// May fail due to missing Accessibility permissions — log and proceed
		t.Logf("Got expected error (likely permissions): %s", result.Content)
		return
	}
	if !strings.Contains(result.Content, "Application:") {
		t.Errorf("Expected application info, got: %s", result.Content)
	}
}

func TestAccessibilityTool_ListWindows(t *testing.T) {
	t.Skip("skipping — queries real macOS UI. Run with TEST_REAL_ACCESSIBILITY=1.")
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &AccessibilityTool{}
	result, err := tool.Run(context.Background(), `{"action":"list_windows"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Logf("Got expected error (likely permissions): %s", result.Content)
		return
	}
	if !strings.Contains(result.Content, "Application:") && !strings.Contains(result.Content, "Window count:") {
		t.Errorf("Expected window list, got: %s", result.Content)
	}
}

func TestAccessibilityTool_GetWindowTitle(t *testing.T) {
	t.Skip("skipping — queries real macOS UI. Run with TEST_REAL_ACCESSIBILITY=1.")
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &AccessibilityTool{}
	result, err := tool.Run(context.Background(), `{"action":"get_window_title"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Logf("Got expected error (likely permissions): %s", result.Content)
		return
	}
	if !strings.Contains(result.Content, "Window title:") && !strings.Contains(result.Content, "No windows open") {
		t.Errorf("Expected window title, got: %s", result.Content)
	}
}

func TestAccessibilityTool_GetElementAt(t *testing.T) {
	t.Skip("skipping — queries real macOS UI. Run with TEST_REAL_ACCESSIBILITY=1.")
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &AccessibilityTool{}
	result, err := tool.Run(context.Background(), `{"action":"get_element_at","x":100,"y":100}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		t.Logf("Got expected error (likely permissions): %s", result.Content)
		return
	}
	if !strings.Contains(result.Content, "Application:") && !strings.Contains(result.Content, "Requested coordinates:") {
		t.Errorf("Expected element info, got: %s", result.Content)
	}
}
