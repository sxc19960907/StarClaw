// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
)

// ScreenshotTool captures the macOS desktop screen using the screencapture command.
type ScreenshotTool struct{}

type screenshotArgs struct {
	Path string `json:"path,omitempty"`
}

// Info returns the tool definition for the LLM.
func (t *ScreenshotTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "screenshot",
		Description: "Capture the macOS desktop screen using the screencapture command. Use only for native macOS UI. For web page screenshots use the browser tool.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Output file path for the screenshot. If not provided, saves to /tmp/screenshot_<timestamp>.png",
				},
			},
		},
	}
}

// Run executes the screenshot tool.
func (t *ScreenshotTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args screenshotArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	path := args.Path
	if path != "" {
		path = ExpandHome(path)
		if err := IsSafePath(path); err != nil {
			return agent.ValidationError("unsafe path: " + err.Error()), nil
		}
	}

	if runtime.GOOS != "darwin" {
		return agent.ToolResult{
			Content: "screenshot only available on macOS",
			IsError: true,
		}, nil
	}

	if path == "" {
		path = fmt.Sprintf("/tmp/screenshot_%d.png", time.Now().UnixMilli())
	}

	cmd := exec.CommandContext(ctx, "screencapture", "-x", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("screencapture error: %v\n%s", err, string(output)),
			IsError: true,
		}, nil
	}

	return agent.ToolResult{Content: fmt.Sprintf("Screenshot saved to: %s", path)}, nil
}

// RequiresApproval returns true because taking a screenshot captures sensitive information.
func (t *ScreenshotTool) RequiresApproval() bool { return true }

// IsReadOnlyCall returns true because taking a screenshot is a read-only operation.
func (t *ScreenshotTool) IsReadOnlyCall(string) bool { return true }
