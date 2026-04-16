package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

// MockEventHandler for testing
type MockEventHandler struct {
	toolCalls  []string
	toolResults []string
	texts      []string
	usages     []client.Usage
}

func (m *MockEventHandler) OnToolCall(name string, args string) {
	m.toolCalls = append(m.toolCalls, name)
}

func (m *MockEventHandler) OnToolResult(name string, result ToolResult) {
	m.toolResults = append(m.toolResults, result.Content)
}

func (m *MockEventHandler) OnText(text string) {
	m.texts = append(m.texts, text)
}

func (m *MockEventHandler) OnUsage(usage client.Usage) {
	m.usages = append(m.usages, usage)
}

func TestNewAgentLoop(t *testing.T) {
	llmClient := client.NewLLMClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	if loop == nil {
		t.Fatal("NewAgentLoop returned nil")
	}
	if loop.maxIter != 25 {
		t.Errorf("Expected default maxIter 25, got %d", loop.maxIter)
	}
	if loop.maxTokens != 8192 {
		t.Errorf("Expected default maxTokens 8192, got %d", loop.maxTokens)
	}
}

func TestAgentLoop_Setters(t *testing.T) {
	llmClient := client.NewLLMClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	loop.SetMaxIterations(10)
	if loop.maxIter != 10 {
		t.Errorf("Expected maxIter 10, got %d", loop.maxIter)
	}

	loop.SetMaxTokens(4096)
	if loop.maxTokens != 4096 {
		t.Errorf("Expected maxTokens 4096, got %d", loop.maxTokens)
	}

	loop.SetResultTruncation(10000)
	if loop.resultTrunc != 10000 {
		t.Errorf("Expected resultTrunc 10000, got %d", loop.resultTrunc)
	}

	handler := &MockEventHandler{}
	loop.SetEventHandler(handler)
	if loop.handler == nil {
		t.Error("Event handler should be set")
	}

	loop.SetSystemPrompt("You are a test assistant")
	if loop.systemPrompt != "You are a test assistant" {
		t.Errorf("Expected system prompt set")
	}
}

func TestAgentLoop_buildTools(t *testing.T) {
	llmClient := client.NewLLMClient("test", "", "")
	registry := NewToolRegistry()

	// Register a mock tool
	registry.Register(&MockTool{
		name:        "test_tool",
		description: "A test tool",
	})

	loop := NewAgentLoop(llmClient, registry)
	tools := loop.buildTools()

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", tools[0].Name)
	}
}

func TestAgentLoop_buildToolResultContent(t *testing.T) {
	llmClient := client.NewLLMClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	toolUse := client.ToolUse{
		ID:   "toolu_123",
		Name: "test_tool",
	}

	result := ToolResult{
		Content: "success result",
		IsError: false,
	}

	content := loop.buildToolResultContent(toolUse, result)
	if !strings.Contains(content, "tool_result") {
		t.Error("Content should contain 'tool_result'")
	}
	if !strings.Contains(content, "toolu_123") {
		t.Error("Content should contain tool use ID")
	}

	// Test error result
	result.IsError = true
	content = loop.buildToolResultContent(toolUse, result)
	if !strings.Contains(content, "is_error") {
		t.Error("Error result should contain is_error flag")
	}
}

func TestAgentLoop_executeTool(t *testing.T) {
	llmClient := client.NewLLMClient("test", "", "")
	registry := NewToolRegistry()

	// Register mock tool
	mockTool := &MockTool{
		name:    "mock_tool",
		execute: func() ToolResult { return ToolResult{Content: "mock result"} },
	}
	registry.Register(mockTool)

	loop := NewAgentLoop(llmClient, registry)
	handler := &MockEventHandler{}
	loop.SetEventHandler(handler)

	toolUse := client.ToolUse{
		ID:    "toolu_123",
		Name:  "mock_tool",
		Input: []byte(`{}`),
	}

	result := loop.executeTool(context.Background(), toolUse)

	if result.Content != "mock result" {
		t.Errorf("Expected 'mock result', got '%s'", result.Content)
	}

	if len(handler.toolCalls) != 1 {
		t.Error("Handler should have received tool call event")
	}
	if len(handler.toolResults) != 1 {
		t.Error("Handler should have received tool result event")
	}
}

func TestAgentLoop_executeTool_UnknownTool(t *testing.T) {
	llmClient := client.NewLLMClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	toolUse := client.ToolUse{
		ID:   "toolu_123",
		Name: "unknown_tool",
	}

	result := loop.executeTool(context.Background(), toolUse)

	if !result.IsError {
		t.Error("Unknown tool should return error")
	}
	if result.ErrorCategory != ErrCategoryValidation {
		t.Errorf("Expected validation error, got %s", result.ErrorCategory)
	}
}

func TestAgentLoop_TruncateResult(t *testing.T) {
	llmClient := client.NewLLMClient("test", "", "")
	registry := NewToolRegistry()

	// Register mock tool that returns large result
	largeContent := strings.Repeat("x", 40000)
	mockTool := &MockTool{
		name:    "large_tool",
		execute: func() ToolResult { return ToolResult{Content: largeContent} },
	}
	registry.Register(mockTool)

	loop := NewAgentLoop(llmClient, registry)
	loop.SetResultTruncation(1000)

	toolUse := client.ToolUse{
		ID:    "toolu_123",
		Name:  "large_tool",
		Input: []byte(`{}`),
	}

	result := loop.executeTool(context.Background(), toolUse)

	if len(result.Content) <= 1000 {
		t.Error("Result should be truncated")
	}
	if !strings.Contains(result.Content, "truncated") {
		t.Error("Truncated result should indicate truncation")
	}
}

// MockTool for testing
type MockTool struct {
	name            string
	description     string
	requiresApproval bool
	execute         func() ToolResult
}

func (m *MockTool) Info() ToolInfo {
	return ToolInfo{
		Name:        m.name,
		Description: m.description,
		Parameters:  map[string]any{},
		Required:    []string{},
	}
}

func (m *MockTool) Run(ctx context.Context, args string) (ToolResult, error) {
	if m.execute != nil {
		return m.execute(), nil
	}
	return ToolResult{Content: "mock result"}, nil
}

func (m *MockTool) RequiresApproval() bool {
	return m.requiresApproval
}
