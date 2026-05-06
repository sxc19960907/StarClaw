package tools

import (
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/mcp"
	"github.com/starclaw/starclaw/internal/session"
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

	// Memory tool
	reg.Register(&MemoryAppendTool{})

	// Wait tool
	reg.Register(&WaitTool{})

	// Skills tool
	skillsDir := config.StarclawDir()
	if skillsDir != "" {
		skillsDir = skillsDir + "/skills"
	}
	reg.Register(NewUseSkillTool(skillsDir))

	return reg
}

// RegisterVersionTool registers the version tool with the build version.
func RegisterVersionTool(reg *agent.ToolRegistry, version string) {
	reg.Register(NewVersionTool(version))
}

// RegisterSessionSearch registers the session search tool.
func RegisterSessionSearch(reg *agent.ToolRegistry, mgr interface{}) {
	// Import cycle avoidance: session.Manager is wired via concrete type check
	if sm, ok := mgr.(*session.Manager); ok {
		reg.Register(NewSessionSearchTool(sm))
	}
}

// RegisterMCPTools registers MCP remote tools in the tool registry.
// Each RemoteTool is wrapped in an MCPTool adapter.
func RegisterMCPTools(reg *agent.ToolRegistry, remoteTools []mcp.RemoteTool, manager *mcp.ClientManager) {
	for _, rt := range remoteTools {
		adapter := NewMCPTool(rt.ServerName, rt.Tool, manager)
		reg.Register(adapter)
	}
}
