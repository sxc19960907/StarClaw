package tools

import (
	"github.com/starclaw/starclaw/internal/agent"
)

// RegisterLocalTools registers all local tools
func RegisterLocalTools() *agent.ToolRegistry {
	reg := agent.NewToolRegistry()

	// File tools
	reg.Register(&FileReadTool{})
	reg.Register(&FileWriteTool{})
	reg.Register(&FileEditTool{})

	// Directory tools
	reg.Register(&GlobTool{})
	reg.Register(&DirectoryListTool{})
	reg.Register(&GrepTool{})

	// Reasoning tool
	reg.Register(&ThinkTool{})

	// System information tool
	reg.Register(&SystemInfoTool{})

	// HTTP tool
	reg.Register(&HTTPTool{})

	// System tools
	reg.Register(&BashTool{})

	return reg
}
