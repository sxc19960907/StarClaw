// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// SystemInfoTool provides system information about the current machine.
type SystemInfoTool struct{}

// Info returns the tool definition for the LLM.
func (t *SystemInfoTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "system_info",
		Description: "Get system information: OS, architecture, hostname, CPU count, memory, and disk usage.",
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Required: nil,
	}
}

// Run executes the system_info tool.
func (t *SystemInfoTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	hostname, _ := os.Hostname()

	var sb strings.Builder
	fmt.Fprintf(&sb, "OS: %s\n", runtime.GOOS)
	fmt.Fprintf(&sb, "Arch: %s\n", runtime.GOARCH)
	fmt.Fprintf(&sb, "Hostname: %s\n", hostname)
	fmt.Fprintf(&sb, "CPUs: %d\n", runtime.NumCPU())

	// Platform-specific info (initialized in platform-specific files)
	if memInfo := getMemoryInfo(); memInfo != "" {
		fmt.Fprintf(&sb, "\nMemory:\n%s", memInfo)
	}
	if diskInfo := getDiskInfo(); diskInfo != "" {
		fmt.Fprintf(&sb, "\nDisk:\n%s", diskInfo)
	}

	return agent.ToolResult{Content: sb.String()}, nil
}

// RequiresApproval returns false because this is a read-only operation.
func (t *SystemInfoTool) RequiresApproval() bool { return false }

// IsReadOnlyCall returns true for all invocations.
func (t *SystemInfoTool) IsReadOnlyCall(string) bool { return true }

// Platform-specific functions (initialized in platform-specific files)
var getMemoryInfo func() string
var getDiskInfo func() string
