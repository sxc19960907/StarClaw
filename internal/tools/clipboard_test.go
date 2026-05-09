package tools

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestClipboardTool_Info(t *testing.T) {
	tool := &ClipboardTool{}
	info := tool.Info()

	if info.Name != "clipboard" {
		t.Errorf("Name = %q, want 'clipboard'", info.Name)
	}
	if info.Description == "" {
		t.Error("expected non-empty description")
	}
	if info.Parameters == nil {
		t.Fatal("expected non-nil parameters")
	}

	// action should be required
	found := false
	for _, r := range info.Required {
		if r == "action" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'action' in required list")
	}
}

func TestClipboardTool_Run_InvalidJSON(t *testing.T) {
	tool := &ClipboardTool{}
	result, err := tool.Run(context.Background(), "{invalid")
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for invalid JSON")
	}
}

func TestClipboardTool_Run_UnknownAction(t *testing.T) {
	tool := &ClipboardTool{}
	result, err := tool.Run(context.Background(), `{"action": "unknown"}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for unknown action")
	}
	if !strings.Contains(result.Content, "unknown") {
		t.Errorf("expected error message to mention unknown action, got: %s", result.Content)
	}
}

func TestClipboardTool_Run_WriteNoText(t *testing.T) {
	tool := &ClipboardTool{}
	result, err := tool.Run(context.Background(), `{"action": "write"}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for write without text")
	}
	if !strings.Contains(result.Content, "text is required") {
		t.Errorf("expected error about missing text, got: %s", result.Content)
	}
}

func TestClipboardTool_RequiresApproval(t *testing.T) {
	tool := &ClipboardTool{}
	if !tool.RequiresApproval() {
		t.Error("clipboard tool should require approval")
	}
}

func TestClipboardTool_IsReadOnlyCall(t *testing.T) {
	tool := &ClipboardTool{}
	if !tool.IsReadOnlyCall(`{"action": "read"}`) {
		t.Error("read action should be read-only")
	}
	if tool.IsReadOnlyCall(`{"action": "write"}`) {
		t.Error("write action should not be read-only")
	}
	if tool.IsReadOnlyCall(`{"action": "write", "text": "hello"}`) {
		t.Error("write with text should not be read-only")
	}
	// Invalid JSON should not be considered read-only
	if tool.IsReadOnlyCall("bad") {
		t.Error("invalid JSON should not be read-only")
	}
}

func TestClipboardTool_Run_Read(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("unsupported platform for clipboard test")
	}

	cmdName := clipboardCmdName()
	if cmdName == "" {
		t.Skip("no clipboard command available on this platform")
	}
	if _, err := exec.LookPath(cmdName); err != nil {
		t.Skipf("%s not found in PATH", cmdName)
	}

	tool := &ClipboardTool{}
	result, err := tool.Run(context.Background(), `{"action": "read"}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	// Reading clipboard should succeed (content may be empty)
	t.Logf("clipboard read returned %d bytes", len(result.Content))
}

func TestClipboardTool_Run_Write(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("unsupported platform for clipboard test")
	}

	cmdName := clipboardWriteCmdName()
	if cmdName == "" {
		t.Skip("no clipboard write command available on this platform")
	}
	if _, err := exec.LookPath(cmdName); err != nil {
		t.Skipf("%s not found in PATH", cmdName)
	}

	tool := &ClipboardTool{}
	result, err := tool.Run(context.Background(), `{"action": "write", "text": "clipboard test"}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if result.IsError {
		t.Fatalf("write failed: %s", result.Content)
	}
	if !strings.Contains(result.Content, "wrote") {
		t.Errorf("expected success message, got: %s", result.Content)
	}

	// Verify by reading back
	readResult, err := tool.Run(context.Background(), `{"action": "read"}`)
	if err != nil {
		t.Fatalf("read after write returned err: %v", err)
	}
	if readResult.IsError {
		t.Fatalf("read after write failed: %s", readResult.Content)
	}
	if !strings.Contains(readResult.Content, "clipboard test") {
		t.Errorf("expected to read back written text, got: %s", readResult.Content)
	}
}

func clipboardCmdName() string {
	switch runtime.GOOS {
	case "darwin":
		return "pbpaste"
	case "linux":
		return "xclip"
	case "windows":
		return "powershell"
	}
	return ""
}

func clipboardWriteCmdName() string {
	switch runtime.GOOS {
	case "darwin":
		return "pbcopy"
	case "linux":
		return "xclip"
	case "windows":
		return "powershell"
	}
	return ""
}
