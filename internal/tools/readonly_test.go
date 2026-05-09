package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

// mockReadOnlyTool implements both Tool and ReadOnlyChecker.
type mockReadOnlyTool struct {
	readOnly bool
}

func (m *mockReadOnlyTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "mock_readonly_tool",
		Description: "Mock tool for testing",
		Parameters:  map[string]any{},
	}
}

func (m *mockReadOnlyTool) Run(_ context.Context, _ string) (agent.ToolResult, error) {
	return agent.ToolResult{Content: "mock ran"}, nil
}

func (m *mockReadOnlyTool) RequiresApproval() bool { return false }

func (m *mockReadOnlyTool) IsReadOnlyCall(_ string) bool { return m.readOnly }

// mockWriteTool implements Tool but NOT ReadOnlyChecker.
type mockWriteTool struct{}

func (m *mockWriteTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "mock_write_tool",
		Description: "Mock write-only tool",
		Parameters:  map[string]any{},
	}
}

func (m *mockWriteTool) Run(_ context.Context, _ string) (agent.ToolResult, error) {
	return agent.ToolResult{Content: "write ran"}, nil
}

func (m *mockWriteTool) RequiresApproval() bool { return true }

func TestReadOnlyMode_ReadOnlyCall_Passes(t *testing.T) {
	inner := &mockReadOnlyTool{readOnly: true}
	wrapper := NewReadOnlyMode(inner)

	result, err := wrapper.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Content != "mock ran" {
		t.Errorf("expected 'mock ran', got %q", result.Content)
	}
	if result.IsError {
		t.Error("expected no error for read-only call")
	}
}

func TestReadOnlyMode_WriteCall_Blocked(t *testing.T) {
	inner := &mockReadOnlyTool{readOnly: false}
	wrapper := NewReadOnlyMode(inner)

	result, err := wrapper.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for write call")
	}
	if !strings.Contains(result.Content, "read-only mode") {
		t.Errorf("expected read-only mode error, got: %s", result.Content)
	}
}

func TestReadOnlyMode_ToolWithoutChecker_Blocked(t *testing.T) {
	inner := &mockWriteTool{}
	wrapper := NewReadOnlyMode(inner)

	result, err := wrapper.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for tool without ReadOnlyChecker")
	}
	if !strings.Contains(result.Content, "read-only mode") {
		t.Errorf("expected read-only mode error, got: %s", result.Content)
	}
}

func TestReadOnlyMode_NilInner_Blocked(t *testing.T) {
	wrapper := NewReadOnlyMode(nil)

	result, err := wrapper.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nil inner tool")
	}
}

func TestReadOnlyMode_Info(t *testing.T) {
	inner := &mockReadOnlyTool{readOnly: true}
	wrapper := NewReadOnlyMode(inner)

	info := wrapper.Info()
	if info.Name != "mock_readonly_tool" {
		t.Errorf("expected mock_readonly_tool, got %q", info.Name)
	}
}

func TestReadOnlyMode_Info_NilInner(t *testing.T) {
	wrapper := NewReadOnlyMode(nil)
	info := wrapper.Info()
	if info.Name != "readonly" {
		t.Errorf("expected 'readonly', got %q", info.Name)
	}
}

func TestReadOnlyMode_RequiresApproval(t *testing.T) {
	inner := &mockWriteTool{}
	wrapper := NewReadOnlyMode(inner)

	if !wrapper.RequiresApproval() {
		t.Error("expected RequiresApproval to delegate to inner tool")
	}
}
