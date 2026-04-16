package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

// MockLLMClient is a mock LLM client for testing
type MockLLMClient struct{}

func (m *MockLLMClient) Chat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int) (*client.Response, error) {
	return &client.Response{Content: "Mock response"}, nil
}

func TestNewModel(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	if model == nil {
		t.Fatal("NewModel() returned nil")
	}

	if model.loop != loop {
		t.Error("Model should store the agent loop")
	}

	if model.state != StateIdle {
		t.Errorf("Initial state should be StateIdle, got %d", model.state)
	}

	if len(model.messages) != 0 {
		t.Error("Initial messages should be empty")
	}
}

func TestModel_Init(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	// Test window resize
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m := newModel.(*Model)

	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}

	if m.height != 50 {
		t.Errorf("Expected height 50, got %d", m.height)
	}
}

func TestModel_Update_ClearScreen(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	// Add some messages
	model.messages = []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}

	// Clear screen with Ctrl+L
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	m := newModel.(*Model)

	if len(m.messages) != 0 {
		t.Errorf("Expected empty messages after clear, got %d", len(m.messages))
	}
}

func TestModel_Update_Quit(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	// Test quit with Ctrl+Q
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})

	if cmd == nil {
		t.Error("Expected quit command")
	}
}

func TestModel_Update_QuitCtrlC(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	// Test quit with Ctrl+C
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("Expected quit command")
	}
}

func TestModel_renderMessage(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	tests := []struct {
		name     string
		msg      Message
		contains string
	}{
		{
			name:     "user message",
			msg:      Message{Role: "user", Content: "Hello"},
			contains: "You:",
		},
		{
			name:     "assistant message",
			msg:      Message{Role: "assistant", Content: "Hi there"},
			contains: "Assistant:",
		},
		{
			name:     "system message",
			msg:      Message{Role: "system", Content: "System info"},
			contains: "System info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.renderMessage(tt.msg)
			if result == "" {
				t.Error("renderMessage should return non-empty string")
			}
		})
	}
}

func TestModel_renderToolCall(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	tool := &ToolCallInfo{
		Name:   "file_read",
		Args:   `{"path":"test.txt"}`,
		Result: "content",
		Error:  false,
	}

	result := model.renderToolCall(tool)
	if result == "" {
		t.Error("renderToolCall should return non-empty string")
	}

	// Test with error
	tool.Error = true
	result = model.renderToolCall(tool)
	if result == "" {
		t.Error("renderToolCall with error should return non-empty string")
	}
}

func TestModel_renderApprovalDialog(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)

	model.pendingTool = &ToolCallInfo{
		Name: "bash",
		Args: `{"command":"ls"}`,
	}

	result := model.renderApprovalDialog()
	if result == "" {
		t.Error("renderApprovalDialog should return non-empty string")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"exactly ten", 11, "exactly ten"},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestTUIEventHandler_OnToolCall(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)
	handler := &TUIEventHandler{model: model}

	handler.OnToolCall("file_read", `{"path":"test.txt"}`)

	if model.pendingTool == nil {
		t.Error("OnToolCall should set pendingTool")
	}

	if model.pendingTool.Name != "file_read" {
		t.Errorf("Expected tool name 'file_read', got '%s'", model.pendingTool.Name)
	}

	if model.state != StateAwaitingApproval {
		t.Errorf("Expected state StateAwaitingApproval, got %d", model.state)
	}
}

func TestTUIEventHandler_OnToolResult(t *testing.T) {
	loop := agent.NewAgentLoop(&MockLLMClient{}, agent.NewToolRegistry())
	model := NewModel(loop)
	model.pendingTool = &ToolCallInfo{Name: "file_read"}

	handler := &TUIEventHandler{model: model}
	handler.OnToolResult("file_read", agent.ToolResult{Content: "result", IsError: false})

	if model.pendingTool.Result != "result" {
		t.Errorf("Expected result 'result', got '%s'", model.pendingTool.Result)
	}
}

func TestStateConstants(t *testing.T) {
	if StateIdle != 0 {
		t.Error("StateIdle should be 0")
	}
	if StateThinking != 1 {
		t.Error("StateThinking should be 1")
	}
	if StateAwaitingApproval != 2 {
		t.Error("StateAwaitingApproval should be 2")
	}
	if StateStreaming != 3 {
		t.Error("StateStreaming should be 3")
	}
}
