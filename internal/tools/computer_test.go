package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestComputerTool_Info(t *testing.T) {
	tool := &ComputerTool{}
	info := tool.Info()
	if info.Name != "computer" {
		t.Errorf("Name = %q, want 'computer'", info.Name)
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
	for _, key := range []string{"action", "x", "y", "key", "text"} {
		if _, ok := props[key]; !ok {
			t.Errorf("Expected '%s' parameter", key)
		}
	}
}

func TestComputerTool_RequiresApproval(t *testing.T) {
	tool := &ComputerTool{}
	if !tool.RequiresApproval() {
		t.Error("computer should require approval")
	}
}

func TestComputerTool_InvalidArgs(t *testing.T) {
	tool := &ComputerTool{}
	result, err := tool.Run(context.Background(), `{invalid json}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for invalid JSON args")
	}
}

func TestComputerTool_UnknownAction(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &ComputerTool{}
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

func TestComputerTool_MissingAction(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &ComputerTool{}
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

func TestComputerTool_NonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping on macOS")
	}

	tool := &ComputerTool{}
	result, err := tool.Run(context.Background(), `{"action":"screenshot"}`)
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

func TestComputerTool_IsReadOnlyCall(t *testing.T) {
	tool := &ComputerTool{}
	if tool.IsReadOnlyCall(`{"action":"mouse_move"}`) {
		t.Error("mouse_move should not be read-only")
	}
	if tool.IsReadOnlyCall(`{"action":"mouse_click"}`) {
		t.Error("mouse_click should not be read-only")
	}
	if tool.IsReadOnlyCall(`{"action":"key_press"}`) {
		t.Error("key_press should not be read-only")
	}
	if tool.IsReadOnlyCall(`{"action":"type_text"}`) {
		t.Error("type_text should not be read-only")
	}
	if tool.IsReadOnlyCall(`{"action":"screenshot"}`) {
		t.Error("screenshot should not be read-only")
	}
}

func TestComputerTool_KeyPressMissingKey(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &ComputerTool{}
	result, err := tool.Run(context.Background(), `{"action":"key_press"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing key")
	}
	if !strings.Contains(result.Content, "requires 'key' parameter") {
		t.Errorf("Expected key parameter error, got: %s", result.Content)
	}
}

func TestComputerTool_TypeTextMissingText(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &ComputerTool{}
	result, err := tool.Run(context.Background(), `{"action":"type_text"}`)
	if err != nil {
		t.Fatalf("Run should not return error, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error for missing text")
	}
	if !strings.Contains(result.Content, "requires 'text' parameter") {
		t.Errorf("Expected text parameter error, got: %s", result.Content)
	}
}

func TestComputerTool_Screenshot(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-macOS")
	}

	tool := &ComputerTool{}
	result, err := tool.Run(context.Background(), `{"action":"screenshot"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.IsError {
		// screencapture may fail due to missing screen recording permissions
		t.Logf("Got expected error (likely permissions): %s", result.Content)
		return
	}
	if !strings.Contains(result.Content, "Screenshot") {
		t.Errorf("Expected screenshot output, got: %s", result.Content)
	}
}

func TestBuildKeyPressScript(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantTell bool
	}{
		{"simple return", "return", true},
		{"simple character", "a", true},
		{"modifier combo", "command+c", true},
		{"double modifier", "cmd+shift+z", true},
		{"modifier key code", "command+up", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			script := buildKeyPressScript(tc.key)
			if !strings.HasPrefix(script, `tell application "System Events"`) {
				t.Errorf("Expected osascript format, got: %s", script)
			}
		})
	}
}

func TestCliclickAvailable(t *testing.T) {
	// This is a best-effort test — cliclick may or may not be installed.
	// The function should not panic or error.
	available := cliclickAvailable()
	if available {
		t.Log("cliclick is installed")
	} else {
		t.Log("cliclick is not installed (expected in CI, ok)")
	}
}
