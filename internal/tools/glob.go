package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/starclaw/starclaw/internal/agent"
)

// GlobTool finds files matching patterns
type GlobTool struct{}

type globArgs struct {
	Pattern string `json:"pattern"`
}

func (t *GlobTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "glob",
		Description: "Find files matching a glob pattern. Supports ** for recursive matching.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{"type": "string", "description": "Glob pattern (e.g., '**/*.go', '*.txt')"},
			},
		},
		Required: []string{"pattern"},
	}
}

func (t *GlobTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args globArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	matches, err := doublestar.Glob(os.DirFS("."), args.Pattern)
	if err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid pattern: %v", err)), nil
	}

	if len(matches) == 0 {
		return agent.ToolResult{Content: "No files matched the pattern."}, nil
	}

	return agent.ToolResult{
		Content: strings.Join(matches, "\n"),
	}, nil
}

func (t *GlobTool) RequiresApproval() bool { return false }

func (t *GlobTool) IsSafeArgs(argsJSON string) bool {
	return true // Glob patterns are read-only and safe
}
