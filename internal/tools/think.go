// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/starclaw/starclaw/internal/agent"
)

// ThinkTool lets the model reason or plan before acting. The model calls this
// tool instead of outputting plan text, giving the loop an explicit continuation
// signal (stop_reason: tool_use) rather than relying on text heuristics.
type ThinkTool struct{}

// thinkArgs represents the arguments for the think tool.
type thinkArgs struct {
	Thought string `json:"thought"`
}

// Info returns the tool definition for the LLM.
func (t *ThinkTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "think",
		Description: "Use this to plan or reason through complex multi-step tasks before acting. Always use this instead of outputting plans as plain text.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"thought": map[string]any{
					"type":        "string",
					"description": "Your reasoning or plan",
				},
			},
		},
		Required: []string{"thought"},
	}
}

// Run executes the think tool.
func (t *ThinkTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args thinkArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("invalid arguments: %v", err),
			IsError: true,
		}, nil
	}
	if args.Thought == "" {
		return agent.ToolResult{
			Content: "thought is required",
			IsError: true,
		}, nil
	}
	return agent.ToolResult{Content: args.Thought}, nil
}

// RequiresApproval returns false because this is a read-only operation.
func (t *ThinkTool) RequiresApproval() bool { return false }

// IsReadOnlyCall returns true for all invocations.
func (t *ThinkTool) IsReadOnlyCall(string) bool { return true }
