package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/starclaw/starclaw/internal/agent"
)

// MCPServer wraps a StarClaw ToolRegistry as an MCP stdio server.
// It converts all local tools to MCP tool definitions and handles
// tools/list and tools/call MCP methods.
type MCPServer struct {
	registry *agent.ToolRegistry
	name     string
	version  string
	srv      *mcpserver.MCPServer
}

// NewMCPServer creates a new MCP server that exposes all registered local tools.
func NewMCPServer(registry *agent.ToolRegistry, name, version string) *MCPServer {
	s := &MCPServer{
		registry: registry,
		name:     name,
		version:  version,
	}

	// Create the underlying mcp-go server with tool capabilities
	srv := mcpserver.NewMCPServer(
		name,
		version,
		mcpserver.WithInstructions("StarClaw local tools available via MCP"),
	)

	// Register all tools from the registry as MCP tools
	for _, tool := range registry.List() {
		info := tool.Info()
		mcpTool := toMCPTool(info, tool)
		handler := toMCPHandler(tool)

		srv.AddTool(mcpTool, handler)
	}

	s.srv = srv
	return s
}

// Serve starts the MCP stdio server. Blocks until stdin closes.
func (s *MCPServer) Serve(ctx context.Context) error {
	return mcpserver.ServeStdio(s.srv)
}

// Server returns the underlying mcp-go server for testing and inspection.
func (s *MCPServer) Server() *mcpserver.MCPServer {
	return s.srv
}

// toMCPTool converts a StarClaw ToolInfo to an mcp.Tool definition.
func toMCPTool(info agent.ToolInfo, tool agent.Tool) mcp.Tool {
	opts := []mcp.ToolOption{
		mcp.WithDescription(info.Description),
	}

	mcpTool := mcp.NewTool(info.Name, opts...)

	// Preserve the full parameter schema from the tool definition
	if params, ok := info.Parameters["properties"].(map[string]any); ok {
		mcpTool.InputSchema.Properties = params
	}

	if info.Required != nil {
		mcpTool.InputSchema.Required = info.Required
	}

	// Set ReadOnlyHint if the tool implements ReadOnlyChecker
	if checker, ok := tool.(agent.ReadOnlyChecker); ok {
		// Use an empty args string to check read-only status.
		// ReadOnlyChecker.IsReadOnlyCall is typically static (ignores args).
		isReadOnly := checker.IsReadOnlyCall("{}")
		mcpTool.Annotations.ReadOnlyHint = &isReadOnly
	}

	return mcpTool
}

// toMCPHandler creates an MCP ToolHandlerFunc that delegates to a StarClaw tool.
// In server mode, tools requiring approval are auto-approved (the MCP consumer
// handles its own authorization).
func toMCPHandler(tool agent.Tool) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Marshal arguments to JSON string for the StarClaw tool
		argsJSON := "{}"
		if req.Params.Arguments != nil {
			if args, ok := req.Params.Arguments.(map[string]any); ok && len(args) > 0 {
				data, err := json.Marshal(args)
				if err != nil {
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							mcp.TextContent{Text: fmt.Sprintf("invalid arguments: %v", err)},
						},
						IsError: true,
					}, nil
				}
				argsJSON = string(data)
			}
		}

		// Execute the tool (auto-approved — no RequiresApproval check)
		result, err := tool.Run(ctx, argsJSON)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Text: fmt.Sprintf("tool error: %v", err)},
				},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Text: result.Content},
			},
			IsError: result.IsError,
		}, nil
	}
}
