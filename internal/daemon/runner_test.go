package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

// mockLLMClient implements agent.LLMClient for testing.
type mockLLMClient struct {
	t *testing.T
}

func (m *mockLLMClient) Chat(_ context.Context, systemPrompt string, messages []client.Message, _ []client.ToolDef, maxTokens int, _ *client.ChatOptions) (*client.Response, error) {
	return &client.Response{
		Content: "This is a mock response.",
		Usage: client.Usage{
			InputTokens:  10,
			OutputTokens: 20,
		},
		StopReason: "end_turn",
	}, nil
}

// mockHandler implements agent.EventHandler for testing.
type mockHandler struct {
	t          *testing.T
	toolCalls  int
	toolResults int
	texts      int
	usages     int
}

func (h *mockHandler) OnToolCall(name string, args string) {
	h.toolCalls++
}

func (h *mockHandler) OnToolResult(name string, result agent.ToolResult) {
	h.toolResults++
}

func (h *mockHandler) OnText(text string) {
	h.texts++
}

func (h *mockHandler) OnUsage(usage client.Usage) {
	h.usages++
}

func (h *mockHandler) OnStreamDelta(delta string) {
	// No-op for testing
}

func TestRunAgent_DefaultAgent(t *testing.T) {
	ctx := context.Background()

	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		AgentsDir:   t.TempDir(),
		LLMClient:   &mockLLMClient{t: t},
		Registry:    agent.NewToolRegistry(),
	}

	handler := &mockHandler{t: t}

	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "hello"}, handler)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}

	if resp.SessionID == "" {
		t.Error("expected non-empty SessionID")
	}

	if len(resp.Messages) == 0 {
		t.Error("expected at least one message")
	} else if resp.Messages[0] != "This is a mock response." {
		t.Errorf("unexpected message content: %s", resp.Messages[0])
	}

	if resp.Usage == nil {
		t.Error("expected non-nil Usage")
	} else {
		if resp.Usage["input_tokens"] != 10 {
			t.Errorf("expected 10 input tokens, got %d", resp.Usage["input_tokens"])
		}
		if resp.Usage["output_tokens"] != 20 {
			t.Errorf("expected 20 output tokens, got %d", resp.Usage["output_tokens"])
		}
		if resp.Usage["total_tokens"] != 30 {
			t.Errorf("expected 30 total tokens, got %d", resp.Usage["total_tokens"])
		}
	}

	if resp.Error != "" {
		t.Errorf("expected no error, got: %s", resp.Error)
	}

	// Verify handler was called with text and usage
	if handler.texts != 1 {
		t.Errorf("expected 1 OnText call, got %d", handler.texts)
	}
	if handler.usages != 1 {
		t.Errorf("expected 1 OnUsage call, got %d", handler.usages)
	}

	// Verify session was saved to disk
	sessionFile := filepath.Join(deps.StarclawDir, "sessions", resp.SessionID+".json")
	if _, err := os.Stat(sessionFile); err != nil {
		t.Errorf("expected session file at %s: %v", sessionFile, err)
	}
}

func TestRunAgent_NamedAgent(t *testing.T) {
	ctx := context.Background()

	agentsDir := t.TempDir()
	agentDir := filepath.Join(agentsDir, "helper")
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("You are a helpful assistant."), 0600); err != nil {
		t.Fatal(err)
	}

	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		AgentsDir:   agentsDir,
		LLMClient:   &mockLLMClient{t: t},
		Registry:    agent.NewToolRegistry(),
	}

	handler := &mockHandler{t: t}

	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "hello", Agent: "helper"}, handler)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}

	if resp.SessionID == "" {
		t.Error("expected non-empty SessionID")
	}
	if resp.Error != "" {
		t.Errorf("expected no error, got: %s", resp.Error)
	}

	// Session should be saved under the agent's subdirectory
	sessionFile := filepath.Join(deps.StarclawDir, "sessions", "helper", resp.SessionID+".json")
	if _, err := os.Stat(sessionFile); err != nil {
		t.Errorf("expected session file at %s: %v", sessionFile, err)
	}
}

func TestRunAgent_NamedAgentNotFound(t *testing.T) {
	ctx := context.Background()

	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		AgentsDir:   t.TempDir(), // empty — no agents
		LLMClient:   &mockLLMClient{t: t},
		Registry:    agent.NewToolRegistry(),
	}

	_, err := RunAgent(ctx, deps, RunAgentRequest{Text: "hello", Agent: "nonexistent"}, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent agent, got nil")
	}
}

func TestRunAgent_Validation(t *testing.T) {
	ctx := context.Background()

	// Nil deps
	_, err := RunAgent(ctx, nil, RunAgentRequest{Text: "hello"}, nil)
	if err == nil {
		t.Error("expected error for nil deps")
	}

	// Empty text
	_, err = RunAgent(ctx, &ServerDeps{}, RunAgentRequest{Text: ""}, nil)
	if err == nil {
		t.Error("expected error for empty text")
	}

	// Missing LLMClient
	_, err = RunAgent(ctx, &ServerDeps{
		StarclawDir: t.TempDir(),
		Registry:    agent.NewToolRegistry(),
		LLMClient:   nil,
	}, RunAgentRequest{Text: "hello"}, nil)
	if err == nil {
		t.Error("expected error for nil LLMClient")
	}

	// Missing Registry
	_, err = RunAgent(ctx, &ServerDeps{
		StarclawDir: t.TempDir(),
		LLMClient:   &mockLLMClient{t: t},
		Registry:    nil,
	}, RunAgentRequest{Text: "hello"}, nil)
	if err == nil {
		t.Error("expected error for nil Registry")
	}
}

func TestSessionsDirFor(t *testing.T) {
	deps := &ServerDeps{
		StarclawDir: "/home/user/.starclaw",
	}

	// Default agent sessions dir
	dir := sessionsDirFor(deps, "")
	if dir != "/home/user/.starclaw/sessions" {
		t.Errorf("expected /home/user/.starclaw/sessions, got %s", dir)
	}

	// Named agent sessions dir
	dir = sessionsDirFor(deps, "my-agent")
	if dir != "/home/user/.starclaw/sessions/my-agent" {
		t.Errorf("expected /home/user/.starclaw/sessions/my-agent, got %s", dir)
	}
}
