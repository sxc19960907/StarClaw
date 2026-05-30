package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ollamaTestHandler returns an http.HandlerFunc that simulates an Ollama
// OpenAI-compatible chat completions endpoint.
func ollamaTestHandler(t *testing.T, statusCode int, responseBody string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/chat/completions") {
			t.Errorf("expected /v1/chat/completions, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(responseBody))
	}
}

func TestNewOllamaClient(t *testing.T) {
	t.Run("custom endpoint and model", func(t *testing.T) {
		c := NewOllamaClient("http://10.0.0.1:11434", "llama3.2")
		if c == nil {
			t.Fatal("NewOllamaClient returned nil")
		}
		if c.endpoint != "http://10.0.0.1:11434" {
			t.Errorf("endpoint = %q, want %q", c.endpoint, "http://10.0.0.1:11434")
		}
		if c.model != "llama3.2" {
			t.Errorf("model = %q, want %q", c.model, "llama3.2")
		}
	})

	t.Run("defaults", func(t *testing.T) {
		c := NewOllamaClient("", "")
		if c.endpoint != "http://localhost:11434" {
			t.Errorf("endpoint = %q, want %q", c.endpoint, "http://localhost:11434")
		}
		if c.model != "llama3.1" {
			t.Errorf("model = %q, want %q", c.model, "llama3.1")
		}
	})
}

func TestOllamaClient_Chat_TextResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "chatcmpl-123",
			"model": "llama3.1",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you?"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 5,
				"total_tokens": 15
			}
		}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	resp, err := client.Chat(context.Background(), "You are helpful.", []Message{{Role: "user", Content: "Say hello"}}, nil, 100, nil)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello! How can I help you?")
	}
	if resp.Usage.InputTokens != 10 {
		t.Errorf("InputTokens = %d, want 10", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 5 {
		t.Errorf("OutputTokens = %d, want 5", resp.Usage.OutputTokens)
	}
	if resp.StopReason != "stop" {
		t.Errorf("StopReason = %q, want %q", resp.StopReason, "stop")
	}
}

func TestOllamaClient_Chat_WithTools(t *testing.T) {
	server := httptest.NewServer(ollamaTestHandler(t, http.StatusOK, `{
		"id": "chatcmpl-456",
		"model": "llama3.1",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "",
				"tool_calls": [{
					"id": "call_abc",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"location\": \"Tokyo\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}],
		"usage": {
			"prompt_tokens": 20,
			"completion_tokens": 10,
			"total_tokens": 30
		}
	}`))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	tools := []ToolDef{
		{
			Name:        "get_weather",
			Description: "Get weather for a location",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{"type": "string"},
				},
			},
		},
	}

	resp, err := client.Chat(context.Background(), "", []Message{{Role: "user", Content: "Weather in Tokyo"}}, tools, 100, nil)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if len(resp.ToolUses) != 1 {
		t.Fatalf("expected 1 tool use, got %d", len(resp.ToolUses))
	}
	if resp.ToolUses[0].Name != "get_weather" {
		t.Errorf("ToolUse name = %q, want %q", resp.ToolUses[0].Name, "get_weather")
	}
	if resp.StopReason != "tool_calls" {
		t.Errorf("StopReason = %q, want %q", resp.StopReason, "tool_calls")
	}
}

func TestOllamaClient_Chat_APIError(t *testing.T) {
	server := httptest.NewServer(ollamaTestHandler(t, http.StatusBadRequest, `{
		"error": {
			"message": "model not found",
			"type": "invalid_request_error"
		}
	}`))
	defer server.Close()

	client := NewOllamaClient(server.URL, "nonexistent-model")
	_, err := client.Chat(context.Background(), "", []Message{{Role: "user", Content: "hi"}}, nil, 100, nil)
	if err == nil {
		t.Fatal("expected error for bad request")
	}
	if !strings.Contains(err.Error(), "model not found") {
		t.Errorf("error message = %q, want it to contain %q", err.Error(), "model not found")
	}
}

func TestOllamaClient_Chat_ServerError(t *testing.T) {
	server := httptest.NewServer(ollamaTestHandler(t, http.StatusInternalServerError, `internal server error`))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	_, err := client.Chat(context.Background(), "", []Message{{Role: "user", Content: "hi"}}, nil, 100, nil)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestOllamaClient_Chat_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(ollamaTestHandler(t, http.StatusOK, `{
		"id": "chatcmpl-789",
		"model": "test-model",
		"choices": [],
		"usage": {}
	}`))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	_, err := client.Chat(context.Background(), "", []Message{{Role: "user", Content: "hi"}}, nil, 100, nil)
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestOllamaClient_Chat_WithSystemPrompt(t *testing.T) {
	var receivedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"x","model":"m","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{}}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	_, err := client.Chat(context.Background(), "You are a cat.", []Message{{Role: "user", Content: "meow"}}, nil, 100, nil)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	msgs, ok := receivedBody["messages"].([]any)
	if !ok || len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	sysMsg := msgs[0].(map[string]any)
	if sysMsg["role"] != "system" || sysMsg["content"] != "You are a cat." {
		t.Errorf("system message = %v, want role=system content=You are a cat.", sysMsg)
	}
	userMsg := msgs[1].(map[string]any)
	if userMsg["role"] != "user" || userMsg["content"] != "meow" {
		t.Errorf("user message = %v, want role=user content=meow", userMsg)
	}
}

func TestOllamaClient_Chat_WithSpecificModel(t *testing.T) {
	var receivedModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		receivedModel, _ = body["model"].(string)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"x","model":"custom-v2","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{}}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "default-model")
	_, err := client.Chat(context.Background(), "", []Message{{Role: "user", Content: "hi"}}, nil, 100, &ChatOptions{SpecificModel: "custom-v2"})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if receivedModel != "custom-v2" {
		t.Errorf("model in request = %q, want %q", receivedModel, "custom-v2")
	}
}

func TestOllamaClient_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "chatcmpl-999",
			"model": "test-model",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Response from Complete"
				},
				"finish_reason": "stop"
			}],
			"usage": {}
		}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	resp, err := client.Complete(context.Background(), CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "Hello"}},
		MaxTokens: 200,
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.OutputText != "Response from Complete" {
		t.Errorf("OutputText = %q, want %q", resp.OutputText, "Response from Complete")
	}
}

func TestOllamaClient_SetModel(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "llama3.1")
	client.SetModel("mistral")
	if client.model != "mistral" {
		t.Errorf("model = %q, want %q", client.model, "mistral")
	}
}

func TestOllamaClient_Chat_ToolCallWithJSONArgs(t *testing.T) {
	server := httptest.NewServer(ollamaTestHandler(t, http.StatusOK, `{
		"id": "chatcmpl-args",
		"model": "test-model",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "",
				"tool_calls": [{
					"id": "call_1",
					"type": "function",
					"function": {
						"name": "search",
						"arguments": "{\"query\": \"hello\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}],
		"usage": {}
	}`))
	defer server.Close()

	client := NewOllamaClient(server.URL, "test-model")
	resp, err := client.Chat(context.Background(), "", []Message{{Role: "user", Content: "search"}}, []ToolDef{
		{Name: "search", Description: "Search", InputSchema: map[string]any{"type": "object"}},
	}, 100, nil)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if len(resp.ToolUses) != 1 {
		t.Fatalf("expected 1 tool use, got %d", len(resp.ToolUses))
	}
	var args map[string]any
	if err := json.Unmarshal(resp.ToolUses[0].Input, &args); err != nil {
		t.Fatalf("failed to unmarshal tool arguments: %v", err)
	}
	if args["query"] != "hello" {
		t.Errorf("arguments query = %v, want %q", args["query"], "hello")
	}
}
