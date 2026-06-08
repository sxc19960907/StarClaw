package tools

import (
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/daemon/desktop_rpc"
	"github.com/starclaw/starclaw/internal/mcp"
	"github.com/starclaw/starclaw/internal/schedule"
	"github.com/starclaw/starclaw/internal/session"
)

// RegisterLocalTools registers all local tools.
// An optional ToolsConfig can be passed to configure tool behaviour
// (e.g. BashMaxOutput). When omitted, tool defaults apply.
func RegisterLocalTools(toolsConfig ...config.ToolsConfig) *agent.ToolRegistry {
	reg := agent.NewToolRegistry()

	// Extract config values (or zero-value defaults)
	var tc config.ToolsConfig
	if len(toolsConfig) > 0 {
		tc = toolsConfig[0]
	}

	// File tools
	reg.Register(&DocumentTextTool{})
	reg.Register(&ArchiveInspectTool{})
	reg.Register(&ArchiveExtractTool{})
	reg.Register(&FileReadTool{})
	reg.Register(&FileWriteTool{})
	reg.Register(&FileEditTool{})
	reg.Register(&FilePreviewTool{})

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
	reg.Register(&BashTool{MaxOutput: tc.BashMaxOutput})

	// Memory tools
	reg.Register(&MemoryAppendTool{})
	reg.Register(&MemoryTool{})

	// Wait tool
	reg.Register(&WaitTool{})

	// Clipboard tool
	reg.Register(&ClipboardTool{})

	// Notify tool
	reg.Register(&NotifyTool{})

	// Screenshot tool
	reg.Register(&ScreenshotTool{})

	// AppleScript tool
	reg.Register(&AppleScriptTool{})

	// Visible terminal workspace tool
	reg.Register(NewTerminalWorkspaceTool())

	// macOS Accessibility tool
	reg.Register(&AccessibilityTool{})

	// macOS Computer control tool
	reg.Register(&ComputerTool{})

	// Browser tool
	reg.Register(&BrowserTool{})

	// Process management tool
	reg.Register(NewProcessTool(0))

	// Image processing tool
	reg.Register(&ImagingTool{})

	// Publish to web tool
	reg.Register(NewPublishToWebTool())
	reg.Register(NewListPublishedFilesTool())
	reg.Register(NewRetractPublishedFileTool())

	// Skills tools
	skillsDir := config.StarclawDir()
	if skillsDir != "" {
		skillsDir = skillsDir + "/skills"
	}
	reg.Register(NewUseSkillTool(skillsDir))
	reg.Register(NewSkillTool())

	// Schedule tools
	starclawDir := config.StarclawDir()
	if starclawDir != "" {
		scheduleMgr := schedule.NewManager(starclawDir + "/schedules.json")
		for _, t := range NewScheduleTools(scheduleMgr) {
			reg.Register(t)
		}
	}

	return reg
}

// RegisterCalendarTools registers Desktop-RPC-backed calendar tools when a
// local Desktop broker is available. No-op for nil registries or brokers so
// CLI/TUI registries do not expose unavailable calendar tools.
func RegisterCalendarTools(reg *agent.ToolRegistry, broker *desktop_rpc.Broker) {
	if reg == nil || broker == nil {
		return
	}
	for _, tool := range NewCalendarTools(broker) {
		reg.Register(tool)
	}
}

// RegisterImageTools registers provider-backed image generation/editing tools
// only when an explicit provider client is supplied. The default local registry
// does not call this function, preserving local-first behavior.
func RegisterImageTools(reg *agent.ToolRegistry, client imageProvider, cdnPrefix string) {
	if reg == nil || client == nil {
		return
	}
	reg.Register(NewGenerateImageTool(client))
	reg.Register(NewEditImageTool(client, cdnPrefix))
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
