package tools

import (
	"context"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

func TestRegisterLocalTools(t *testing.T) {
	reg := RegisterLocalTools()

	if reg == nil {
		t.Fatal("RegisterLocalTools() returned nil")
	}

	// Check that all expected tools are registered
	expectedTools := []string{
		"file_read",
		"file_write",
		"file_edit",
		"glob",
		"directory_list",
		"grep",
		"think",
		"system_info",
		"http",
		"bash",
	}

	if reg.Count() != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), reg.Count())
	}

	// Verify each tool can be retrieved
	for _, toolName := range expectedTools {
		tool, ok := reg.Get(toolName)
		if !ok {
			t.Errorf("Expected to find tool '%s' in registry", toolName)
			continue
		}

		info := tool.Info()
		if info.Name != toolName {
			t.Errorf("Expected tool name '%s', got '%s'", toolName, info.Name)
		}
	}
}

func TestRegisterLocalTools_ToolInfo(t *testing.T) {
	reg := RegisterLocalTools()

	// Check file_read tool info
	if tool, ok := reg.Get("file_read"); ok {
		info := tool.Info()
		if info.Description == "" {
			t.Error("file_read tool should have a description")
		}
		if info.Parameters == nil {
			t.Error("file_read tool should have parameters")
		}
	}

	// Check bash tool has RequiresApproval set correctly
	if tool, ok := reg.Get("bash"); ok {
		if !tool.RequiresApproval() {
			t.Error("bash tool should require approval")
		}
	}
}

func TestRegisterLocalTools_List(t *testing.T) {
	reg := RegisterLocalTools()

	tools := reg.List()
	if len(tools) == 0 {
		t.Fatal("Expected non-empty tool list")
	}

	// Verify all tools are sorted by name
	names := reg.Names()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Error("Tools should be sorted alphabetically by name")
		}
	}
}

func TestRegisterLocalTools_CloneIndependence(t *testing.T) {
	reg := RegisterLocalTools()

	// Clone should be independent
	clone := reg.Clone()

	// Modify clone
	clone.Remove("bash")

	// Original should be unchanged
	if _, ok := reg.Get("bash"); !ok {
		t.Error("Modifying clone should not affect original registry")
	}

	// Clone should not have bash
	if _, ok := clone.Get("bash"); ok {
		t.Error("Clone should not have bash after removal")
	}
}

func TestRegisterLocalTools_Filter(t *testing.T) {
	reg := RegisterLocalTools()

	// Test FilterByAllow
	allowed := reg.FilterByAllow([]string{"file_read", "bash"})
	if allowed.Count() != 2 {
		t.Errorf("Expected 2 allowed tools, got %d", allowed.Count())
	}

	// Test FilterByDeny
	denied := reg.FilterByDeny([]string{"bash"})
	if denied.Count() != reg.Count()-1 {
		t.Errorf("Expected %d tools after denial, got %d", reg.Count()-1, denied.Count())
	}

	// Bash should not be in filtered registry
	if _, ok := denied.Get("bash"); ok {
		t.Error("bash should not be in FilterByDeny result")
	}
}

// MockTool is a minimal tool implementation for testing
type MockTool struct {
	name        string
	description string
}

func (m *MockTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        m.name,
		Description: m.description,
		Parameters:  map[string]any{},
	}
}

func (m *MockTool) Run(ctx context.Context, args string) (agent.ToolResult, error) {
	return agent.ToolResult{Content: "mock result"}, nil
}

func (m *MockTool) RequiresApproval() bool {
	return false
}
