package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

type mockCompleter struct {
	result string
	err    error
}

func (m *mockCompleter) Complete(ctx context.Context, req client.CompletionRequest) (*client.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &client.CompletionResponse{OutputText: m.result}, nil
}

func TestMicroCompactResult(t *testing.T) {
	mock := &mockCompleter{result: "This is a summary of the tool result."}

	content := strings.Repeat("long result text ", 200) // > 2000 chars
	summary, ok := microCompactResult(context.Background(), mock, "file_read", content)

	if !ok {
		t.Fatal("Expected successful summarization")
	}
	if !strings.HasPrefix(summary, microCompactMarker) {
		t.Errorf("Summary should have microCompactMarker prefix, got: %s", summary)
	}
}

func TestMicroCompactResult_NilCompleter(t *testing.T) {
	_, ok := microCompactResult(context.Background(), nil, "file_read", "content")
	if ok {
		t.Error("Should return false for nil completer")
	}
}

func TestMicroCompactResult_EmptyResponse(t *testing.T) {
	mock := &mockCompleter{result: ""}

	_, ok := microCompactResult(context.Background(), mock, "file_read", "content")
	if ok {
		t.Error("Should return false for empty response")
	}
}

func TestMicroCompactResult_Error(t *testing.T) {
	mock := &mockCompleter{err: context.DeadlineExceeded}

	_, ok := microCompactResult(context.Background(), mock, "file_read", "content")
	if ok {
		t.Error("Should return false when completer returns error")
	}
}

func TestIsMicroCompacted(t *testing.T) {
	if isMicroCompacted("plain text") {
		t.Error("Plain text should not be micro-compacted")
	}
	if !isMicroCompacted(microCompactMarker + "summary here") {
		t.Error("Text with marker should be recognized as micro-compacted")
	}
	if isMicroCompacted("") {
		t.Error("Empty string should not match")
	}
}

func TestMicroCompactSkipTools(t *testing.T) {
	if !microCompactSkipTools["think"] {
		t.Error("think should be in skip list")
	}
	if !microCompactSkipTools["cloud_delegate"] {
		t.Error("cloud_delegate should be in skip list")
	}
}
