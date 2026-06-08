package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/tools"
)

// mockLLMClient implements client.LLMClient for testing.
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

type captureLLMClient struct {
	mu           sync.Mutex
	systemPrompt string
	tools        []client.ToolDef
	maxTokens    int
	opts         *client.ChatOptions
}

type browserLeaseLLMClient struct {
	calls int
}

func (c *browserLeaseLLMClient) Chat(_ context.Context, _ string, _ []client.Message, _ []client.ToolDef, _ int, _ *client.ChatOptions) (*client.Response, error) {
	c.calls++
	if c.calls == 1 {
		return &client.Response{
			ToolUses: []client.ToolUse{{
				ID:    "toolu_browser",
				Name:  "browser",
				Input: json.RawMessage(`{"action":"status"}`),
			}},
		}, nil
	}
	return &client.Response{Content: "done"}, nil
}

func (c *captureLLMClient) Chat(_ context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions) (*client.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.systemPrompt = systemPrompt
	c.tools = append([]client.ToolDef(nil), tools...)
	c.maxTokens = maxTokens
	if opts != nil {
		copied := *opts
		c.opts = &copied
	}
	return &client.Response{Content: "captured"}, nil
}

func (c *captureLLMClient) toolNames() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	names := make([]string, 0, len(c.tools))
	for _, tool := range c.tools {
		names = append(names, tool.Name)
	}
	return names
}

// mockHandler implements agent.EventHandler for testing.
type mockHandler struct {
	t           *testing.T
	toolCalls   int
	toolResults int
	texts       int
	usages      int
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

func (h *mockHandler) OnPreamble(preamble string) {
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

func TestRunAgentReleasesBrowserLease(t *testing.T) {
	ctx := context.Background()
	browser := &tools.BrowserTool{}
	registry := agent.NewToolRegistry()
	registry.Register(browser)
	llm := &browserLeaseLLMClient{}

	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		AgentsDir:   t.TempDir(),
		LLMClient:   llm,
		Registry:    registry,
	}
	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "browser status"}, nil)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("response error = %q", resp.Error)
	}
	if got := tools.BrowserOwnerActiveCount(browser); got != 0 {
		t.Fatalf("browser owner count after run = %d, want 0", got)
	}
	if got := tools.GlobalBrowserTrackerActiveCountForTest(); got != 0 {
		t.Fatalf("global browser active count = %d, want 0", got)
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

func TestRunAgent_NamedAgentRuntimeParity(t *testing.T) {
	ctx := context.Background()

	agentsDir := t.TempDir()
	agentDir := filepath.Join(agentsDir, "helper")
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte("You are a daemon helper."), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "MEMORY.md"), []byte("Remember daemon parity memory."), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "config.yaml"), []byte(`agent:
  model: daemon-agent-model
  max_iterations: 3
  max_tokens: 1234
  thinking: true
  thinking_mode: enabled
  thinking_budget: 222
  reasoning_effort: high
tools:
  allow:
    - allowed_tool
    - denied_tool
  deny:
    - denied_tool
`), 0600); err != nil {
		t.Fatal(err)
	}

	registry := agent.NewToolRegistry()
	registry.Register(agent.TestTool("allowed_tool"))
	registry.Register(agent.TestTool("denied_tool"))
	registry.Register(agent.TestTool("other_tool"))

	llmClient := &captureLLMClient{}
	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		Config: &config.Config{
			Agent: config.AgentConfig{
				MaxIterations:   25,
				MaxTokens:       8192,
				Thinking:        false,
				ThinkingMode:    "adaptive",
				ThinkingBudget:  10000,
				ReasoningEffort: "",
				Model:           "global-model",
			},
			Tools: config.ToolsConfig{
				ResultTruncation: 30000,
			},
		},
		AgentsDir: agentsDir,
		LLMClient: llmClient,
		Registry:  registry,
	}

	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "hello", Agent: "helper"}, nil)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("expected no response error, got %q", resp.Error)
	}
	if resp.Messages[0] != "captured" {
		t.Fatalf("message = %q, want captured", resp.Messages[0])
	}

	if llmClient.systemPrompt == "" {
		t.Fatal("expected system prompt to be captured")
	}
	if !strings.Contains(llmClient.systemPrompt, "You are a daemon helper.") {
		t.Fatalf("system prompt missing agent prompt: %q", llmClient.systemPrompt)
	}
	if !strings.Contains(llmClient.systemPrompt, "<agent_memory>") ||
		!strings.Contains(llmClient.systemPrompt, "Remember daemon parity memory.") {
		t.Fatalf("system prompt missing agent memory: %q", llmClient.systemPrompt)
	}
	if llmClient.maxTokens != 1234 {
		t.Fatalf("maxTokens = %d, want 1234", llmClient.maxTokens)
	}
	if llmClient.opts == nil {
		t.Fatal("expected chat options")
	}
	if llmClient.opts.SpecificModel != "daemon-agent-model" {
		t.Fatalf("SpecificModel = %q, want daemon-agent-model", llmClient.opts.SpecificModel)
	}
	if llmClient.opts.ReasoningEffort != "high" {
		t.Fatalf("ReasoningEffort = %q, want high", llmClient.opts.ReasoningEffort)
	}
	if llmClient.opts.Thinking == nil ||
		llmClient.opts.Thinking.Type != "enabled" ||
		llmClient.opts.Thinking.BudgetTokens != 222 {
		t.Fatalf("Thinking opts = %#v, want enabled/222", llmClient.opts.Thinking)
	}

	names := llmClient.toolNames()
	if len(names) != 1 || names[0] != "allowed_tool" {
		t.Fatalf("tools = %#v, want [allowed_tool]", names)
	}
	if _, ok := deps.Registry.Get("denied_tool"); !ok {
		t.Fatal("base registry should not be mutated")
	}
	if _, ok := deps.Registry.Get("other_tool"); !ok {
		t.Fatal("base registry should keep tools excluded from per-run filter")
	}
}

func TestRunAgent_RequestModelOverridesConfig(t *testing.T) {
	ctx := context.Background()
	llmClient := &captureLLMClient{}
	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		Config: &config.Config{
			Agent: config.AgentConfig{
				MaxIterations: 25,
				MaxTokens:     8192,
				Model:         "configured-model",
			},
			Tools: config.ToolsConfig{ResultTruncation: 30000},
		},
		AgentsDir: t.TempDir(),
		LLMClient: llmClient,
		Registry:  agent.NewToolRegistry(),
	}

	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "hello", Model: "request-model"}, nil)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("expected no response error, got %q", resp.Error)
	}
	if llmClient.opts == nil {
		t.Fatal("expected chat options")
	}
	if llmClient.opts.SpecificModel != "request-model" {
		t.Fatalf("SpecificModel = %q, want request-model", llmClient.opts.SpecificModel)
	}
}

func TestRunAgent_SurfacesBudgetStatus(t *testing.T) {
	ctx := context.Background()
	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		Config: &config.Config{
			Agent: config.AgentConfig{
				MaxIterations: 25,
				MaxTokens:     8192,
				TokenBudget: config.TokenBudgetConfig{
					MaxTotalTokens: 100,
					HardStop:       true,
				},
			},
			Tools: config.ToolsConfig{ResultTruncation: 30000},
		},
		AgentsDir: t.TempDir(),
		LLMClient: &mockLLMClient{t: t},
		Registry:  agent.NewToolRegistry(),
	}

	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "hello"}, nil)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if resp.BudgetStatus == nil {
		t.Fatal("expected budget status")
	}
	if resp.BudgetStatus.Status != agent.TokenBudgetStatusOK {
		t.Fatalf("budget status = %#v, want ok", resp.BudgetStatus)
	}
	if resp.BudgetStatus.TotalTokens != 30 {
		t.Fatalf("total tokens = %d, want 30", resp.BudgetStatus.TotalTokens)
	}
}

func TestRunAgent_SurfacesRoutingMetadata(t *testing.T) {
	ctx := context.Background()
	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		Config: &config.Config{
			Agent: config.AgentConfig{
				MaxIterations: 25,
				MaxTokens:     8192,
			},
			Tools: config.ToolsConfig{ResultTruncation: 30000},
		},
		AgentsDir: t.TempDir(),
		LLMClient: &mockLLMClient{t: t},
		Registry:  agent.NewToolRegistry(),
	}

	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "Research and cite sources for this claim"}, nil)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if resp.Routing == nil {
		t.Fatal("expected routing metadata")
	}
	if resp.Routing.Complexity != agent.ComplexityEvidenceHeavy {
		t.Fatalf("complexity = %q, want evidence_heavy", resp.Routing.Complexity)
	}
	if resp.Routing.Route != agent.RouteResearch {
		t.Fatalf("route = %q, want research", resp.Routing.Route)
	}
	if resp.Fallback != nil {
		t.Fatalf("fallback = %#v, want nil", resp.Fallback)
	}
}

func TestRunAgent_BudgetExhaustionFallback(t *testing.T) {
	ctx := context.Background()
	deps := &ServerDeps{
		StarclawDir: t.TempDir(),
		Config: &config.Config{
			Agent: config.AgentConfig{
				MaxIterations: 25,
				MaxTokens:     8192,
				TokenBudget: config.TokenBudgetConfig{
					MaxInputTokens: 1,
					HardStop:       true,
				},
			},
			Tools: config.ToolsConfig{ResultTruncation: 30000},
		},
		AgentsDir: t.TempDir(),
		LLMClient: &mockLLMClient{t: t},
		Registry:  agent.NewToolRegistry(),
	}

	resp, err := RunAgent(ctx, deps, RunAgentRequest{Text: "This prompt should exceed the tiny input budget"}, nil)
	if err != nil {
		t.Fatalf("RunAgent failed: %v", err)
	}
	if resp.BudgetStatus == nil || resp.BudgetStatus.Status != agent.TokenBudgetStatusExhausted {
		t.Fatalf("budget status = %#v, want exhausted", resp.BudgetStatus)
	}
	if resp.Fallback == nil {
		t.Fatal("expected fallback metadata")
	}
	if resp.Fallback.Reason != agent.FallbackBudgetExhausted {
		t.Fatalf("fallback reason = %q, want budget_exhausted", resp.Fallback.Reason)
	}
	if resp.Fallback.Route != agent.RouteBudget {
		t.Fatalf("fallback route = %q, want budget_guard", resp.Fallback.Route)
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
