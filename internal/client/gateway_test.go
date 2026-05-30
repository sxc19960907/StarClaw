package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGatewayClient(t *testing.T) {
	c := NewGatewayClient("http://localhost:8080", "test-key")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL 'http://localhost:8080', got %q", c.baseURL)
	}
	if c.apiKey != "test-key" {
		t.Errorf("expected apiKey 'test-key', got %q", c.apiKey)
	}
	if c.httpClient.Timeout == 0 {
		t.Error("expected non-zero timeout")
	}
}

func TestGatewayClientPost(t *testing.T) {
	// Set up a test server that echoes the request.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json")
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c := NewGatewayClient(srv.URL, "test-key")
	body := []byte(`{"hello":"world"}`)
	data, err := c.Post(context.Background(), "/test", body)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", result["status"])
	}
}

func TestGatewayClientGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":42}`))
	}))
	defer srv.Close()

	c := NewGatewayClient(srv.URL, "test-key")
	data, err := c.Get(context.Background(), "/resource")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	var result map[string]float64
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if result["value"] != 42 {
		t.Errorf("expected value=42, got %v", result["value"])
	}
}

func TestGatewayClientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	c := NewGatewayClient(srv.URL, "key")
	_, err := c.Get(context.Background(), "/bad")
	if err == nil {
		t.Fatal("expected error for 400 status")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Body != `{"error":"bad request"}` {
		t.Errorf("expected body with error message, got %q", apiErr.Body)
	}
}

func TestGatewayClientNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer srv.Close()

	c := NewGatewayClient(srv.URL, "key")
	_, err := c.Get(context.Background(), "/missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestGatewayClientNoAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "" {
			t.Error("expected no X-API-Key header when api key is empty")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewGatewayClient(srv.URL, "")
	_, err := c.Get(context.Background(), "/no-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 403, Body: "forbidden"}
	msg := err.Error()
	if msg != "API returned 403: forbidden" {
		t.Errorf("unexpected error message: %q", msg)
	}

	err2 := &APIError{StatusCode: 500}
	msg2 := err2.Error()
	if msg2 != "API returned 500" {
		t.Errorf("unexpected error message: %q", msg2)
	}
}

func TestGatewayClientHTTPClient(t *testing.T) {
	c := NewGatewayClient("http://localhost", "key")
	hc := c.HTTPClient()
	if hc == nil {
		t.Fatal("HTTPClient() should not return nil")
	}
	if hc.Timeout != c.httpClient.Timeout {
		t.Error("HTTPClient() should return the same client")
	}
}
