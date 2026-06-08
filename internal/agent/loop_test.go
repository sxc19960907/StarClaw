package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/session"
)

// MockEventHandler for testing
type MockEventHandler struct {
	toolCalls   []string
	toolResults []string
	texts       []string
	usages      []client.Usage
}

func (m *MockEventHandler) OnToolCall(name string, args string) {
	m.toolCalls = append(m.toolCalls, name)
}

func (m *MockEventHandler) OnToolResult(name string, result ToolResult) {
	m.toolResults = append(m.toolResults, result.Content)
}

func (m *MockEventHandler) OnText(text string) {
	m.texts = append(m.texts, text)
}

func (m *MockEventHandler) OnUsage(usage client.Usage) {
	m.usages = append(m.usages, usage)
}

func (m *MockEventHandler) OnStreamDelta(delta string) {
	// No-op for testing
}

func (m *MockEventHandler) OnPreamble(preamble string) {
	// No-op for testing
}

func TestNewAgentLoop(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	if loop == nil {
		t.Fatal("NewAgentLoop returned nil")
	}
	if loop.maxIter != 25 {
		t.Errorf("Expected default maxIter 25, got %d", loop.maxIter)
	}
	if loop.maxTokens != 8192 {
		t.Errorf("Expected default maxTokens 8192, got %d", loop.maxTokens)
	}
}

func TestAgentLoop_Setters(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	loop.SetMaxIterations(10)
	if loop.maxIter != 10 {
		t.Errorf("Expected maxIter 10, got %d", loop.maxIter)
	}

	loop.SetMaxTokens(4096)
	if loop.maxTokens != 4096 {
		t.Errorf("Expected maxTokens 4096, got %d", loop.maxTokens)
	}

	loop.SetResultTruncation(10000)
	if loop.resultTrunc != 10000 {
		t.Errorf("Expected resultTrunc 10000, got %d", loop.resultTrunc)
	}

	handler := &MockEventHandler{}
	loop.SetEventHandler(handler)
	if loop.handler == nil {
		t.Error("Event handler should be set")
	}

	loop.SetSystemPrompt("You are a test assistant")
	if loop.systemPrompt != "You are a test assistant" {
		t.Errorf("Expected system prompt set")
	}
}

type preflightCaptureClient struct {
	messages []client.Message
}

func (c *preflightCaptureClient) Chat(_ context.Context, _ string, messages []client.Message, _ []client.ToolDef, _ int, _ *client.ChatOptions) (*client.Response, error) {
	c.messages = append([]client.Message(nil), messages...)
	return &client.Response{Content: "ok"}, nil
}

type fakeMemoryPreflightProvider struct {
	result MemoryPreflightResult
}

func (p fakeMemoryPreflightProvider) PreflightMemory(_ context.Context, _ string) (MemoryPreflightResult, error) {
	return p.result, nil
}

func TestAgentLoopMemoryPreflightInjectsModelInputOnly(t *testing.T) {
	llm := &preflightCaptureClient{}
	loop := NewAgentLoop(llm, NewToolRegistry())
	dir := t.TempDir()
	mgr := session.NewManager(dir)
	sess := mgr.NewSession()
	loop.SetSession(sess)
	loop.SetSessionManager(mgr)
	loop.SetMemoryPreflightProvider(fakeMemoryPreflightProvider{result: MemoryPreflightResult{
		Attempted:    true,
		Provider:     "local",
		Outcome:      "matched",
		ResultsCount: 1,
		Block:        "<private_memory>\n- User likes local-first runtime.\n</private_memory>",
	}})

	resp, err := loop.Run(context.Background(), "what should I remember?")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("response = %q", resp.Content)
	}
	if len(llm.messages) != 1 || !strings.Contains(llm.messages[0].Content, "<private_memory>") {
		t.Fatalf("model messages = %#v, want private memory injected", llm.messages)
	}

	data, err := os.ReadFile(filepath.Join(dir, sess.ID+".json"))
	if err != nil {
		t.Fatalf("read session: %v", err)
	}
	if strings.Contains(string(data), "private_memory") || strings.Contains(string(data), "local-first runtime") {
		t.Fatalf("session leaked private memory: %s", data)
	}
	if !strings.Contains(string(data), "what should I remember?") {
		t.Fatalf("session missing original query: %s", data)
	}
}

func TestStripPrivateMemoryBlock(t *testing.T) {
	got := stripPrivateMemoryBlock("hello\n\n<private_memory>\nsecret\n</private_memory>")
	if got != "hello" {
		t.Fatalf("stripped = %q, want hello", got)
	}
}

func TestAgentLoop_buildTools(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()

	// Register a mock tool
	registry.Register(&MockTool{
		name:        "test_tool",
		description: "A test tool",
	})

	loop := NewAgentLoop(llmClient, registry)
	tools := loop.buildTools()

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", tools[0].Name)
	}
}

func TestAgentLoop_buildToolResultContent(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	toolUse := client.ToolUse{
		ID:   "toolu_123",
		Name: "test_tool",
	}

	result := ToolResult{
		Content: "success result",
		IsError: false,
	}

	content := loop.buildToolResultContent(toolUse, result)
	if !strings.Contains(content, "tool_result") {
		t.Error("Content should contain 'tool_result'")
	}
	if !strings.Contains(content, "toolu_123") {
		t.Error("Content should contain tool use ID")
	}

	// Test error result
	result.IsError = true
	content = loop.buildToolResultContent(toolUse, result)
	if !strings.Contains(content, "is_error") {
		t.Error("Error result should contain is_error flag")
	}
}

func TestAgentLoop_executeTool(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()

	// Register mock tool
	mockTool := &MockTool{
		name:    "mock_tool",
		execute: func() ToolResult { return ToolResult{Content: "mock result"} },
	}
	registry.Register(mockTool)

	loop := NewAgentLoop(llmClient, registry)
	handler := &MockEventHandler{}
	loop.SetEventHandler(handler)

	toolUse := client.ToolUse{
		ID:    "toolu_123",
		Name:  "mock_tool",
		Input: []byte(`{}`),
	}

	result := loop.executeTool(context.Background(), toolUse)

	if result.Content != "mock result" {
		t.Errorf("Expected 'mock result', got '%s'", result.Content)
	}

	if len(handler.toolCalls) != 1 {
		t.Error("Handler should have received tool call event")
	}
	if len(handler.toolResults) != 1 {
		t.Error("Handler should have received tool result event")
	}
}

func TestAgentLoop_executeTool_UnknownTool(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	toolUse := client.ToolUse{
		ID:   "toolu_123",
		Name: "unknown_tool",
	}

	result := loop.executeTool(context.Background(), toolUse)

	if !result.IsError {
		t.Error("Unknown tool should return error")
	}
	if result.ErrorCategory != ErrCategoryValidation {
		t.Errorf("Expected validation error, got %s", result.ErrorCategory)
	}
}

type streamingOnlyClient struct {
	streamCalls int
	chatCalls   int
	lastOpts    *client.ChatOptions
}

func (c *streamingOnlyClient) Chat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions) (*client.Response, error) {
	c.chatCalls++
	return nil, fmt.Errorf("non-streaming Chat should not be called after successful stream")
}

func (c *streamingOnlyClient) StreamChat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions, onDelta func(delta string)) (*client.Response, error) {
	c.streamCalls++
	c.lastOpts = opts
	if onDelta != nil {
		onDelta("streamed")
	}
	return &client.Response{Content: "streamed response"}, nil
}

func TestAgentLoop_StreamingSuccessDoesNotCallChat(t *testing.T) {
	llmClient := &streamingOnlyClient{}
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)
	loop.SetEnableStreaming(true)

	resp, err := loop.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if resp.Content != "streamed response" {
		t.Fatalf("Response content = %q, want streamed response", resp.Content)
	}
	if llmClient.streamCalls != 1 {
		t.Fatalf("StreamChat calls = %d, want 1", llmClient.streamCalls)
	}
	if llmClient.chatCalls != 0 {
		t.Fatalf("Chat calls = %d, want 0", llmClient.chatCalls)
	}
}

type optionCaptureClient struct {
	lastOpts *client.ChatOptions
}

func (c *optionCaptureClient) Chat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions) (*client.Response, error) {
	c.lastOpts = opts
	return &client.Response{Content: "ok"}, nil
}

func TestAgentLoop_PassesChatOptions(t *testing.T) {
	llmClient := &optionCaptureClient{}
	loop := NewAgentLoop(llmClient, NewToolRegistry())
	loop.SetThinking(&client.ThinkingConfig{Type: "enabled", BudgetTokens: 2048})
	loop.SetReasoningEffort("high")
	loop.SetSpecificModel("claude-opus-test")

	resp, err := loop.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("Response content = %q, want ok", resp.Content)
	}
	if llmClient.lastOpts == nil {
		t.Fatal("ChatOptions not passed")
	}
	if llmClient.lastOpts.Thinking == nil || llmClient.lastOpts.Thinking.Type != "enabled" || llmClient.lastOpts.Thinking.BudgetTokens != 2048 {
		t.Fatalf("Thinking options = %#v, want enabled/2048", llmClient.lastOpts.Thinking)
	}
	if llmClient.lastOpts.ReasoningEffort != "high" {
		t.Fatalf("ReasoningEffort = %q, want high", llmClient.lastOpts.ReasoningEffort)
	}
	if llmClient.lastOpts.SpecificModel != "claude-opus-test" {
		t.Fatalf("SpecificModel = %q, want claude-opus-test", llmClient.lastOpts.SpecificModel)
	}
}

func TestAgentLoop_TruncateResult(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()

	// Register mock tool that returns large result
	largeContent := strings.Repeat("x", 40000)
	mockTool := &MockTool{
		name:    "large_tool",
		execute: func() ToolResult { return ToolResult{Content: largeContent} },
	}
	registry.Register(mockTool)

	loop := NewAgentLoop(llmClient, registry)
	loop.SetResultTruncation(1000)

	toolUse := client.ToolUse{
		ID:    "toolu_123",
		Name:  "large_tool",
		Input: []byte(`{}`),
	}

	result := loop.executeTool(context.Background(), toolUse)

	if len(result.Content) <= 1000 {
		t.Error("Result should be truncated")
	}
	if !strings.Contains(result.Content, "truncated") {
		t.Error("Truncated result should indicate truncation")
	}
}

func TestAgentLoop_LastRunStatus_Default(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	status := loop.LastRunStatus()
	if status.Code != "" {
		t.Errorf("Expected empty RunStatus code, got: %q", status.Code)
	}
}

func TestAgentLoop_ContextBloat_Detection(t *testing.T) {
	// Build messages where tool results dominate (>50%)
	messages := []client.Message{
		{Role: "user", Content: "short query"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"test_tool","input":{}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"` + strings.Repeat("x", 5000) + `"}`},
	}

	detail := detectContextBloat(messages)
	if detail == "" {
		t.Fatal("Expected context bloat detail, got empty string")
	}
	if !strings.Contains(detail, "context") {
		t.Errorf("Detail should mention context: %s", detail)
	}
	if !strings.Contains(detail, "5000") || !strings.Contains(detail, "50") {
		t.Errorf("Detail should mention sizes: %s", detail)
	}
}

func TestAgentLoop_ContextBloat_NoBloat(t *testing.T) {
	// Build messages where tool results are small
	messages := []client.Message{
		{Role: "user", Content: "a long query that is very long " + strings.Repeat("x", 5000)},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"test_tool","input":{}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"small result"}`},
	}

	detail := detectContextBloat(messages)
	if detail != "" {
		t.Errorf("Expected no bloat, got: %s", detail)
	}
}

func TestAgentLoop_RunStatusHandler(t *testing.T) {
	llmClient := client.NewAnthropicClient("test", "", "")
	registry := NewToolRegistry()
	loop := NewAgentLoop(llmClient, registry)

	// Handler that implements RunStatusHandler
	handler := &runStatusRecorder{EventHandler: &MockEventHandler{}}
	loop.SetEventHandler(handler)

	// Create test messages that would trigger bloat
	messages := []client.Message{
		{Role: "user", Content: "query"},
		{Role: "assistant", Content: `{"type":"tool_use","id":"tu1","name":"test_tool","input":{}}`},
		{Role: "user", Content: `{"type":"tool_result","tool_use_id":"tu1","content":"` + strings.Repeat("x", 10000) + `"}`},
	}

	detail := detectContextBloat(messages)
	if detail == "" {
		t.Fatal("Expected context bloat detection")
	}

	// Directly set the status as the loop would
	loop.lastRunStatus = RunStatus{Code: "context_bloat", Detail: detail}
	if rs, ok := loop.handler.(RunStatusHandler); ok {
		rs.OnRunStatus("context_bloat", detail)
	}

	if len(handler.statuses) != 1 {
		t.Fatalf("Expected 1 status event, got %d", len(handler.statuses))
	}
	if handler.statuses[0].code != "context_bloat" {
		t.Errorf("Expected code 'context_bloat', got %q", handler.statuses[0].code)
	}

	status := loop.LastRunStatus()
	if status.Code != "context_bloat" {
		t.Errorf("LastRunStatus code should be 'context_bloat', got %q", status.Code)
	}
}

// runStatusRecorder records OnRunStatus events for testing.
type runStatusRecorder struct {
	EventHandler
	statuses []struct{ code, detail string }
	budgets  []TokenBudgetUsage
}

func (r *runStatusRecorder) OnRunStatus(code, detail string) {
	r.statuses = append(r.statuses, struct{ code, detail string }{code, detail})
}

func (r *runStatusRecorder) OnBudgetStatus(status TokenBudgetUsage) {
	r.budgets = append(r.budgets, status)
}

// MockTool for testing
type MockTool struct {
	name             string
	description      string
	requiresApproval bool
	execute          func() ToolResult
	source           ToolSource
	params           map[string]any
	required         []string
}

func (m *MockTool) Info() ToolInfo {
	params := m.params
	if params == nil {
		params = map[string]any{}
	}
	return ToolInfo{
		Name:        m.name,
		Description: m.description,
		Parameters:  params,
		Required:    m.required,
	}
}

func (m *MockTool) Run(ctx context.Context, args string) (ToolResult, error) {
	if m.execute != nil {
		return m.execute(), nil
	}
	return ToolResult{Content: "mock result"}, nil
}

func (m *MockTool) RequiresApproval() bool {
	return m.requiresApproval
}

func (m *MockTool) ToolSource() ToolSource {
	if m.source != "" {
		return m.source
	}
	return SourceLocal
}
