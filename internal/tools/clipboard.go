package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/starclaw/starclaw/internal/agent"
)

// ClipboardTool reads and writes the system clipboard.
type ClipboardTool struct{}

type clipboardArgs struct {
	Action string `json:"action"`
	Text   string `json:"text,omitempty"`
}

func (t *ClipboardTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "clipboard",
		Description: "Read or write the system clipboard. Use action 'read' to get clipboard contents, 'write' to set them.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{"type": "string", "description": "Action: 'read' or 'write'"},
				"text":   map[string]any{"type": "string", "description": "Text to write to clipboard (required for write action)"},
			},
		},
		Required: []string{"action"},
	}
}

func (t *ClipboardTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args clipboardArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	switch args.Action {
	case "read":
		content, err := readClipboard(ctx)
		if err != nil {
			return agent.ToolResult{Content: fmt.Sprintf("clipboard read error: %v", err), IsError: true}, nil
		}
		return agent.ToolResult{Content: content}, nil

	case "write":
		if args.Text == "" {
			return agent.ToolResult{Content: "text is required for write action", IsError: true}, nil
		}
		if err := writeClipboard(ctx, args.Text); err != nil {
			return agent.ToolResult{Content: fmt.Sprintf("clipboard write error: %v", err), IsError: true}, nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("wrote %d bytes to clipboard", len(args.Text))}, nil

	default:
		return agent.ToolResult{Content: fmt.Sprintf("unknown action: %q (use 'read' or 'write')", args.Action), IsError: true}, nil
	}
}

func (t *ClipboardTool) RequiresApproval() bool { return true }

func (t *ClipboardTool) IsReadOnlyCall(argsJSON string) bool {
	var args struct {
		Action string `json:"action"`
	}
	if json.Unmarshal([]byte(argsJSON), &args) != nil {
		return false
	}
	return args.Action == "read"
}

func readClipboard(ctx context.Context) (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "pbpaste")
	case "linux":
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-o")
	case "windows":
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "Get-Clipboard")
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", cmd.Path, err)
	}
	return string(output), nil
}

func writeClipboard(ctx context.Context, text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "pbcopy")
	case "linux":
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "$input | Set-Clipboard")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	cmd.Stdin = bytes.NewReader([]byte(text))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmd.Path, err)
	}
	return nil
}
