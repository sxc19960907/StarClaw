package tools

import (
	"context"
	"fmt"
	"runtime"

	"github.com/starclaw/starclaw/internal/agent"
)

// VersionTool returns version and environment information.
type VersionTool struct {
	version string
}

// NewVersionTool creates a version tool with the given build version.
func NewVersionTool(version string) *VersionTool {
	return &VersionTool{version: version}
}

func (t *VersionTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "version",
		Description: "Get the current StarClaw version and runtime environment.",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

func (t *VersionTool) Run(_ context.Context, _ string) (agent.ToolResult, error) {
	info := fmt.Sprintf("StarClaw %s\nGo %s\n%s/%s",
		t.version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return agent.ToolResult{Content: info}, nil
}

func (t *VersionTool) RequiresApproval() bool { return false }
