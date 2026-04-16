package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// DirectoryListTool lists directory contents
type DirectoryListTool struct{}

type directoryListArgs struct {
	Path string `json:"path"`
}

func (t *DirectoryListTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "directory_list",
		Description: "List contents of a directory with file sizes and types.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "Directory path (default: current directory)"},
			},
		},
		Required: []string{},
	}
}

func (t *DirectoryListTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args directoryListArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	if args.Path == "" {
		args.Path = "."
	}

	args.Path = ExpandHome(args.Path)

	// Security check
	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	entries, err := os.ReadDir(args.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ValidationError(fmt.Sprintf("directory not found: %s", args.Path)), nil
		}
		if os.IsPermission(err) {
			return agent.PermissionError(fmt.Sprintf("permission denied: %s", args.Path)), nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("error reading directory: %v", err),
			IsError: true,
		}, nil
	}

	var lines []string
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		prefix := "📄"
		size := ""
		if entry.IsDir() {
			prefix = "📁"
		} else {
			size = fmt.Sprintf(" (%d bytes)", info.Size())
		}

		lines = append(lines, fmt.Sprintf("%s %s%s", prefix, entry.Name(), size))
	}

	if len(lines) == 0 {
		return agent.ToolResult{Content: "Directory is empty."}, nil
	}

	return agent.ToolResult{
		Content: strings.Join(lines, "\n"),
	}, nil
}

func (t *DirectoryListTool) RequiresApproval() bool { return false }

func (t *DirectoryListTool) IsSafeArgs(argsJSON string) bool {
	return true // directory_list is read-only
}
