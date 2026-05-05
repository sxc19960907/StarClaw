package tools

import (
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/mcp"
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

	// Skills tool
	skillsDir := config.StarclawDir()
	if skillsDir != "" {
		skillsDir = skillsDir + "/skills"
	}
	reg.Register(NewUseSkillTool(skillsDir))

	return reg
}

// RegisterMCPTools registers MCP remote tools in the tool registry.
// Each RemoteTool is wrapped in an MCPTool adapter.
func RegisterMCPTools(reg *agent.ToolRegistry, remoteTools []mcp.RemoteTool, manager *mcp.ClientManager) {
	for _, rt := range remoteTools {
		adapter := NewMCPTool(rt.ServerName, rt.Tool, manager)
		reg.Register(adapter)
	}
}
