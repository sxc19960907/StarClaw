package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

// CloudDelegateTool delegates sub-tasks to a remote cloud agent.
type CloudDelegateTool struct {
	cloudClient *client.CloudClient
	onProgress  func(client.CloudProgress)
}

type cloudDelegateArgs struct {
	Task    string `json:"task"`
	Agent   string `json:"agent,omitempty"`
	Context string `json:"context,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// NewCloudDelegateTool creates a new cloud delegate tool.
// cloudClient may be nil if cloud is disabled.
func NewCloudDelegateTool(cloudClient *client.CloudClient, onProgress func(client.CloudProgress)) *CloudDelegateTool {
	return &CloudDelegateTool{
		cloudClient: cloudClient,
		onProgress:  onProgress,
	}
}

func (t *CloudDelegateTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "cloud_delegate",
		Description: "Delegate a sub-task to a remote cloud agent. Use for research, code generation, or review tasks that benefit from parallel execution.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task": map[string]any{
					"type":        "string",
					"description": "Description of the task to delegate",
				},
				"agent": map[string]any{
					"type":        "string",
					"description": "Optional: specific remote agent to use",
				},
				"context": map[string]any{
					"type":        "string",
					"description": "Optional: additional context for the remote agent",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "Optional: timeout in seconds (default: 300)",
				},
			},
		},
		Required: []string{"task"},
	}
}

func (t *CloudDelegateTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	if t.cloudClient == nil {
		return agent.ToolResult{
			Content: "Cloud delegation is not configured. Set cloud.enabled=true and cloud.endpoint in config.",
			IsError: true,
		}, nil
	}

	var args cloudDelegateArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	if args.Task == "" {
		return agent.ValidationError("task is required"), nil
	}

	timeout := 300 * time.Second
	if args.Timeout > 0 {
		timeout = time.Duration(args.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req := client.CloudDelegateRequest{
		Task:    args.Task,
		Agent:   args.Agent,
		Context: args.Context,
	}

	resp, err := t.cloudClient.DelegateStream(ctx, req, t.onProgress)
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("Cloud delegation failed: %v", err),
			IsError: true,
		}, nil
	}

	if resp.Error != "" {
		return agent.ToolResult{
			Content: fmt.Sprintf("Cloud agent returned error: %s", resp.Error),
			IsError: true,
		}, nil
	}

	return agent.ToolResult{Content: resp.Result}, nil
}

func (t *CloudDelegateTool) RequiresApproval() bool {
	return true
}
