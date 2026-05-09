// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// ProcessTool manages processes: list running processes and kill by PID or name.
type ProcessTool struct{}

type processArgs struct {
	Action string `json:"action"`
	PID    int    `json:"pid,omitempty"`
	Name   string `json:"name,omitempty"`
}

// Info returns the tool definition for the LLM.
func (t *ProcessTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "process",
		Description: "Manage processes: list running processes or kill a process by PID or name.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action: 'list' to list all processes, 'kill' to terminate a process",
				},
				"pid": map[string]any{
					"type":        "integer",
					"description": "Process ID to kill (required for kill action if name is not provided)",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Process name to kill by name (alternative to pid for kill action). Uses pkill on Unix systems.",
				},
			},
		},
		Required: []string{"action"},
	}
}

// Run executes the process tool.
func (t *ProcessTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args processArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	switch args.Action {
	case "list":
		return t.listProcesses(ctx)

	case "kill":
		return t.killProcess(ctx, args)

	default:
		return agent.ValidationError(fmt.Sprintf("unknown action: %q (use 'list' or 'kill')", args.Action)), nil
	}
}

func (t *ProcessTool) listProcesses(ctx context.Context) (agent.ToolResult, error) {
	cmd := exec.CommandContext(ctx, "ps", "aux")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("ps error: %v\n%s", err, string(output)),
			IsError: true,
		}, nil
	}

	result := string(output)
	if len(result) > 30000 {
		result = result[:30000] + "\n... (truncated)"
	}
	return agent.ToolResult{Content: result}, nil
}

func (t *ProcessTool) killProcess(ctx context.Context, args processArgs) (agent.ToolResult, error) {
	if args.PID > 0 {
		cmd := exec.CommandContext(ctx, "kill", fmt.Sprintf("%d", args.PID))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return agent.ToolResult{
				Content: fmt.Sprintf("kill error: %v\n%s", err, string(output)),
				IsError: true,
			}, nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("sent SIGTERM to PID %d", args.PID)}, nil
	}

	if strings.TrimSpace(args.Name) != "" {
		cmd := exec.CommandContext(ctx, "pkill", args.Name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return agent.ToolResult{
				Content: fmt.Sprintf("pkill error: %v\n%s", err, string(output)),
				IsError: true,
			}, nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("sent SIGTERM to process '%s'", args.Name)}, nil
	}

	return agent.ValidationError("pid or name is required for kill action"), nil
}

// RequiresApproval returns true for kill actions, false for list.
func (t *ProcessTool) RequiresApproval() bool { return false }

// IsSafeArgs returns true for list action (read-only), false for kill (modifies state).
func (t *ProcessTool) IsSafeArgs(argsJSON string) bool {
	var args processArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return args.Action == "list"
}

// IsReadOnlyCall returns true for list action, false for kill.
func (t *ProcessTool) IsReadOnlyCall(argsJSON string) bool {
	return t.IsSafeArgs(argsJSON)
}
