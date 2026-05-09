package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// NotifyTool sends desktop notifications.
type NotifyTool struct{}

type notifyArgs struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func (t *NotifyTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "notify",
		Description: "Send a desktop notification.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":   map[string]any{"type": "string", "description": "Notification title"},
				"message": map[string]any{"type": "string", "description": "Notification body text"},
			},
		},
		Required: []string{"title", "message"},
	}
}

func (t *NotifyTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args notifyArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if args.Title == "" {
		return agent.ToolResult{Content: "title is required", IsError: true}, nil
	}
	if args.Message == "" {
		return agent.ToolResult{Content: "message is required", IsError: true}, nil
	}

	if err := sendNotification(ctx, args.Title, args.Message); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("notification error: %v", err), IsError: true}, nil
	}

	return agent.ToolResult{Content: "notification sent"}, nil
}

func (t *NotifyTool) RequiresApproval() bool { return false }

func sendNotification(ctx context.Context, title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		return sendMacOSNotification(ctx, title, message)
	case "linux":
		return sendLinuxNotification(ctx, title, message)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func sendMacOSNotification(ctx context.Context, title, message string) error {
	title = escapeAppleScript(title)
	message = escapeAppleScript(message)
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", cmd.Path, err, string(out))
	}
	return nil
}

func sendLinuxNotification(ctx context.Context, title, message string) error {
	cmd := exec.CommandContext(ctx, "notify-send", title, message)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", cmd.Path, err, string(out))
	}
	return nil
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
