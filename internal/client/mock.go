package client

import (
	"context"
	"encoding/json"
)

// MockClient is a mock LLM client for testing
type MockClient struct {
	response       string
	toolCallName   string
	toolCallArgs   string
	responseFunc   func(input string) *MockMessage
	lastMessages   []Message
	lastTools      []ToolDef
	callCount      int
}

// MockMessage is a simple message struct for mock responses
type MockMessage struct {
	Role      string
	Content   string
	ToolCalls []MockToolCall
}

// MockToolCall represents a tool call in a mock message
type MockToolCall struct {
	ID   string
	Name string
	Args string
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		response: "Mock response",
	}
}

// SetResponse sets a simple text response
func (m *MockClient) SetResponse(response string) {
	m.response = response
	m.toolCallName = ""
	m.toolCallArgs = ""
	m.responseFunc = nil
}

// SetToolCallResponse configures the mock to return a tool call
func (m *MockClient) SetToolCallResponse(toolName, toolArgs string) {
	m.toolCallName = toolName
	m.toolCallArgs = toolArgs
	m.response = ""
}

// SetHandler sets a custom response handler
func (m *MockClient) SetHandler(handler func(input string) *MockMessage) {
	m.responseFunc = handler
}

// Chat implements the LLMClient interface
func (m *MockClient) Chat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int) (*Response, error) {
	m.callCount++
	m.lastMessages = messages
	m.lastTools = tools

	// Use custom handler if set
	if m.responseFunc != nil {
		msg := m.responseFunc("")
		if msg == nil {
			return &Response{Content: ""}, nil
		}

		resp := &Response{
			Content: msg.Content,
		}

		// Convert ToolCalls to ToolUses
		for _, tc := range msg.ToolCalls {
			resp.ToolUses = append(resp.ToolUses, ToolUse{
				ID:    tc.ID,
				Name:  tc.Name,
				Input: json.RawMessage(tc.Args),
			})
		}

		return resp, nil
	}

	// Return tool call if configured
	if m.toolCallName != "" {
		return &Response{
			ToolUses: []ToolUse{
				{
					ID:    "mock_tool_call_1",
					Name:  m.toolCallName,
					Input: json.RawMessage(m.toolCallArgs),
				},
			},
		}, nil
	}

	// Return text response
	return &Response{
		Content: m.response,
		Usage: Usage{
			InputTokens:  10,
			OutputTokens: len(m.response),
		},
	}, nil
}

// GetCallCount returns the number of Chat calls
func (m *MockClient) GetCallCount() int {
	return m.callCount
}

// GetLastMessages returns the messages from the last call
func (m *MockClient) GetLastMessages() []Message {
	return m.lastMessages
}

// GetLastTools returns the tools from the last call
func (m *MockClient) GetLastTools() []ToolDef {
	return m.lastTools
}

// SetModel is a no-op for the mock client
func (m *MockClient) SetModel(model string) {}
