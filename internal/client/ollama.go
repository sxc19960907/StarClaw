package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// OllamaClient implements the LLMClient interface for local Ollama instances.
// It communicates via Ollama's OpenAI-compatible /v1/chat/completions endpoint.
type OllamaClient struct {
	mu                sync.Mutex
	endpoint          string
	model             string
	streamIdleTimeout time.Duration
	httpClient        *http.Client
}

// NewOllamaClient creates a new client for a local Ollama instance.
// endpoint is the base URL (e.g. "http://localhost:11434").
// model is the default model name (e.g. "llama3.1").
func NewOllamaClient(endpoint, model string) *OllamaClient {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.1"
	}
	return &OllamaClient{
		endpoint: endpoint,
		model:    model,
		httpClient: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
}

// Chat sends a chat request to the Ollama instance and returns the response.
func (c *OllamaClient) Chat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions) (*Response, error) {
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Determine model
	model := c.model
	if opts != nil && opts.SpecificModel != "" {
		model = opts.SpecificModel
	}

	// Convert messages to OpenAI format
	openAIMessages := make([]map[string]any, 0, len(messages)+1)

	// System prompt goes as a system role message (at the beginning)
	if systemPrompt != "" {
		openAIMessages = append(openAIMessages, map[string]any{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	// Convert StarClaw messages to OpenAI messages
	for _, msg := range messages {
		openAIMessages = append(openAIMessages, map[string]any{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Build request body
	reqBody := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"messages":   openAIMessages,
	}

	// Convert tools to OpenAI function calling format
	if len(tools) > 0 {
		openAITools := make([]map[string]any, len(tools))
		for i, tool := range tools {
			openAITools[i] = map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        tool.Name,
					"description": tool.Description,
					"parameters":  tool.InputSchema,
				},
			}
		}
		reqBody["tools"] = openAITools
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	reqURL := c.endpoint + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResult map[string]any
		if err := json.Unmarshal(body, &errorResult); err == nil {
			if errObj, ok := errorResult["error"].(map[string]any); ok {
				if msg, ok := errObj["message"].(string); ok {
					return nil, fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, msg)
				}
			}
		}
		return nil, fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseOpenAIResponse(result)
}

// Complete performs a non-chat completion. Uses the same Ollama endpoint.
func (c *OllamaClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1000
	}

	var systemPrompt string
	var userContent string
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			userContent = msg.Content
		}
	}

	opts := &ChatOptions{
		Thinking:        req.Thinking,
		ReasoningEffort: req.ReasoningEffort,
		SpecificModel:   req.SpecificModel,
	}

	resp, err := c.Chat(ctx, systemPrompt, []Message{{Role: "user", Content: userContent}}, nil, maxTokens, opts)
	if err != nil {
		return nil, err
	}
	return &CompletionResponse{OutputText: resp.Content}, nil
}

// SetModel sets the model to use for subsequent requests.
func (c *OllamaClient) SetModel(model string) {
	c.mu.Lock()
	c.model = model
	c.mu.Unlock()
}

// SetStreamIdleTimeout configures the per-line watchdog for streaming
// responses. A zero duration disables the watchdog.
func (c *OllamaClient) SetStreamIdleTimeout(timeout time.Duration) {
	c.mu.Lock()
	c.streamIdleTimeout = timeout
	c.mu.Unlock()
}

func (c *OllamaClient) StreamIdleTimeout() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.streamIdleTimeout
}

// StreamChat implements StreamingLLMClient for Ollama's OpenAI-compatible endpoint.
func (c *OllamaClient) StreamChat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions, onDelta func(delta string)) (*Response, error) {
	if maxTokens == 0 {
		maxTokens = 4096
	}

	model := c.model
	if opts != nil && opts.SpecificModel != "" {
		model = opts.SpecificModel
	}

	openAIMessages := make([]map[string]any, 0, len(messages)+1)
	if systemPrompt != "" {
		openAIMessages = append(openAIMessages, map[string]any{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	for _, msg := range messages {
		openAIMessages = append(openAIMessages, map[string]any{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	reqBody := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"messages":   openAIMessages,
		"stream":     true,
	}

	if len(tools) > 0 {
		openAITools := make([]map[string]any, len(tools))
		for i, tool := range tools {
			openAITools[i] = map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        tool.Name,
					"description": tool.Description,
					"parameters":  tool.InputSchema,
				},
			}
		}
		reqBody["tools"] = openAITools
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := c.endpoint + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (%d): %s", resp.StatusCode, string(body))
	}

	return ParseOpenAIStreamWithOptions(ctx, resp.Body, onDelta, StreamParseOptions{IdleTimeout: c.StreamIdleTimeout()})
}
