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

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolUse represents a tool use request from the model
type ToolUse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ThinkingConfig for Anthropic extended thinking.
type ThinkingConfig struct {
	Type         string `json:"type"`                    // "adaptive", "enabled", or "disabled"
	BudgetTokens int    `json:"budget_tokens,omitempty"` // thinking token budget
}

// ChatOptions holds optional fields for LLM Chat requests.
type ChatOptions struct {
	Thinking        *ThinkingConfig `json:"thinking,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	SpecificModel   string          `json:"specific_model,omitempty"`
}

// StreamDelta represents an incremental text chunk from streaming.
type StreamDelta struct {
	Text string
}

// ToolDef defines a tool for the model
// FunctionDef describes a tool function schema.
type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Tool represents a complete tool definition (function or native).
type Tool struct {
	Type            string      `json:"type"`
	Name            string      `json:"name,omitempty"`
	Function        FunctionDef `json:"function,omitempty"`
	DisplayWidthPx  int         `json:"display_width_px,omitempty"`
	DisplayHeightPx int         `json:"display_height_px,omitempty"`
}

type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// Response represents a complete response from the model
type Response struct {
	Content    string
	ToolUses   []ToolUse
	Usage      Usage
	StopReason string
}

// LLMClient defines the interface for LLM API clients.
type LLMClient interface {
	Chat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions) (*Response, error)
}

// AnthropicClient implements the Anthropic Messages API.
type AnthropicClient struct {
	mu       sync.Mutex
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
}

// NewAnthropicClient creates a new Anthropic client.
func NewAnthropicClient(apiKey, endpoint, model string) *AnthropicClient {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-4-sonnet-20250514"
	}

	return &AnthropicClient{
		apiKey:   apiKey,
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

// SetModel sets the model to use
func (c *AnthropicClient) SetModel(model string) {
	c.mu.Lock()
	c.model = model
	c.mu.Unlock()
}

// Complete performs a non-chat completion. Implements the context.Completer interface.
func (c *AnthropicClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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

// Chat sends a chat request to the Anthropic Messages API and returns the response.
func (c *AnthropicClient) Chat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions) (*Response, error) {
	if maxTokens == 0 {
		maxTokens = 8192
	}

	reqBody := c.buildMessagesRequestBody(systemPrompt, messages, tools, maxTokens, opts, false)
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	reqURL := c.endpoint + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		// Try to parse error
		if msg := parseAnthropicErrorMessage(body); msg != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseResponse(result)
}

// StreamChat sends a streaming request to the Anthropic Messages API.
func (c *AnthropicClient) StreamChat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions, onDelta func(delta string)) (*Response, error) {
	if maxTokens == 0 {
		maxTokens = 8192
	}

	reqBody := c.buildMessagesRequestBody(systemPrompt, messages, tools, maxTokens, opts, true)
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := c.endpoint + "/v1/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if msg := parseAnthropicErrorMessage(body); msg != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return ParseAnthropicStream(resp.Body, onDelta)
}

func (c *AnthropicClient) buildMessagesRequestBody(systemPrompt string, messages []Message, tools []ToolDef, maxTokens int, opts *ChatOptions, stream bool) map[string]any {
	c.mu.Lock()
	model := c.model
	c.mu.Unlock()
	if opts != nil && opts.SpecificModel != "" {
		model = opts.SpecificModel
	}

	reqBody := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"messages":   messages,
	}
	if stream {
		reqBody["stream"] = true
	}

	if systemPrompt != "" {
		reqBody["system"] = systemPrompt
	}

	if len(tools) > 0 {
		anthropicTools := make([]map[string]any, len(tools))
		for i, tool := range tools {
			anthropicTools[i] = map[string]any{
				"name":         tool.Name,
				"description":  tool.Description,
				"input_schema": tool.InputSchema,
			}
		}
		reqBody["tools"] = anthropicTools
	}

	if opts != nil && opts.Thinking != nil {
		reqBody["thinking"] = opts.Thinking
	}

	if opts != nil && opts.ReasoningEffort != "" {
		reqBody["reasoning_effort"] = opts.ReasoningEffort
	}

	return reqBody
}

// parseResponse parses the API response
func parseResponse(result map[string]any) (*Response, error) {
	resp := &Response{}

	// Parse content blocks
	if content, ok := result["content"].([]any); ok {
		for _, block := range content {
			if blockMap, ok := block.(map[string]any); ok {
				blockType, _ := blockMap["type"].(string)
				switch blockType {
				case "text":
					if text, ok := blockMap["text"].(string); ok {
						resp.Content += text
					}
				case "tool_use":
					toolUse := ToolUse{
						ID:   getString(blockMap, "id"),
						Name: getString(blockMap, "name"),
					}
					if input, ok := blockMap["input"]; ok {
						toolUse.Input, _ = json.Marshal(input)
					}
					resp.ToolUses = append(resp.ToolUses, toolUse)
				}
			}
		}
	}

	// Parse usage
	if usage, ok := result["usage"].(map[string]any); ok {
		resp.Usage.InputTokens = getInt(usage, "input_tokens")
		resp.Usage.OutputTokens = getInt(usage, "output_tokens")
	}

	// Parse stop reason
	resp.StopReason = getString(result, "stop_reason")

	return resp, nil
}

// Helper functions
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func parseAnthropicErrorMessage(body []byte) string {
	var errorResult map[string]any
	if err := json.Unmarshal(body, &errorResult); err != nil {
		return ""
	}
	errObj, ok := errorResult["error"].(map[string]any)
	if !ok {
		return ""
	}
	msg, _ := errObj["message"].(string)
	return msg
}

// CompletionRequest is a generic LLM completion request (non-chat).
type CompletionRequest struct {
	Messages        []Message       `json:"messages"`
	ModelTier       string          `json:"model_tier,omitempty"`
	Temperature     float64         `json:"temperature,omitempty"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
	Thinking        *ThinkingConfig `json:"thinking,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	SpecificModel   string          `json:"specific_model,omitempty"`
}

// CompletionResponse is a generic LLM completion response.
type CompletionResponse struct {
	OutputText string `json:"output_text"`
}

// NewTextContent creates a Message with text content (for non-chat completions).
func NewTextContent(text string) string {
	return text
}

func getInt(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}
