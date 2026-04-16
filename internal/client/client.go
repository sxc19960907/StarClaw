package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// ToolDef defines a tool for the model
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

// LLMClient provides a simplified interface for LLM APIs
type LLMClient struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
}

// NewLLMClient creates a new LLM client
func NewLLMClient(apiKey, endpoint, model string) *LLMClient {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-4-sonnet-20250514"
	}

	return &LLMClient{
		apiKey:   apiKey,
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

// SetModel sets the model to use
func (c *LLMClient) SetModel(model string) {
	c.model = model
}

// Chat sends a chat request and returns the response
func (c *LLMClient) Chat(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDef, maxTokens int) (*Response, error) {
	if maxTokens == 0 {
		maxTokens = 8192
	}

	// Build request body
	reqBody := map[string]any{
		"model":      c.model,
		"max_tokens": maxTokens,
		"messages":   messages,
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
		var errorResult map[string]any
		if err := json.Unmarshal(body, &errorResult); err == nil {
			if errObj, ok := errorResult["error"].(map[string]any); ok {
				if msg, ok := errObj["message"].(string); ok {
					return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
				}
			}
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
