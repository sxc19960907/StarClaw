package agent

import (
	"context"
	"strings"
	"testing"
)

func TestToolSearchTool_Info(t *testing.T) {
	reg := NewToolRegistry()
	deferred := map[string]bool{"mcp_github": true}
	ts := newToolSearchTool(reg, deferred)

	info := ts.Info()
	if info.Name != "tool_search" {
		t.Errorf("Name = %q, want 'tool_search'", info.Name)
	}
	if len(info.Required) != 1 || info.Required[0] != "query" {
		t.Error("Required should be ['query']")
	}
}

func TestToolSearchTool_SelectMatch(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "mcp_github", description: "GitHub API"})
	deferred := map[string]bool{"mcp_github": true}

	ts := newToolSearchTool(reg, deferred)
	result, err := ts.Run(context.Background(), `{"query":"select:mcp_github"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(result.Content, "LOADED:mcp_github") {
		t.Errorf("Expected LOADED:mcp_github, got: %s", result.Content)
	}
	if !strings.Contains(result.Content, "mcp_github") {
		t.Error("Should include tool schema")
	}
}

func TestToolSearchTool_KeywordSearch(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "mcp_github", description: "GitHub API for PR management"})
	reg.Register(&MockTool{name: "mcp_fetch", description: "Web fetch tool"})
	deferred := map[string]bool{"mcp_github": true, "mcp_fetch": true}

	ts := newToolSearchTool(reg, deferred)
	result, err := ts.Run(context.Background(), `{"query":"github"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(result.Content, "mcp_github") {
		t.Errorf("Should find mcp_github by keyword, got: %s", result.Content)
	}
}

func TestToolSearchTool_NoMatch(t *testing.T) {
	reg := NewToolRegistry()
	deferred := map[string]bool{}

	ts := newToolSearchTool(reg, deferred)
	result, err := ts.Run(context.Background(), `{"query":"nonexistent"}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(result.Content, "No matching") {
		t.Errorf("Expected 'No matching', got: %s", result.Content)
	}
}

func TestToolSearchTool_EmptyQuery(t *testing.T) {
	reg := NewToolRegistry()
	deferred := map[string]bool{}

	ts := newToolSearchTool(reg, deferred)
	result, err := ts.Run(context.Background(), `{"query":""}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("Empty query should return validation error")
	}
}

func TestDeferredToolNames(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "local_a"})
	reg.Register(&MockTool{name: "mcp_a", source: SourceMCP})

	names := DeferredToolNames(reg)
	if names["local_a"] {
		t.Error("local_a should not be deferred")
	}
	if !names["mcp_a"] {
		t.Error("mcp_a should be deferred")
	}
}

func TestDeferredToolSummaries(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "local_a", description: "Local tool"})
	reg.Register(&MockTool{name: "mcp_a", description: "MCP tool", source: SourceMCP})

	summaries := DeferredToolSummaries(reg)
	if len(summaries) != 1 {
		t.Fatalf("Expected 1 deferred summary, got %d", len(summaries))
	}
	if summaries[0].Name != "mcp_a" {
		t.Errorf("Name = %q", summaries[0].Name)
	}
}

func TestBuildLocalOnlySchemas(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "local_a"})
	reg.Register(&MockTool{name: "mcp_a", source: SourceMCP})

	schemas := BuildLocalOnlySchemas(reg)
	if len(schemas) != 1 {
		t.Fatalf("Expected 1 local schema, got %d", len(schemas))
	}
	if schemas[0].Function.Name != "local_a" {
		t.Errorf("Name = %q", schemas[0].Function.Name)
	}
}
