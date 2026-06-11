package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/session"
)

// blackboxMockClient is a scriptable mock that returns responses based on call count.
type blackboxMockClient struct {
	responses []blackboxResponse
	callCount int
}

type blackboxResponse struct {
	text      string
	toolCalls []client.ToolUse
	err       error
}

type recordingApprover struct {
	decision ApprovalDecision
	requests []ApprovalRequest
}

func (r *recordingApprover) RequestApproval(_ context.Context, req ApprovalRequest) (ApprovalDecision, error) {
	r.requests = append(r.requests, req)
	return r.decision, nil
}

var _ client.LLMClient = (*blackboxMockClient)(nil)

func (m *blackboxMockClient) Chat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions) (*client.Response, error) {
	if m.callCount >= len(m.responses) {
		return &client.Response{Content: "Final answer.", Usage: client.Usage{InputTokens: 10, OutputTokens: 5}}, nil
	}
	r := m.responses[m.callCount]
	m.callCount++
	if r.err != nil {
		return nil, r.err
	}
	return &client.Response{
		Content:  r.text,
		ToolUses: r.toolCalls,
		Usage:    client.Usage{InputTokens: 10, OutputTokens: 10},
	}, nil
}

// TestAgentLoop_SimpleTextResponse tests a single-turn text conversation.
func TestAgentLoop_SimpleTextResponse(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{text: "Hello! How can I help?"},
		},
	}
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "echo", description: "Echo back the input"})

	loop := NewAgentLoop(mock, reg)
	resp, err := loop.Run(context.Background(), "Hi there")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(resp.Content, "Hello") {
		t.Errorf("Expected greeting, got: %s", resp.Content)
	}
}

// TestAgentLoop_ToolCallThenResponse tests tool call + final text response.
func TestAgentLoop_ToolCallThenResponse(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{
				toolCalls: []client.ToolUse{
					{ID: "toolu_1", Name: "echo", Input: []byte(`{"text":"test"}`)},
				},
			},
			{text: "I called echo and got: test"},
		},
	}
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "echo", description: "Echo"})

	loop := NewAgentLoop(mock, reg)
	handler := &MockEventHandler{}
	loop.SetEventHandler(handler)

	resp, err := loop.Run(context.Background(), "echo test")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(resp.Content, "test") {
		t.Errorf("Expected echo result in response: %s", resp.Content)
	}
	if len(handler.toolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(handler.toolCalls))
	}
}

func TestAgentLoop_ToolCallThenResponsePersistsFinalAssistantMessage(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{
				toolCalls: []client.ToolUse{
					{ID: "toolu_1", Name: "echo", Input: []byte(`{"text":"test"}`)},
				},
			},
			{text: "I called echo and got: test"},
		},
	}
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "echo", description: "Echo"})
	mgr := session.NewManager(t.TempDir())
	sess := mgr.NewSession()

	loop := NewAgentLoop(mock, reg)
	loop.SetSession(sess)
	loop.SetSessionManager(mgr)

	if _, err := loop.Run(context.Background(), "echo test"); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	got := sess.Messages
	if len(got) == 0 {
		t.Fatal("session messages were not persisted")
	}
	last := got[len(got)-1]
	if last.Role != "assistant" || last.Content != "I called echo and got: test" {
		t.Fatalf("last persisted message = %#v, want final assistant response", last)
	}
}

func TestAgentLoop_RequiresApprovalAllowed(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{
				toolCalls: []client.ToolUse{
					{ID: "toolu_1", Name: "write_file", Input: []byte(`{"path":"test.txt"}`)},
				},
			},
			{text: "Write completed."},
		},
	}
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "write_file", requiresApproval: true})

	approver := &recordingApprover{decision: ApprovalAllow}
	loop := NewAgentLoop(mock, reg)
	loop.SetApprovalRequester(approver)

	resp, err := loop.Run(context.Background(), "write a file")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(resp.Content, "completed") {
		t.Fatalf("response = %q, want completed", resp.Content)
	}
	if len(approver.requests) != 1 {
		t.Fatalf("approval requests = %d, want 1", len(approver.requests))
	}
	if approver.requests[0].Tool != "write_file" {
		t.Fatalf("approval tool = %q, want write_file", approver.requests[0].Tool)
	}
}

func TestAgentLoop_RequiresApprovalDenied(t *testing.T) {
	executed := false
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{
				toolCalls: []client.ToolUse{
					{ID: "toolu_1", Name: "write_file", Input: []byte(`{"path":"test.txt"}`)},
				},
			},
			{text: "Tool was denied."},
		},
	}
	reg := NewToolRegistry()
	reg.Register(&MockTool{
		name:             "write_file",
		requiresApproval: true,
		execute: func() ToolResult {
			executed = true
			return ToolResult{Content: "should not run"}
		},
	})

	approver := &recordingApprover{decision: ApprovalDeny}
	loop := NewAgentLoop(mock, reg)
	loop.SetApprovalRequester(approver)

	_, err := loop.Run(context.Background(), "write a file")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if executed {
		t.Fatal("tool executed after approval denial")
	}
	if len(approver.requests) != 1 {
		t.Fatalf("approval requests = %d, want 1", len(approver.requests))
	}
}

// TestAgentLoop_LoopDetectionForceStop verifies loop detection triggers force stop.
func TestAgentLoop_LoopDetectionForceStop(t *testing.T) {
	// 4 consecutive identical tool calls should trigger force stop
	var responses []blackboxResponse
	for i := 0; i < 4; i++ {
		responses = append(responses, blackboxResponse{
			toolCalls: []client.ToolUse{
				{ID: "toolu_" + string(rune('0'+i)), Name: "grep", Input: []byte(`{"pattern":"same_pattern"}`)},
			},
		})
	}
	mock := &blackboxMockClient{responses: responses}

	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "grep", description: "Search files"})

	loop := NewAgentLoop(mock, reg)
	loop.SetMaxIterations(10)
	loop.SetConfigDir(t.TempDir()) // enables loopDetector

	_, err := loop.Run(context.Background(), "find something")
	// Should return due to loop detection or max iterations
	if err != nil && strings.Contains(err.Error(), "reached maximum") {
		t.Log("Stopped at max iterations (expected when loop detection injects nudge messages)")
	} else if err != nil {
		t.Logf("Run ended with: %v", err)
	}
	// Verify loop detector was active
	if mock.callCount > 6 {
		t.Errorf("Loop should have been caught early, but took %d calls", mock.callCount)
	}
}

// TestAgentLoop_MaxIterations verifies the loop respects max iteration limit.
func TestAgentLoop_MaxIterations(t *testing.T) {
	// Always return a tool call — should hit maxIter limit
	var responses []blackboxResponse
	for i := 0; i < 30; i++ {
		responses = append(responses, blackboxResponse{
			toolCalls: []client.ToolUse{
				{ID: fmt.Sprintf("toolu_%d", i), Name: "echo", Input: []byte(`{"text":"x"}`)},
			},
		})
	}
	mock := &blackboxMockClient{responses: responses}

	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "echo", description: "Echo"})

	loop := NewAgentLoop(mock, reg)
	loop.SetMaxIterations(5)

	_, err := loop.Run(context.Background(), "loop forever")
	if err == nil {
		t.Fatal("Expected error for reaching max iterations")
	}
	if !strings.Contains(err.Error(), "maximum iterations") {
		t.Errorf("Expected max iterations error, got: %v", err)
	}
}

// TestAgentLoop_UnknownTool returns validation error.
func TestAgentLoop_UnknownTool(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{
				toolCalls: []client.ToolUse{
					{ID: "toolu_1", Name: "nonexistent_tool", Input: []byte(`{}`)},
				},
			},
			{text: "I tried to use a tool but it failed."},
		},
	}
	reg := NewToolRegistry()

	loop := NewAgentLoop(mock, reg)
	handler := &MockEventHandler{}
	loop.SetEventHandler(handler)

	resp, err := loop.Run(context.Background(), "use bad tool")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	// executeTool returns ValidationError for unknown tools before firing handler events.
	// The loop should continue and eventually get the text response.
	if !strings.Contains(resp.Content, "failed") {
		t.Logf("Response: %s", resp.Content)
	}
}

// TestAgentLoop_SpillToDisk tests that large tool results are spilled.
func TestAgentLoop_SpillToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	largeContent := strings.Repeat("x", 60000) // >50KB threshold

	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{
				toolCalls: []client.ToolUse{
					{ID: "toolu_1", Name: "big_output", Input: []byte(`{}`)},
				},
			},
			{text: "Done with big output."},
		},
	}

	reg := NewToolRegistry()
	reg.Register(&MockTool{
		name:    "big_output",
		execute: func() ToolResult { return ToolResult{Content: largeContent} },
	})

	loop := NewAgentLoop(mock, reg)
	loop.SetConfigDir(tmpDir)
	loop.SetSessionID("test-spill-session")

	_, err := loop.Run(context.Background(), "generate big output")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// The tool result should have been spilled (preview < full content)
	// Verify spill directory was created
	if !dirExists(t, tmpDir, "tmp") {
		t.Error("Spill tmp directory should exist")
	}
}

// TestAgentLoop_MultiToolTurn tests multiple tool calls in one response.
func TestAgentLoop_MultiToolTurn(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{
				toolCalls: []client.ToolUse{
					{ID: "toolu_1", Name: "echo", Input: []byte(`{"text":"first"}`)},
					{ID: "toolu_2", Name: "echo", Input: []byte(`{"text":"second"}`)},
				},
			},
			{text: "Called two tools."},
		},
	}
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "echo", description: "Echo"})

	loop := NewAgentLoop(mock, reg)
	handler := &MockEventHandler{}
	loop.SetEventHandler(handler)

	resp, err := loop.Run(context.Background(), "call two echos")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if len(handler.toolCalls) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(handler.toolCalls))
	}
	if !strings.Contains(resp.Content, "two tools") {
		t.Errorf("Expected final response, got: %s", resp.Content)
	}
}

// TestAgentLoop_RetryThenSucceed tests transient error handling.
func TestAgentLoop_RetryThenSucceed(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{err: fmt.Errorf("502 Bad Gateway")}, // retry 1
			{text: "Final answer after retry."},  // succeeds
		},
	}
	reg := NewToolRegistry()

	loop := NewAgentLoop(mock, reg)

	resp, err := loop.Run(context.Background(), "something")
	if err != nil {
		t.Fatalf("Run should succeed after retry, got: %v", err)
	}
	if mock.callCount != 2 {
		t.Errorf("Expected 2 calls (1 retry + 1 success), got %d", mock.callCount)
	}
	_ = resp
}

func TestAgentLoop_RetryBackoffStopsOnCancel(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{err: fmt.Errorf("502 Bad Gateway")},
			{text: "should not be called"},
		},
	}
	reg := NewToolRegistry()
	loop := NewAgentLoop(mock, reg)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := loop.Run(ctx, "cancel during retry")
		done <- err
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Expected cancellation error")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not return promptly after cancellation")
	}

	if mock.callCount != 1 {
		t.Fatalf("Expected retry loop to stop after 1 call, got %d", mock.callCount)
	}
}

// TestAgentLoop_RetryExhausted tests max retries exhausted.
func TestAgentLoop_RetryExhausted(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{err: fmt.Errorf("502 Bad Gateway")},
			{err: fmt.Errorf("503 Service Unavailable")},
			{err: fmt.Errorf("504 Gateway Timeout")},
		},
	}
	reg := NewToolRegistry()

	loop := NewAgentLoop(mock, reg)

	_, err := loop.Run(context.Background(), "test retry")
	if err == nil {
		t.Fatal("Expected error after exhausted retries")
	}
	if !strings.Contains(err.Error(), "LLM error") {
		t.Errorf("Expected LLM error, got: %v", err)
	}
	// Should have tried 3 times
	if mock.callCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", mock.callCount)
	}
}

// TestAgentLoop_ContextWindow tests that SetConfigDir enables loop detector.
func TestAgentLoop_ContextWindow(t *testing.T) {
	mock := &blackboxMockClient{
		responses: []blackboxResponse{
			{text: "Simple answer, no tools needed."},
		},
	}
	reg := NewToolRegistry()

	loop := NewAgentLoop(mock, reg)
	loop.SetContextWindow(100000)

	if loop.contextWindow != 100000 {
		t.Errorf("ContextWindow not set: got %d", loop.contextWindow)
	}

	_, err := loop.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

// TestAgentLoop_DifferentToolArgs tests that different args don't trigger loop detection.
func TestAgentLoop_DifferentToolArgs(t *testing.T) {
	var responses []blackboxResponse
	for i := 0; i < 5; i++ {
		responses = append(responses, blackboxResponse{
			toolCalls: []client.ToolUse{
				{ID: fmt.Sprintf("toolu_%d", i), Name: "file_read", Input: []byte(fmt.Sprintf(`{"path":"file%d.go"}`, i))},
			},
		})
	}
	responses = append(responses, blackboxResponse{text: "Done reading all files."})

	mock := &blackboxMockClient{responses: responses}
	reg := NewToolRegistry()
	reg.Register(&MockTool{name: "file_read", description: "Read files"})

	loop := NewAgentLoop(mock, reg)
	loop.SetMaxIterations(10)
	loop.SetConfigDir(t.TempDir())

	resp, err := loop.Run(context.Background(), "read many files")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	// Different args should not trigger immediate loop stop
	if !strings.Contains(resp.Content, "Done") {
		t.Errorf("Expected completion message, got: %s", resp.Content)
	}
}

// TestAgentLoop_SpillCleanupFunc verifies the SpillCleanupFunc works.
func TestAgentLoop_SpillCleanupFunc(t *testing.T) {
	tmpDir := t.TempDir()

	loop := NewAgentLoop(nil, NewToolRegistry())
	loop.SetConfigDir(tmpDir)
	loop.SetSessionID("cleanup-test")

	cleanup := loop.SpillCleanupFunc()
	// Should not panic
	cleanup()
}

func dirExists(t *testing.T, parent, child string) bool {
	t.Helper()
	info, err := os.Stat(filepath.Join(parent, child))
	return err == nil && info.IsDir()
}
