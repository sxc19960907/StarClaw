package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTool_Info(t *testing.T) {
	tool := &HTTPTool{}
	info := tool.Info()

	assert.Equal(t, "http", info.Name)
	assert.NotEmpty(t, info.Description)
	assert.Contains(t, info.Description, "HTTP")
	assert.Equal(t, []string{"url"}, info.Required)
}

func TestHTTPTool_RequiresApproval(t *testing.T) {
	tool := &HTTPTool{}
	assert.True(t, tool.RequiresApproval())
}

func TestHTTPTool_Run_GET(t *testing.T) {
	// Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), fmt.Sprintf(`{"url": "%s"}`, server.URL))

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "200")
	assert.Contains(t, result.Content, `"status": "ok"`)
	assert.Contains(t, result.Content, "Content-Type: application/json")
}

func TestHTTPTool_Run_POSTWithBody(t *testing.T) {
	// Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Read body
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		assert.Equal(t, `{"name": "test"}`, string(body))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), fmt.Sprintf(`{
		"url": "%s",
		"method": "POST",
		"body": "{\"name\": \"test\"}"
	}`, server.URL))

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "201")
}

func TestHTTPTool_Run_CustomHeaders(t *testing.T) {
	// Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		w.Write([]byte(`"ok"`))
	}))
	defer server.Close()

	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), fmt.Sprintf(`{
		"url": "%s",
		"headers": {
			"Authorization": "Bearer token123",
			"Content-Type": "application/json"
		}
	}`, server.URL))

	require.NoError(t, err)
	assert.False(t, result.IsError)
}

func TestHTTPTool_Run_PUT(t *testing.T) {
	// Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(200)
		w.Write([]byte(`"updated"`))
	}))
	defer server.Close()

	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), fmt.Sprintf(`{
		"url": "%s",
		"method": "PUT"
	}`, server.URL))

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "200")
}

func TestHTTPTool_Run_DELETE(t *testing.T) {
	// Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	}))
	defer server.Close()

	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), fmt.Sprintf(`{
		"url": "%s",
		"method": "DELETE"
	}`, server.URL))

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "204")
}

func TestHTTPTool_Run_Timeout(t *testing.T) {
	// Start slow server that will definitely timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer server.Close()

	tool := &HTTPTool{}
	// 1 second timeout should fail against 2 second delay
	result, err := tool.Run(context.Background(), fmt.Sprintf(`{
		"url": "%s",
		"timeout": 1
	}`, server.URL))

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "request failed")
}

func TestHTTPTool_Run_InvalidURL(t *testing.T) {
	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), `{"url": "not-a-valid-url"}`)

	require.NoError(t, err)
	// This returns a transient error (request failed) rather than validation error
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "failed")
}

func TestHTTPTool_Run_InvalidJSON(t *testing.T) {
	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), "not valid json")

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestHTTPTool_IsSafeArgs_LocalhostGET(t *testing.T) {
	tool := &HTTPTool{}

	// Safe: GET to localhost
	assert.True(t, tool.IsSafeArgs(`{"url": "http://localhost:8080/api"}`))
	assert.True(t, tool.IsSafeArgs(`{"url": "http://localhost:8080/api", "method": "GET"}`))
	assert.True(t, tool.IsSafeArgs(`{"url": "http://127.0.0.1:3000"}`))
	assert.True(t, tool.IsSafeArgs(`{"url": "http://127.0.0.1:3000/path"}`))
}

func TestHTTPTool_IsSafeArgs_NotSafe(t *testing.T) {
	tool := &HTTPTool{}

	// Not safe: POST to localhost
	assert.False(t, tool.IsSafeArgs(`{"url": "http://localhost:8080/api", "method": "POST"}`))
	assert.False(t, tool.IsSafeArgs(`{"url": "http://localhost:8080/api", "method": "PUT"}`))
	assert.False(t, tool.IsSafeArgs(`{"url": "http://localhost:8080/api", "method": "DELETE"}`))

	// Not safe: GET to external
	assert.False(t, tool.IsSafeArgs(`{"url": "http://example.com", "method": "GET"}`))
	assert.False(t, tool.IsSafeArgs(`{"url": "https://api.github.com", "method": "GET"}`))
}

func TestHTTPTool_IsSafeArgs_InvalidInput(t *testing.T) {
	tool := &HTTPTool{}

	// Invalid JSON
	assert.False(t, tool.IsSafeArgs("not json"))

	// Invalid URL
	assert.False(t, tool.IsSafeArgs(`{"url": "://invalid-url"}`))

	// Missing URL
	assert.False(t, tool.IsSafeArgs(`{}`))
}

func TestHTTPTool_Run_ResponseTruncation(t *testing.T) {
	// Start server that returns large response
	largeResponse := make([]byte, 20000)
	for i := range largeResponse {
		largeResponse[i] = 'x'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(largeResponse)
	}))
	defer server.Close()

	tool := &HTTPTool{}
	result, err := tool.Run(context.Background(), fmt.Sprintf(`{"url": "%s"}`, server.URL))

	require.NoError(t, err)
	assert.False(t, result.IsError)
	// Response should be truncated to 10KB
	assert.Less(t, len(result.Content), 15000) // Should be around 10KB + headers
}
