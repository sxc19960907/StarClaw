package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CloudConfig holds configuration for cloud agent delegation.
type CloudConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Endpoint      string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
	APIKey        string `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	Timeout       int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	MaxConcurrent int    `mapstructure:"max_concurrent" yaml:"max_concurrent" json:"max_concurrent"`
}

// CloudDelegateRequest is the payload sent to the cloud delegation endpoint.
type CloudDelegateRequest struct {
	Task    string `json:"task"`
	Agent   string `json:"agent,omitempty"`
	Context string `json:"context,omitempty"`
}

// CloudDelegateResponse is the final result from a cloud delegation.
type CloudDelegateResponse struct {
	Result string `json:"result"`
	Usage  Usage  `json:"usage,omitempty"`
	Error  string `json:"error,omitempty"`
}

// CloudProgress represents a streaming progress event from cloud delegation.
type CloudProgress struct {
	Type string `json:"type"` // "thinking", "tool_call", "text", "done", "error"
	Data string `json:"data"`
}

// CloudClient communicates with a remote cloud agent endpoint.
type CloudClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

// NewCloudClient creates a new CloudClient.
func NewCloudClient(cfg CloudConfig) *CloudClient {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 3600 * time.Second
	}
	return &CloudClient{
		endpoint: cfg.Endpoint,
		apiKey:   cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Delegate sends a task to the cloud agent and blocks until completion.
func (c *CloudClient) Delegate(ctx context.Context, req CloudDelegateRequest) (*CloudDelegateResponse, error) {
	return c.DelegateStream(ctx, req, nil)
}

// DelegateStream sends a task to the cloud agent and streams progress events.
// onProgress is called for each intermediate event. Returns the final response.
func (c *CloudClient) DelegateStream(ctx context.Context, req CloudDelegateRequest, onProgress func(CloudProgress)) (*CloudDelegateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal delegate request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v1/delegate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create delegate request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cloud delegate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloud delegate error (%d): %s", resp.StatusCode, string(respBody))
	}

	return c.readDelegateStream(resp.Body, onProgress)
}

func (c *CloudClient) readDelegateStream(body io.Reader, onProgress func(CloudProgress)) (*CloudDelegateResponse, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var finalResult CloudDelegateResponse
	var resultText strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var progress CloudProgress
		if err := json.Unmarshal([]byte(data), &progress); err != nil {
			continue
		}

		switch progress.Type {
		case "done":
			finalResult.Result = resultText.String()
			if progress.Data != "" {
				finalResult.Result = progress.Data
			}
			return &finalResult, nil
		case "error":
			return nil, fmt.Errorf("cloud agent error: %s", progress.Data)
		case "text":
			resultText.WriteString(progress.Data)
		}

		if onProgress != nil {
			onProgress(progress)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cloud delegate stream read error: %w", err)
	}

	finalResult.Result = resultText.String()
	return &finalResult, nil
}
