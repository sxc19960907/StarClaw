package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// FileReadTool reads file contents
type FileReadTool struct{}

type fileReadArgs struct {
	Path   string `json:"path"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func (t *FileReadTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "file_read",
		Description: "Read a file's contents with line numbers. Use offset and limit for large files.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":   map[string]any{"type": "string", "description": "Absolute or relative file path"},
				"offset": map[string]any{"type": "integer", "description": "Start line (0-based, default 0)"},
				"limit":  map[string]any{"type": "integer", "description": "Max lines to read (default: all)"},
			},
		},
		Required: []string{"path"},
	}
}

func (t *FileReadTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args fileReadArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	args.Path = ExpandHome(args.Path)

	// Security check
	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	data, err := os.ReadFile(args.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ValidationError(fmt.Sprintf("file not found: %s", args.Path)), nil
		}
		if os.IsPermission(err) {
			return agent.PermissionError(fmt.Sprintf("permission denied: %s", args.Path)), nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("error reading file: %v", err),
			IsError: true,
		}, nil
	}

	lines := strings.Split(string(data), "\n")
	start := args.Offset
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		start = len(lines)
	}

	end := len(lines)
	if args.Limit > 0 && start+args.Limit < end {
		end = start + args.Limit
	}

	var sb strings.Builder
	for i := start; i < end; i++ {
		fmt.Fprintf(&sb, "%4d | %s\n", i+1, lines[i])
	}

	return agent.ToolResult{Content: sb.String()}, nil
}

func (t *FileReadTool) RequiresApproval() bool { return true }

func (t *FileReadTool) IsReadOnlyCall(string) bool { return true }

func (t *FileReadTool) IsSafeArgs(argsJSON string) bool {
	var args fileReadArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return IsPathUnderCWD(args.Path)
}
