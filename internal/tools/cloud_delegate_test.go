package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

func TestCloudDelegateTool_Info(t *testing.T) {
	tool := NewCloudDelegateTool(nil, nil)
	info := tool.Info()
	if info.Name != "cloud_delegate" {
		t.Errorf("name = %q, want cloud_delegate", info.Name)
	}
}

func TestCloudDelegateTool_NilClient(t *testing.T) {
	tool := NewCloudDelegateTool(nil, nil)
	result, err := tool.Run(context.Background(), `{"task":"test"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when cloud client is nil")
	}
}

func TestCloudDelegateTool_EmptyTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	}))
	defer server.Close()

	cc := client.NewCloudClient(client.CloudConfig{Endpoint: server.URL, Timeout: 30})
	tool := NewCloudDelegateTool(cc, nil)

	result, err := tool.Run(context.Background(), `{"task":""}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected validation error for empty task")
	}
}

func TestCloudDelegateTool_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "data: {\"type\":\"text\",\"data\":\"done result\"}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"type\":\"done\",\"data\":\"\"}\n\n")
	}))
	defer server.Close()

	cc := client.NewCloudClient(client.CloudConfig{Endpoint: server.URL, Timeout: 30})
	tool := NewCloudDelegateTool(cc, nil)

	result, err := tool.Run(context.Background(), `{"task":"research golang SSE"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %s", result.Content)
	}
	if result.Content != "done result" {
		t.Errorf("content = %q, want %q", result.Content, "done result")
	}
}

func TestCloudDelegateTool_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, "service unavailable")
	}))
	defer server.Close()

	cc := client.NewCloudClient(client.CloudConfig{Endpoint: server.URL, Timeout: 30})
	tool := NewCloudDelegateTool(cc, nil)

	result, err := tool.Run(context.Background(), `{"task":"test"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for server error")
	}
}

func TestCloudDelegateTool_RequiresApproval(t *testing.T) {
	tool := NewCloudDelegateTool(nil, nil)
	if !tool.RequiresApproval() {
		t.Error("cloud_delegate should require approval")
	}
}
