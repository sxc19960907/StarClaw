package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/starclaw/starclaw/internal/agent"
)

// FileWriteTool writes content to files
type FileWriteTool struct{}

type fileWriteArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (t *FileWriteTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "file_write",
		Description: "Write content to a file. Creates the file if it doesn't exist, overwrites if it does.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]any{"type": "string", "description": "File path"},
				"content": map[string]any{"type": "string", "description": "Content to write"},
			},
		},
		Required: []string{"path", "content"},
	}
}

func (t *FileWriteTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args fileWriteArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	args.Path = ExpandHome(args.Path)

	// Security check
	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	// Create parent directories if needed
	dir := filepath.Dir(args.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("error creating directory: %v", err),
			IsError: true,
		}, nil
	}

	// Check if file exists
	_, err := os.Stat(args.Path)
	existed := !os.IsNotExist(err)

	if err := os.WriteFile(args.Path, []byte(args.Content), 0644); err != nil {
		if os.IsPermission(err) {
			return agent.PermissionError(fmt.Sprintf("permission denied: %s", args.Path)), nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("error writing file: %v", err),
			IsError: true,
		}, nil
	}

	if existed {
		return agent.ToolResult{Content: fmt.Sprintf("File overwritten: %s", args.Path)}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("File created: %s", args.Path)}, nil
}

func (t *FileWriteTool) RequiresApproval() bool { return true }

func (t *FileWriteTool) IsSafeArgs(argsJSON string) bool {
	var args fileWriteArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return IsPathUnderCWD(args.Path)
}
