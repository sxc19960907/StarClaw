package tools

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestNotifyTool_Info(t *testing.T) {
	tool := &NotifyTool{}
	info := tool.Info()

	if info.Name != "notify" {
		t.Errorf("Name = %q, want 'notify'", info.Name)
	}
	if info.Description == "" {
		t.Error("expected non-empty description")
	}
	if info.Parameters == nil {
		t.Fatal("expected non-nil parameters")
	}

	// title and message should be required
	for _, required := range []string{"title", "message"} {
		found := false
		for _, r := range info.Required {
			if r == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in required list", required)
		}
	}
}

func TestNotifyTool_Run_InvalidJSON(t *testing.T) {
	tool := &NotifyTool{}
	result, err := tool.Run(context.Background(), "{invalid")
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for invalid JSON")
	}
}

func TestNotifyTool_Run_EmptyTitle(t *testing.T) {
	tool := &NotifyTool{}
	result, err := tool.Run(context.Background(), `{"title": "", "message": "hello"}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for empty title")
	}
	if !strings.Contains(result.Content, "title is required") {
		t.Errorf("expected title is required message, got: %s", result.Content)
	}
}

func TestNotifyTool_Run_EmptyMessage(t *testing.T) {
	tool := &NotifyTool{}
	result, err := tool.Run(context.Background(), `{"title": "Test", "message": ""}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for empty message")
	}
	if !strings.Contains(result.Content, "message is required") {
		t.Errorf("expected message is required message, got: %s", result.Content)
	}
}

func TestNotifyTool_RequiresApproval(t *testing.T) {
	tool := &NotifyTool{}
	if tool.RequiresApproval() {
		t.Error("notify tool should not require approval")
	}
}

func TestNotifyTool_Run(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("notification tests only supported on darwin and linux")
	}

	cmdName := notifyCmdName()
	if cmdName == "" {
		t.Skip("no notification command available on this platform")
	}
	if _, err := exec.LookPath(cmdName); err != nil {
		t.Skipf("%s not found in PATH", cmdName)
	}

	tool := &NotifyTool{}
	result, err := tool.Run(context.Background(), `{"title": "Test Title", "message": "Test message body"}`)
	if err != nil {
		t.Fatalf("Run returned err: %v", err)
	}
	if result.IsError {
		t.Fatalf("notification failed: %s", result.Content)
	}
	if !strings.Contains(result.Content, "notification sent") {
		t.Errorf("expected success message, got: %s", result.Content)
	}
}

func notifyCmdName() string {
	switch runtime.GOOS {
	case "darwin":
		return "osascript"
	case "linux":
		return "notify-send"
	}
	return ""
}
