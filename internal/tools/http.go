// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
)

// HTTPTool makes HTTP requests with configurable method, headers, body, and timeout.
type HTTPTool struct {
	DefaultTimeout int // From config, or 30
}

// httpArgs represents the arguments for the http tool.
type httpArgs struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

// Info returns the tool definition for the LLM.
func (t *HTTPTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "http",
		Description: "Make an HTTP request. Returns status code, response headers, and body (truncated to 10KB).",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "Request URL",
				},
				"method": map[string]any{
					"type":        "string",
					"description": "HTTP method (default: GET)",
				},
				"headers": map[string]any{
					"type":        "object",
					"description": "Request headers as key-value pairs",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Request body",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "Timeout in seconds (default: 30)",
				},
			},
		},
		Required: []string{"url"},
	}
}

// Run executes the HTTP request.
func (t *HTTPTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args httpArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	// Set defaults
	method := strings.ToUpper(args.Method)
	if method == "" {
		method = "GET"
	}

	// Determine timeout
	timeout := 30 * time.Second
	if t.DefaultTimeout > 0 {
		timeout = time.Duration(t.DefaultTimeout) * time.Second
	}
	if args.Timeout > 0 {
		timeout = time.Duration(args.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build request body
	var bodyReader io.Reader
	if args.Body != "" {
		bodyReader = strings.NewReader(args.Body)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, args.URL, bodyReader)
	if err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid request: %v", err)), nil
	}

	// Add headers
	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}

	// Execute request
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return agent.TransientError(fmt.Sprintf("request failed: %v", err)), nil
	}
	defer resp.Body.Close()

	// Read body (limited to 10KB)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10240))
	if err != nil {
		return agent.TransientError(fmt.Sprintf("error reading body: %v", err)), nil
	}

	// Format response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Status: %d %s\n\nHeaders:\n", resp.StatusCode, resp.Status)
	for k, vals := range resp.Header {
		for _, v := range vals {
			fmt.Fprintf(&sb, "  %s: %s\n", k, v)
		}
	}
	fmt.Fprintf(&sb, "\nBody:\n%s", string(body))

	return agent.ToolResult{Content: sb.String()}, nil
}

// RequiresApproval returns true because HTTP requests can have security implications.
func (t *HTTPTool) RequiresApproval() bool { return true }

// IsSafeArgs returns true for GET requests to localhost/127.0.0.1.
func (t *HTTPTool) IsSafeArgs(argsJSON string) bool {
	var args httpArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}

	// Must be GET
	method := strings.ToUpper(args.Method)
	if method == "" {
		method = "GET"
	}
	if method != "GET" {
		return false
	}

	// Must be localhost
	parsed, err := url.Parse(args.URL)
	if err != nil {
		return false
	}

	host := parsed.Hostname()
	return host == "localhost" || host == "127.0.0.1"
}
