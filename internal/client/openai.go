package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OpenAIClient implements the LLMClient interface for OpenAI-compatible APIs.
type OpenAIClient struct {
	mu                sync.Mutex
	apiKey            string
	baseURL           string
	model             string
	streamIdleTimeout time.Duration
	httpClient        *http.Client
}

// NewOpenAIClient creates a new OpenAI client.
// baseURL defaults to "https://api.openai.com".
// model defaults to "gpt-4o".
func NewOpenAIClient(apiKey, baseURL, model string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// Chat sends a chat request to the OpenAI Chat Completions API and returns the response.
func (c *OpenAIClient) Chat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions) (*Response, error) {
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

	// Wire reasoning effort if provided (supported by o-series models)
	if opts != nil && opts.ReasoningEffort != "" {
		reqBody["reasoning_effort"] = opts.ReasoningEffort
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	reqURL := openAIChatCompletionsURL(c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
					return nil, fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, msg)
				}
			}
		}
		return nil, fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseOpenAIResponse(result)
}

// parseOpenAIResponse parses an OpenAI Chat Completions API response into StarClaw's Response format.
func parseOpenAIResponse(result map[string]any) (*Response, error) {
	resp := &Response{}

	// Parse choices
	choices, ok := result["choices"].([]any)
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice, ok := choices[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected choice type in response")
	}

	// Parse message
	msg, ok := choice["message"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing message in choice")
	}

	// Parse content
	if content, ok := msg["content"].(string); ok {
		resp.Content = content
	}

	// Parse tool calls
	if toolCalls, ok := msg["tool_calls"].([]any); ok {
		for _, tc := range toolCalls {
			tcMap, ok := tc.(map[string]any)
			if !ok {
				continue
			}

			toolUse := ToolUse{
				ID: getString(tcMap, "id"),
			}

			if fn, ok := tcMap["function"].(map[string]any); ok {
				toolUse.Name = getString(fn, "name")
				if args, ok := fn["arguments"].(string); ok {
					toolUse.Input = json.RawMessage(args)
				}
			}

			resp.ToolUses = append(resp.ToolUses, toolUse)
		}
	}

	// Parse usage (OpenAI uses prompt_tokens / completion_tokens)
	if usage, ok := result["usage"].(map[string]any); ok {
		resp.Usage.InputTokens = getInt(usage, "prompt_tokens")
		resp.Usage.OutputTokens = getInt(usage, "completion_tokens")
	}

	// Parse stop reason (OpenAI uses finish_reason)
	resp.StopReason = getString(choice, "finish_reason")

	return resp, nil
}

// Complete performs a non-chat completion. Implements the context.Completer interface.
func (c *OpenAIClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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
func (c *OpenAIClient) SetModel(model string) {
	c.mu.Lock()
	c.model = model
	c.mu.Unlock()
}

// SetStreamIdleTimeout configures the per-line watchdog for streaming
// responses. A zero duration disables the watchdog.
func (c *OpenAIClient) SetStreamIdleTimeout(timeout time.Duration) {
	c.mu.Lock()
	c.streamIdleTimeout = timeout
	c.mu.Unlock()
}

func (c *OpenAIClient) StreamIdleTimeout() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.streamIdleTimeout
}

// StreamChat implements StreamingLLMClient for OpenAI-compatible APIs.
func (c *OpenAIClient) StreamChat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions, onDelta func(delta string)) (*Response, error) {
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
		"stream_options": map[string]any{
			"include_usage": true,
		},
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

	if opts != nil && opts.ReasoningEffort != "" {
		reqBody["reasoning_effort"] = opts.ReasoningEffort
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := openAIChatCompletionsURL(c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
		return nil, fmt.Errorf("openAI API error (%d): %s", resp.StatusCode, string(body))
	}

	return ParseOpenAIStreamWithOptions(ctx, resp.Body, onDelta, StreamParseOptions{IdleTimeout: c.StreamIdleTimeout()})
}

func openAIChatCompletionsURL(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}
