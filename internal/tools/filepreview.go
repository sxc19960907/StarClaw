package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// FilePreviewTool previews the beginning of a file without loading the entire
// file into memory. Uses bufio.Scanner for efficient line-by-line reading.
type FilePreviewTool struct{}

type filePreviewArgs struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
}

func (t *FilePreviewTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "file_preview",
		Description: "Preview the first N lines of a file. Uses buffered reading — does not load the entire file into memory. Useful for quickly inspecting file contents.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Absolute or relative file path",
				},
				"lines": map[string]any{
					"type":        "integer",
					"description": "Number of lines to preview (default: 20)",
				},
			},
			"required": []string{"path"},
		},
	}
}

func (t *FilePreviewTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args filePreviewArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	if args.Path == "" {
		return agent.ValidationError("path is required"), nil
	}

	args.Path = ExpandHome(args.Path)

	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	lines := args.Lines
	if lines <= 0 {
		lines = 20
	}

	f, err := os.Open(args.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ValidationError(fmt.Sprintf("file not found: %s", args.Path)), nil
		}
		if os.IsPermission(err) {
			return agent.PermissionError(fmt.Sprintf("permission denied: %s", args.Path)), nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("error opening file: %v", err),
			IsError: true,
		}, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var sb strings.Builder
	lineNum := 0
	moreLines := false

	for scanner.Scan() {
		if lineNum < lines {
			lineNum++
			fmt.Fprintf(&sb, "%4d | %s\n", lineNum, scanner.Text())
		} else {
			moreLines = true
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("error reading file: %v", err),
			IsError: true,
		}, nil
	}

	if moreLines {
		sb.WriteString(fmt.Sprintf("... (file truncated, showing first %d lines)\n", lines))
	}

	return agent.ToolResult{Content: sb.String()}, nil
}

func (t *FilePreviewTool) RequiresApproval() bool { return true }

func (t *FilePreviewTool) IsReadOnlyCall(string) bool { return true }
