package tools

import (
	"context"
	"encoding/json"

	"github.com/starclaw/starclaw/internal/agent"
	ctxwin "github.com/starclaw/starclaw/internal/context"
)

// MemoryAppendTool appends entries to MEMORY.md via BoundedAppend.
// The memory directory is resolved from context (set by AgentLoop.Run).
// Writes are atomic, flock-protected, and auto-overflow when the line limit is reached.
// IMPORTANT: does NOT reload memory after write — preserves prompt cache.
type MemoryAppendTool struct{}

type memoryAppendArgs struct {
	Content string `json:"content"`
}

func (t *MemoryAppendTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "memory_append",
		Description: "Append new entries to MEMORY.md. Use this instead of file_write or file_edit for memory updates. Writes are atomic, flock-protected, and auto-overflow to detail files when MEMORY.md exceeds the line limit.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"content": map[string]any{
					"type":        "string",
					"description": "New entries to append (markdown bullet points, one per line)",
				},
			},
		},
		Required: []string{"content"},
	}
}

func (t *MemoryAppendTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args memoryAppendArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	if args.Content == "" {
		return agent.ValidationError("content is required"), nil
	}

	memoryDir := ctxwin.MemoryDirFromContext(ctx)
	if memoryDir == "" {
		return agent.ToolResult{
			Content: "memory_append: memory directory not available (set --agent or configure memory)",
			IsError: true,
		}, nil
	}

	if err := ctxwin.BoundedAppend(memoryDir, args.Content); err != nil {
		return agent.ToolResult{Content: "memory_append failed: " + err.Error(), IsError: true}, nil
	}

	return agent.ToolResult{Content: "Memory updated. New entries will be available in the next session."}, nil
}

func (t *MemoryAppendTool) RequiresApproval() bool { return false }
