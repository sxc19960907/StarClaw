package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestNewAnthropicClient(t *testing.T) {
	client := NewAnthropicClient("test-api-key", "", "")
	if client == nil {
		t.Fatal("NewAnthropicClient returned nil")
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
	client := NewAnthropicClient("test", "", "")
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
		ID:    "toolu_123",
		Name:  "file_read",
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

func TestAnthropicClient_StreamChat(t *testing.T) {
	var gotRequest map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("path = %q, want /v1/messages", r.URL.Path)
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("Accept = %q, want text/event-stream", r.Header.Get("Accept"))
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Fatalf("x-api-key = %q, want test-key", r.Header.Get("x-api-key"))
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" Anthropic"}}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":3}}

event: message_stop
data: {"type":"message_stop"}

`)
	}))
	defer server.Close()

	client := NewAnthropicClient("test-key", server.URL, "default-model")
	var deltas []string
	resp, err := client.StreamChat(
		context.Background(),
		"system prompt",
		[]Message{{Role: "user", Content: "hello"}},
		[]ToolDef{{
			Name:        "file_read",
			Description: "Read files",
			InputSchema: map[string]any{
				"type": "object",
			},
		}},
		123,
		&ChatOptions{
			Thinking:        &ThinkingConfig{Type: "enabled", BudgetTokens: 1024},
			ReasoningEffort: "high",
			SpecificModel:   "specific-model",
		},
		func(delta string) {
			deltas = append(deltas, delta)
		},
	)
	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
	}
	if resp.Content != "Hello Anthropic" {
		t.Fatalf("content = %q, want Hello Anthropic", resp.Content)
	}
	if resp.StopReason != "end_turn" {
		t.Fatalf("stop reason = %q, want end_turn", resp.StopReason)
	}
	if resp.Usage.OutputTokens != 3 {
		t.Fatalf("output tokens = %d, want 3", resp.Usage.OutputTokens)
	}
	if strings.Join(deltas, "") != "Hello Anthropic" {
		t.Fatalf("deltas = %v, want Hello Anthropic", deltas)
	}
	if gotRequest["stream"] != true {
		t.Fatalf("stream = %v, want true", gotRequest["stream"])
	}
	if gotRequest["model"] != "specific-model" {
		t.Fatalf("model = %v, want specific-model", gotRequest["model"])
	}
	if gotRequest["system"] != "system prompt" {
		t.Fatalf("system = %v, want system prompt", gotRequest["system"])
	}
	if gotRequest["max_tokens"] != float64(123) {
		t.Fatalf("max_tokens = %v, want 123", gotRequest["max_tokens"])
	}
	if _, ok := gotRequest["tools"].([]any); !ok {
		t.Fatalf("tools missing from request: %#v", gotRequest["tools"])
	}
	thinking, ok := gotRequest["thinking"].(map[string]any)
	if !ok {
		t.Fatalf("thinking missing from request: %#v", gotRequest["thinking"])
	}
	if thinking["type"] != "enabled" || thinking["budget_tokens"] != float64(1024) {
		t.Fatalf("thinking = %#v, want enabled/1024", thinking)
	}
	if gotRequest["reasoning_effort"] != "high" {
		t.Fatalf("reasoning_effort = %v, want high", gotRequest["reasoning_effort"])
	}
}

func TestAnthropicClient_StreamChat_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":{"message":"bad stream request"}}`)
	}))
	defer server.Close()

	client := NewAnthropicClient("test-key", server.URL, "model")
	_, err := client.StreamChat(context.Background(), "", []Message{{Role: "user", Content: "hi"}}, nil, 0, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "API error (400): bad stream request" {
		t.Fatalf("error = %q, want parsed API error", got)
	}
}

func TestMockClient_DefensiveCopies(t *testing.T) {
	mock := NewMockClient()
	messages := []Message{{Role: "user", Content: "original"}}
	tools := []ToolDef{{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"path": map[string]any{"type": "string"}},
		},
	}}

	if _, err := mock.Chat(context.Background(), "", messages, tools, 0, nil); err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	messages[0].Content = "mutated"
	tools[0].InputSchema["type"] = "mutated"
	gotMessages := mock.GetLastMessages()
	gotTools := mock.GetLastTools()

	if gotMessages[0].Content != "original" {
		t.Fatalf("stored messages were mutated through caller slice: %q", gotMessages[0].Content)
	}
	if gotTools[0].InputSchema["type"] != "object" {
		t.Fatalf("stored tool schema was mutated through caller map: %v", gotTools[0].InputSchema["type"])
	}

	gotMessages[0].Content = "changed again"
	gotTools[0].InputSchema["type"] = "changed again"
	gotMessages = mock.GetLastMessages()
	gotTools = mock.GetLastTools()

	if gotMessages[0].Content != "original" {
		t.Fatalf("GetLastMessages exposed internal slice: %q", gotMessages[0].Content)
	}
	if gotTools[0].InputSchema["type"] != "object" {
		t.Fatalf("GetLastTools exposed internal schema: %v", gotTools[0].InputSchema["type"])
	}
}

func TestMockClient_ConcurrentUse(t *testing.T) {
	mock := NewMockClient()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mock.SetResponse("ok")
			if _, err := mock.Chat(context.Background(), "", []Message{{Role: "user", Content: "hi"}}, nil, 0, nil); err != nil {
				t.Errorf("Chat() error = %v", err)
			}
			_ = mock.GetCallCount()
			_ = mock.GetLastMessages()
			_ = mock.GetLastTools()
		}()
	}
	wg.Wait()
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
