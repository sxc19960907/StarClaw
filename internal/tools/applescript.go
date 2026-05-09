// Package tools provides local tool implementations for the agent.
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

// AppleScriptTool executes AppleScript on macOS using the osascript command.
type AppleScriptTool struct{}

type appleScriptArgs struct {
	Script string `json:"script"`
}

// Info returns the tool definition for the LLM.
func (t *AppleScriptTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "applescript",
		Description: "Execute an AppleScript script via osascript on macOS. Use for opening/activating apps, window management, and macOS-specific operations.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"script": map[string]any{
					"type":        "string",
					"description": "AppleScript code to execute",
				},
			},
		},
		Required: []string{"script"},
	}
}

// Run executes the AppleScript tool.
func (t *AppleScriptTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	if runtime.GOOS != "darwin" {
		return agent.ToolResult{
			Content: "applescript only available on macOS",
			IsError: true,
		}, nil
	}

	var args appleScriptArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	if strings.TrimSpace(args.Script) == "" {
		return agent.ValidationError("script is required"), nil
	}

	// Split multi-line scripts into separate -e arguments for osascript
	cmdArgs := buildOsascriptArgs(args.Script)
	cmd := exec.CommandContext(ctx, "osascript", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := string(output)
	if len(result) > 10240 {
		result = result[:10240] + "\n... (truncated)"
	}

	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("osascript error: %v\n%s", err, result),
			IsError: true,
		}, nil
	}

	if result == "" {
		return agent.ToolResult{Content: "script executed successfully (no output)"}, nil
	}
	return agent.ToolResult{Content: result}, nil
}

// RequiresApproval returns true because AppleScript can manipulate applications and system state.
func (t *AppleScriptTool) RequiresApproval() bool { return true }

// IsReadOnlyCall returns false because AppleScript can modify system state.
func (t *AppleScriptTool) IsReadOnlyCall(string) bool { return false }

// buildOsascriptArgs splits an AppleScript into individual lines for -e args.
func buildOsascriptArgs(script string) []string {
	lines := strings.Split(script, "\n")
	var args []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			args = append(args, "-e", line)
		}
	}
	if len(args) == 0 {
		return []string{"-e", script}
	}
	return args
}
