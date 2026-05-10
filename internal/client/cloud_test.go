package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCloudClient_Delegate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/delegate" {
			t.Errorf("expected /v1/delegate, got %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key=test-key, got %q", r.Header.Get("X-API-Key"))
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"type\":\"thinking\",\"data\":\"analyzing...\"}\n\n")
		fmt.Fprint(w, "data: {\"type\":\"text\",\"data\":\"Hello \"}\n\n")
		fmt.Fprint(w, "data: {\"type\":\"text\",\"data\":\"world\"}\n\n")
		fmt.Fprint(w, "data: {\"type\":\"done\",\"data\":\"\"}\n\n")
	}))
	defer server.Close()

	c := NewCloudClient(CloudConfig{
		Endpoint: server.URL,
		APIKey:   "test-key",
		Timeout:  30,
	})

	var progressEvents []CloudProgress
	resp, err := c.DelegateStream(context.Background(), CloudDelegateRequest{
		Task: "test task",
	}, func(p CloudProgress) {
		progressEvents = append(progressEvents, p)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Result != "Hello world" {
		t.Errorf("result = %q, want %q", resp.Result, "Hello world")
	}
	if len(progressEvents) != 3 {
		t.Errorf("got %d progress events, want 3", len(progressEvents))
	}
	if progressEvents[0].Type != "thinking" {
		t.Errorf("first event type = %q, want thinking", progressEvents[0].Type)
	}
}

func TestCloudClient_Delegate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"type\":\"error\",\"data\":\"out of credits\"}\n\n")
	}))
	defer server.Close()

	c := NewCloudClient(CloudConfig{Endpoint: server.URL, Timeout: 30})

	_, err := c.Delegate(context.Background(), CloudDelegateRequest{Task: "test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "cloud agent error: out of credits" {
		t.Errorf("error = %q", got)
	}
}

func TestCloudClient_Delegate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal server error")
	}))
	defer server.Close()

	c := NewCloudClient(CloudConfig{Endpoint: server.URL, Timeout: 30})

	_, err := c.Delegate(context.Background(), CloudDelegateRequest{Task: "test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCloudClient_Delegate_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewCloudClient(CloudConfig{Endpoint: server.URL, Timeout: 1})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := c.Delegate(ctx, CloudDelegateRequest{Task: "test"})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestCloudClient_Delegate_DoneWithData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"type\":\"text\",\"data\":\"partial\"}\n\n")
		fmt.Fprint(w, "data: {\"type\":\"done\",\"data\":\"final result\"}\n\n")
	}))
	defer server.Close()

	c := NewCloudClient(CloudConfig{Endpoint: server.URL, Timeout: 30})

	resp, err := c.Delegate(context.Background(), CloudDelegateRequest{Task: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When done has data, it takes precedence over accumulated text
	if resp.Result != "final result" {
		t.Errorf("result = %q, want %q", resp.Result, "final result")
	}
}

func TestCloudClient_Delegate_StreamDone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"type\":\"text\",\"data\":\"result text\"}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	c := NewCloudClient(CloudConfig{Endpoint: server.URL, Timeout: 30})

	resp, err := c.Delegate(context.Background(), CloudDelegateRequest{Task: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Result != "result text" {
		t.Errorf("result = %q, want %q", resp.Result, "result text")
	}
}
