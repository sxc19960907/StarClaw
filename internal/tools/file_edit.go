package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// FileEditTool edits files by replacing content
type FileEditTool struct{}

type fileEditArgs struct {
	Path     string `json:"path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

func (t *FileEditTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "file_edit",
		Description: "Edit a file by replacing old_string with new_string. Replaces only the first occurrence.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":       map[string]any{"type": "string", "description": "File path"},
				"old_string": map[string]any{"type": "string", "description": "String to replace"},
				"new_string": map[string]any{"type": "string", "description": "Replacement string"},
			},
		},
		Required: []string{"path", "old_string", "new_string"},
	}
}

func (t *FileEditTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args fileEditArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	args.Path = ExpandHome(args.Path)

	// Security check
	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	// Read file
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

	content := string(data)

	// Check if old_string exists
	if !strings.Contains(content, args.OldString) {
		return agent.ValidationError(fmt.Sprintf("old_string not found in file: %s", args.OldString)), nil
	}

	// Replace only first occurrence
	newContent := strings.Replace(content, args.OldString, args.NewString, 1)

	// Write back
	if err := os.WriteFile(args.Path, []byte(newContent), 0644); err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("error writing file: %v", err),
			IsError: true,
		}, nil
	}

	return agent.ToolResult{
		Content: fmt.Sprintf("File edited: %s (replaced '%s' with '%s')", args.Path, args.OldString, args.NewString),
	}, nil
}

func (t *FileEditTool) RequiresApproval() bool { return true }

func (t *FileEditTool) IsSafeArgs(argsJSON string) bool {
	var args fileEditArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return IsPathUnderCWD(args.Path)
}
