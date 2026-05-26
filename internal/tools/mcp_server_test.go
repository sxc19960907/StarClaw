package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/starclaw/starclaw/internal/agent"
)

func TestNewMCPServer_Creation(t *testing.T) {
	reg := RegisterLocalTools()
	srv := NewMCPServer(reg, "test-server", "1.0.0", MCPServerConfig{})

	if srv == nil {
		t.Fatal("NewMCPServer returned nil")
	}

	if srv.name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", srv.name)
	}

	if srv.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", srv.version)
	}

	if srv.registry != reg {
		t.Error("Registry should be the same instance")
	}

	if srv.Server() == nil {
		t.Error("Underlying mcp server should not be nil")
	}
}

func TestNewMCPServer_ToolRegistration(t *testing.T) {
	reg := RegisterLocalTools()
	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	// All registered tools should be available as MCP tools
	for _, tool := range reg.List() {
		info := tool.Info()
		serverTool := mcpSrv.GetTool(info.Name)
		if serverTool == nil {
			t.Errorf("Tool %q not found in MCP server", info.Name)
			continue
		}

		if serverTool.Tool.Name != info.Name {
			t.Errorf("Expected tool name %q, got %q", info.Name, serverTool.Tool.Name)
		}

		if serverTool.Tool.Description != info.Description {
			t.Errorf("Expected description %q, got %q", info.Description, serverTool.Tool.Description)
		}
	}
}

func TestNewMCPServer_ToolListCount(t *testing.T) {
	reg := RegisterLocalTools()
	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	tools := mcpSrv.ListTools()
	if tools == nil {
		t.Fatal("ListTools returned nil")
	}

	if len(tools) != reg.Count() {
		t.Errorf("Expected %d MCP tools, got %d", reg.Count(), len(tools))
	}
}

func TestNewMCPServer_ExposeToolsFilter(t *testing.T) {
	reg := agent.NewToolRegistry()
	reg.Register(&MockTool{name: "first_tool", description: "First mock tool"})
	reg.Register(&MockTool{name: "second_tool", description: "Second mock tool"})

	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{
		ExposeTools: []string{"second_tool"},
	})
	mcpSrv := srv.Server()

	if tool := mcpSrv.GetTool("second_tool"); tool == nil {
		t.Fatal("Expected second_tool to be exposed")
	}

	if tool := mcpSrv.GetTool("first_tool"); tool != nil {
		t.Fatal("Expected first_tool to be hidden by expose filter")
	}

	tools := mcpSrv.ListTools()
	if len(tools) != 1 {
		t.Fatalf("Expected 1 exposed MCP tool, got %d", len(tools))
	}
}

func TestMCPServer_ServerInfoCountsExposedTools(t *testing.T) {
	reg := agent.NewToolRegistry()
	reg.Register(&MockTool{name: "first_tool"})
	reg.Register(&MockTool{name: "second_tool"})

	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{
		ToolTimeout: 7,
		ExposeTools: []string{"first_tool"},
	})

	info := srv.ServerInfo()
	if info["tool_count"] != 1 {
		t.Fatalf("Expected tool_count 1, got %v", info["tool_count"])
	}
	if info["timeout"] != 7 {
		t.Fatalf("Expected timeout 7, got %v", info["timeout"])
	}
}

func TestNewMCPServer_ReadOnlyDetection(t *testing.T) {
	// Create a registry with known read-only and read-write tools
	reg := agent.NewToolRegistry()
	reg.Register(&FileReadTool{}) // read-only
	reg.Register(&BashTool{})     // read-write

	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	// file_read should have ReadOnlyHint = true
	fileReadTool := mcpSrv.GetTool("file_read")
	if fileReadTool == nil {
		t.Fatal("file_read tool not found")
	}
	if fileReadTool.Tool.Annotations.ReadOnlyHint == nil {
		t.Fatal("file_read ReadOnlyHint should not be nil")
	}
	if !*fileReadTool.Tool.Annotations.ReadOnlyHint {
		t.Error("file_read should have ReadOnlyHint = true")
	}

	// bash should have ReadOnlyHint = false
	bashTool := mcpSrv.GetTool("bash")
	if bashTool == nil {
		t.Fatal("bash tool not found")
	}
	if bashTool.Tool.Annotations.ReadOnlyHint == nil {
		t.Fatal("bash ReadOnlyHint should not be nil")
	}
	if *bashTool.Tool.Annotations.ReadOnlyHint {
		t.Error("bash should have ReadOnlyHint = false")
	}
}

func TestMCPServer_ToolExecution_TimesOut(t *testing.T) {
	reg := agent.NewToolRegistry()
	reg.Register(&SlowTool{name: "slow_tool"})

	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{ToolTimeout: 1})
	serverTool := srv.Server().GetTool("slow_tool")
	if serverTool == nil {
		t.Fatal("slow_tool not found in MCP server")
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "slow_tool",
			Arguments: map[string]any{},
		},
	}

	start := time.Now()
	result, err := serverTool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("Expected timeout near 1s, took %s", elapsed)
	}

	if !result.IsError {
		t.Fatal("Expected timeout to return an MCP error result")
	}
	if len(result.Content) == 0 {
		t.Fatal("Expected timeout result content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}
	if !strings.Contains(textContent.Text, context.DeadlineExceeded.Error()) {
		t.Fatalf("Expected deadline error text, got %q", textContent.Text)
	}
}

func TestMCPServer_ToolExecution(t *testing.T) {
	reg := agent.NewToolRegistry()
	mock := &MockTool{name: "mock_tool", description: "A mock tool for testing"}
	reg.Register(mock)

	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	serverTool := mcpSrv.GetTool("mock_tool")
	if serverTool == nil {
		t.Fatal("mock_tool not found in MCP server")
	}

	// Execute the tool via MCP handler
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "mock_tool",
			Arguments: map[string]any{},
		},
	}

	result, err := serverTool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		t.Error("Expected no error in result")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected non-empty content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	if textContent.Text != "mock result" {
		t.Errorf("Expected content 'mock result', got '%s'", textContent.Text)
	}
}

func TestMCPServer_ToolExecution_WithArgs(t *testing.T) {
	reg := agent.NewToolRegistry()
	reg.Register(&FileReadTool{})

	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	serverTool := mcpSrv.GetTool("file_read")
	if serverTool == nil {
		t.Fatal("file_read tool not found")
	}

	// Call with invalid path - should return an error result (not a handler error)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "file_read",
			Arguments: map[string]any{
				"path": "/nonexistent/file.txt",
			},
		},
	}

	result, err := serverTool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler should not return error for validation issues: %v", err)
	}

	if !result.IsError {
		t.Log("Expected isError=true for nonexistent file: ", result.Content)
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected non-empty content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	if textContent.Text == "" {
		t.Error("Expected non-empty error text")
	}
}

func TestMCPServer_InputSchemaPreserved(t *testing.T) {
	reg := agent.NewToolRegistry()
	reg.Register(&FileReadTool{})

	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	serverTool := mcpSrv.GetTool("file_read")
	if serverTool == nil {
		t.Fatal("file_read tool not found")
	}

	// Verify the input schema has the expected properties
	props := serverTool.Tool.InputSchema.Properties
	if props == nil {
		t.Fatal("Properties should not be nil")
	}

	pathProp, ok := props["path"]
	if !ok {
		t.Fatal("Expected 'path' property in schema")
	}

	pathMap, ok := pathProp.(map[string]any)
	if !ok {
		t.Fatal("path property should be a map")
	}

	if pathMap["type"] != "string" {
		t.Errorf("Expected path type 'string', got '%v'", pathMap["type"])
	}

	if pathMap["description"] == "" {
		t.Error("Expected path description to be non-empty")
	}

	// Verify required fields
	required := serverTool.Tool.InputSchema.Required
	if len(required) == 0 {
		t.Fatal("Expected required fields to be non-empty")
	}

	foundPath := false
	for _, r := range required {
		if r == "path" {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Error("Expected 'path' in required fields")
	}
}

func TestMCPServer_ToolHandler_MissingTool(t *testing.T) {
	reg := RegisterLocalTools()
	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	// Should return nil for non-existent tool
	serverTool := mcpSrv.GetTool("nonexistent_tool")
	if serverTool != nil {
		t.Error("Expected nil for non-existent tool")
	}
}

func TestMCPServer_VersionToolRegistered(t *testing.T) {
	reg := RegisterLocalTools()
	RegisterVersionTool(reg, "2.0.0")
	srv := NewMCPServer(reg, "test", "2.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	versionTool := mcpSrv.GetTool("version")
	if versionTool == nil {
		t.Fatal("version tool not found in MCP server")
	}

	if versionTool.Tool.Name != "version" {
		t.Errorf("Expected name 'version', got '%s'", versionTool.Tool.Name)
	}
}

func TestMCPServer_MarshalRoundTrip(t *testing.T) {
	// Test that an mcp.Tool created from a StarClaw tool can be marshaled/unmarshaled
	info := agent.ToolInfo{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{
					"type":        "string",
					"description": "Input value",
				},
			},
		},
		Required: []string{"input"},
	}

	mcpTool := toMCPTool(info, &MockTool{name: "test_tool"})

	// Marshal to JSON
	data, err := json.Marshal(mcpTool)
	if err != nil {
		t.Fatalf("Failed to marshal tool: %v", err)
	}

	// Unmarshal back
	var decoded mcp.Tool
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal tool: %v\nJSON: %s", err, string(data))
	}

	if decoded.Name != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", decoded.Name)
	}

	if decoded.Description != "A test tool" {
		t.Errorf("Expected description 'A test tool', got '%s'", decoded.Description)
	}

	if decoded.InputSchema.Type != "object" {
		t.Errorf("Expected input schema type 'object', got '%s'", decoded.InputSchema.Type)
	}
}

// TestMCPServer_EmptyRegistry tests that an empty registry produces no tools.
func TestMCPServer_EmptyRegistry(t *testing.T) {
	reg := agent.NewToolRegistry()
	srv := NewMCPServer(reg, "test", "1.0.0", MCPServerConfig{})
	mcpSrv := srv.Server()

	tools := mcpSrv.ListTools()
	// ListTools returns nil when empty; that's valid - means no tools
	if tools != nil && len(tools) != 0 {
		t.Errorf("Expected 0 tools for empty registry, got %d", len(tools))
	}

	// Also verify no tool can be found
	if serverTool := mcpSrv.GetTool("anything"); serverTool != nil {
		t.Error("Expected nil for tool lookup on empty server")
	}
}

type SlowTool struct {
	name string
}

func (s *SlowTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        s.name,
		Description: "A slow test tool",
		Parameters:  map[string]any{},
	}
}

func (s *SlowTool) Run(ctx context.Context, args string) (agent.ToolResult, error) {
	<-ctx.Done()
	return agent.ToolResult{}, ctx.Err()
}

func (s *SlowTool) RequiresApproval() bool {
	return false
}
