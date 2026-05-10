package agent

import (
	"context"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

// NewMockAgentLoop creates an AgentLoop wired to a MockClient and an
// empty registry. Useful for tests that only need the loop scaffolding.
func NewMockAgentLoop() *AgentLoop {
	return NewAgentLoop(client.NewMockClient(), NewToolRegistry())
}

// TestTool returns a simple Tool that records its name for assertions.
func TestTool(name string) Tool {
	return &testTool{name: name, description: "test tool: " + name}
}

// AssertNoError fails the test immediately when err is not nil.
func AssertNoError(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

// TestMessage is a shorthand for creating a client.Message.
func TestMessage(role, content string) client.Message {
	return client.Message{Role: role, Content: content}
}

// testTool implements the Tool interface for testing.
type testTool struct {
	name             string
	description      string
	requiresApproval bool
	err              error
}

func (tt *testTool) Info() ToolInfo {
	return ToolInfo{
		Name:        tt.name,
		Description: tt.description,
		Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

func (tt *testTool) Run(_ context.Context, _ string) (ToolResult, error) {
	if tt.err != nil {
		return ToolResult{Content: tt.err.Error(), IsError: true}, tt.err
	}
	return ToolResult{Content: tt.name + " executed"}, nil
}

func (tt *testTool) RequiresApproval() bool {
	return tt.requiresApproval
}
