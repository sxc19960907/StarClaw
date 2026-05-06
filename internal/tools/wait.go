package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
)

// WaitTool pauses execution for a specified duration.
// Use this instead of `bash sleep` to avoid the loop detection sleep penalty.
type WaitTool struct{}

type waitArgs struct {
	Seconds float64 `json:"seconds"`
}

func (t *WaitTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "wait",
		Description: "Wait for a specified number of seconds. Use this instead of 'sleep' in bash to avoid triggering the sleep detector. Useful for waiting on builds, services, or timed operations. Max 30 seconds.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"seconds": map[string]any{
					"type":        "number",
					"description": "Seconds to wait (default: 5, max: 30)",
				},
			},
		},
	}
}

func (t *WaitTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args waitArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	secs := args.Seconds
	if secs <= 0 {
		secs = 5
	}
	if secs > 30 {
		return agent.ValidationError(fmt.Sprintf("max wait is 30 seconds, got %.0f", secs)), nil
	}

	duration := time.Duration(secs * float64(time.Second))

	select {
	case <-time.After(duration):
	case <-ctx.Done():
		return agent.ToolResult{Content: "wait cancelled", IsError: true}, nil
	}

	if secs == float64(int(secs)) {
		return agent.ToolResult{Content: fmt.Sprintf("Waited %.0f seconds.", secs)}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Waited %.1f seconds.", secs)}, nil
}

func (t *WaitTool) RequiresApproval() bool { return false }
