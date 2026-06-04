package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GatewayClient wraps http.Client with a base URL and API key for API calls.
type GatewayClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewGatewayClient creates a new GatewayClient.
func NewGatewayClient(baseURL, apiKey string) *GatewayClient {
	return &GatewayClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HTTPClient exposes the underlying *http.Client for callers that need to
// configure timeouts or reuse the transport.
func (c *GatewayClient) HTTPClient() *http.Client {
	return c.httpClient
}

// Post sends a POST request to the given path with the given JSON body.
// Returns the response body bytes. Non-2xx status codes produce an *APIError.
func (c *GatewayClient) Post(ctx context.Context, path string, body []byte) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	return c.do(req)
}

// Get sends a GET request to the given path.
// Returns the response body bytes. Non-2xx status codes produce an *APIError.
func (c *GatewayClient) Get(ctx context.Context, path string) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create GET request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	return c.do(req)
}

// do executes an HTTP request and reads the response body.
func (c *GatewayClient) do(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gateway request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(data)}
	}

	return data, nil
}

// APIError represents an HTTP error with a status code and response body.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("API returned %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("API returned %d", e.StatusCode)
}
