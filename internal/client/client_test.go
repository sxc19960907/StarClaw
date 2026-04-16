package client

import (
	"encoding/json"
	"testing"
)

func TestNewLLMClient(t *testing.T) {
	client := NewLLMClient("test-api-key", "", "")
	if client == nil {
		t.Fatal("NewLLMClient returned nil")
	}
	if client.apiKey != "test-api-key" {
		t.Error("Client should store API key")
	}
	if client.endpoint == "" {
		t.Error("Client should have default endpoint")
	}
	if client.model == "" {
		t.Error("Client should have default model")
	}
}

func TestSetModel(t *testing.T) {
	client := NewLLMClient("test", "", "")
	client.SetModel("claude-4-opus")
	if client.model != "claude-4-opus" {
		t.Errorf("Expected model 'claude-4-opus', got '%s'", client.model)
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello",
	}
	if msg.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", msg.Role)
	}
}

func TestToolUse(t *testing.T) {
	toolUse := ToolUse{
		ID:   "toolu_123",
		Name: "file_read",
		Input: json.RawMessage(`{"path": "/test.txt"}`),
	}
	if toolUse.Name != "file_read" {
		t.Errorf("Expected name 'file_read', got '%s'", toolUse.Name)
	}
}

func TestToolResult(t *testing.T) {
	result := ToolResult{
		ToolUseID: "toolu_123",
		Content:   "file contents",
		IsError:   false,
	}
	if result.ToolUseID != "toolu_123" {
		t.Errorf("Expected ToolUseID 'toolu_123', got '%s'", result.ToolUseID)
	}
}

func TestToolDef(t *testing.T) {
	toolDef := ToolDef{
		Name:        "file_read",
		Description: "Read a file",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
			},
		},
	}
	if toolDef.Name != "file_read" {
		t.Errorf("Expected name 'file_read', got '%s'", toolDef.Name)
	}
}

func TestUsage(t *testing.T) {
	usage := Usage{
		InputTokens:  100,
		OutputTokens: 50,
	}
	if usage.InputTokens != 100 {
		t.Errorf("Expected InputTokens 100, got %d", usage.InputTokens)
	}
	if usage.OutputTokens != 50 {
		t.Errorf("Expected OutputTokens 50, got %d", usage.OutputTokens)
	}
}

func TestResponse(t *testing.T) {
	resp := &Response{
		Content:    "Hello",
		ToolUses:   []ToolUse{{Name: "file_read"}},
		Usage:      Usage{InputTokens: 10, OutputTokens: 5},
		StopReason: "end_turn",
	}
	if resp.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", resp.Content)
	}
	if len(resp.ToolUses) != 1 {
		t.Errorf("Expected 1 tool use, got %d", len(resp.ToolUses))
	}
}

func TestGetString(t *testing.T) {
	m := map[string]any{"key": "value"}
	if getString(m, "key") != "value" {
		t.Error("getString should return value")
	}
	if getString(m, "missing") != "" {
		t.Error("getString should return empty for missing key")
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		input    any
		expected int
	}{
		{int(42), 42},
		{float64(42.5), 42},
		{int64(42), 42},
		{"string", 0},
	}

	for _, tt := range tests {
		m := map[string]any{"key": tt.input}
		if getInt(m, "key") != tt.expected {
			t.Errorf("getInt(%v) = %d, expected %d", tt.input, getInt(m, "key"), tt.expected)
		}
	}
}
