package agent

import (
	"testing"
)

func TestNewMockAgentLoop(t *testing.T) {
	loop := NewMockAgentLoop()
	if loop == nil {
		t.Fatal("NewMockAgentLoop returned nil")
	}
	if loop.registry == nil {
		t.Error("expected non-nil registry")
	}
	if loop.llmClient == nil {
		t.Error("expected non-nil llmClient")
	}
	if loop.maxIter != 25 {
		t.Errorf("expected maxIter 25, got %d", loop.maxIter)
	}
	if loop.maxTokens != 8192 {
		t.Errorf("expected maxTokens 8192, got %d", loop.maxTokens)
	}
}

func TestTestTool(t *testing.T) {
	tool := TestTool("my_tool")
	if tool == nil {
		t.Fatal("TestTool returned nil")
	}

	info := tool.Info()
	if info.Name != "my_tool" {
		t.Errorf("Name = %q; want %q", info.Name, "my_tool")
	}
	if info.Description != "test tool: my_tool" {
		t.Errorf("Description = %q; want %q", info.Description, "test tool: my_tool")
	}
	if tool.RequiresApproval() {
		t.Error("expected RequiresApproval() = false")
	}

	result, err := tool.Run(nil, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Content != "my_tool executed" {
		t.Errorf("Content = %q; want %q", result.Content, "my_tool executed")
	}
}

func TestTestTool_Multiple(t *testing.T) {
	t1 := TestTool("alpha")
	t2 := TestTool("beta")
	if t1.Info().Name == t2.Info().Name {
		t.Error("expected different tool names")
	}
}

func TestAssertNoError(t *testing.T) {
	// Should not fail
	AssertNoError(t, nil)
}

// errHelper wraps *testing.T so we can capture Fatal calls.
type errHelper struct {
	*testing.T
	lastFatal string
}

func (e *errHelper) Fatalf(format string, args ...any) {
	e.lastFatal = format
}

func TestAssertNoError_FatalfCalled(t *testing.T) {
	ht := &errHelper{T: t}
	AssertNoError(ht, assertError("oops"))
	if ht.lastFatal == "" {
		t.Error("expected Fatalf to be called")
	}
}

type assertError string

func (e assertError) Error() string { return string(e) }

func TestTestMessage(t *testing.T) {
	msg := TestMessage("user", "hello world")
	if msg.Role != "user" {
		t.Errorf("Role = %q; want %q", msg.Role, "user")
	}
	if msg.Content != "hello world" {
		t.Errorf("Content = %q; want %q", msg.Content, "hello world")
	}
}

func TestTestMessage_Assistant(t *testing.T) {
	msg := TestMessage("assistant", "I can help")
	if msg.Role != "assistant" {
		t.Errorf("Role = %q; want %q", msg.Role, "assistant")
	}
}

func TestNewMockAgentLoop_RegistersTools(t *testing.T) {
	loop := NewMockAgentLoop()
	// Register a tool and verify the loop's registry picks it up
	tool := TestTool("test_cmd")
	loop.registry.Register(tool)

	got, ok := loop.registry.Get("test_cmd")
	if !ok {
		t.Fatal("expected tool to be registered")
	}
	if got.Info().Name != "test_cmd" {
		t.Errorf("Name = %q", got.Info().Name)
	}
}

func TestTestTool_EmptyName(t *testing.T) {
	tool := TestTool("")
	if tool == nil {
		t.Fatal("TestTool('') returned nil")
	}
	if tool.Info().Name != "" {
		t.Errorf("expected empty name, got %q", tool.Info().Name)
	}
	_, err := tool.Run(nil, "")
	if err != nil {
		t.Errorf("unexpected error for empty-name tool: %v", err)
	}
}
