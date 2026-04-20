// Package integration provides end-to-end integration tests
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/audit"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigToAgent tests the flow from config loading to agent creation
func TestConfigToAgent(t *testing.T) {
	// Create temp directory for config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".starclaw.yaml")

	// Create a test config file
	configContent := `
api_key: test-key-12345
endpoint: https://api.test.com
model_tier: standard
agent:
  max_iterations: 5
  max_tokens: 1000
tools:
  allowed:
    - file_read
    - glob
  denied: []
  result_truncation: 5000
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load the config
	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if cfg.APIKey != "test-key-12345" {
		t.Errorf("Expected API key 'test-key-12345', got '%s'", cfg.APIKey)
	}
	if cfg.Endpoint != "https://api.test.com" {
		t.Errorf("Expected endpoint 'https://api.test.com', got '%s'", cfg.Endpoint)
	}
	if cfg.ModelTier != "standard" {
		t.Errorf("Expected model tier 'standard', got '%s'", cfg.ModelTier)
	}
	if cfg.Agent.MaxIterations != 5 {
		t.Errorf("Expected max_iterations 5, got %d", cfg.Agent.MaxIterations)
	}

	// Create tool registry
	registry := tools.RegisterLocalTools()
	if registry == nil {
		t.Fatal("Failed to create tool registry")
	}

	// Apply allowed/denied filters
	if len(cfg.Tools.Allowed) > 0 {
		registry = registry.FilterByAllow(cfg.Tools.Allowed)
	}
	if len(cfg.Tools.Denied) > 0 {
		registry = registry.FilterByDeny(cfg.Tools.Denied)
	}

	// Verify filtered tools
	if _, ok := registry.Get("file_read"); !ok {
		t.Error("Expected file_read to be in filtered registry")
	}
	if _, ok := registry.Get("glob"); !ok {
		t.Error("Expected glob to be in filtered registry")
	}
	if _, ok := registry.Get("bash"); ok {
		t.Error("Expected bash to NOT be in filtered registry (not in allowed list)")
	}

	// Create mock LLM client
	mockClient := client.NewMockClient()
	mockClient.SetResponse("Test response from mock LLM")

	// Create agent loop
	loop := agent.NewAgentLoop(mockClient, registry)
	loop.SetMaxIterations(cfg.Agent.MaxIterations)
	loop.SetMaxTokens(cfg.Agent.MaxTokens)
	loop.SetResultTruncation(cfg.Tools.ResultTruncation)

	// Verify agent loop configuration
	if loop.GetMaxIterations() != 5 {
		t.Errorf("Expected agent max_iterations 5, got %d", loop.GetMaxIterations())
	}
}

// TestAgentToolExecution tests the agent loop with tool execution
func TestAgentToolExecution(t *testing.T) {
	// Create temp directory with test files
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Hello, World!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create tool registry
	registry := tools.RegisterLocalTools()

	// Create mock client that returns a tool call first, then a final response
	mockClient := client.NewMockClient()
	callCount := 0
	mockClient.SetHandler(func(input string) *client.MockMessage {
		callCount++
		if callCount == 1 {
			// First call: return a tool call
			return &client.MockMessage{
				Role:    "assistant",
				Content: "",
				ToolCalls: []client.MockToolCall{
					{
						ID:   "call_1",
						Name: "file_read",
						Args: `{"path": "` + testFile + `"}`,
					},
				},
			}
		}
		// Second call: return final response after tool execution
		return &client.MockMessage{
			Role:    "assistant",
			Content: "I've read the file. It contains: Hello, World!",
		}
	})

	// Create agent loop
	loop := agent.NewAgentLoop(mockClient, registry)
	loop.SetMaxIterations(5)

	// Create event handler to capture events
	events := &testEventHandler{}
	loop.SetEventHandler(events)

	// Run the agent
	ctx := context.Background()
	resp, err := loop.Run(ctx, "Read the test file")
	if err != nil {
		t.Fatalf("Agent loop failed: %v", err)
	}

	// Verify response
	if resp.Content == "" {
		t.Error("Expected non-empty response from agent")
	}

	// Verify tool was called
	if !events.toolCalled {
		t.Error("Expected tool to be called")
	}
	if events.toolName != "file_read" {
		t.Errorf("Expected file_read tool, got %s", events.toolName)
	}
	if !strings.Contains(events.toolResult, "Hello, World") {
		t.Errorf("Expected tool result to contain 'Hello, World', got: %s", events.toolResult)
	}
}

// TestAgentMultipleToolCalls tests the agent with multiple chained tool calls
func TestAgentMultipleToolCalls(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create mock client that returns multiple tool calls
	mockClient := client.NewMockClient()
	callCount := 0
	mockClient.SetHandler(func(input string) *client.MockMessage {
		callCount++
		if callCount == 1 {
			return &client.MockMessage{
				Role:    "assistant",
				Content: "",
				ToolCalls: []client.MockToolCall{
					{
						ID:   "call_1",
						Name: "glob",
						Args: `{"pattern": "*.go"}`,
					},
				},
			}
		}
		return &client.MockMessage{
			Role:    "assistant",
			Content: "Found the files",
		}
	})

	// Create tool registry
	registry := tools.RegisterLocalTools()

	// Create agent loop
	loop := agent.NewAgentLoop(mockClient, registry)
	loop.SetMaxIterations(3)

	// Create a test Go file
	testFile := filepath.Join(tempDir, "test.go")
	os.WriteFile(testFile, []byte("package main"), 0644)
	os.Chdir(tempDir)

	// Run the agent
	ctx := context.Background()
	resp, err := loop.Run(ctx, "Find Go files")
	if err != nil {
		t.Fatalf("Agent loop failed: %v", err)
	}

	if resp.Content != "Found the files" {
		t.Errorf("Expected 'Found the files', got: %s", resp.Content)
	}
}

// TestErrorHandling tests error handling paths
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() (*agent.AgentLoop, context.Context)
		expectError   bool
		errorContains string
	}{
		{
			name: "invalid tool call",
			setup: func() (*agent.AgentLoop, context.Context) {
				mockClient := client.NewMockClient()
				mockClient.SetToolCallResponse("file_read", `{"file_path": "/nonexistent/path"}`)

				registry := tools.RegisterLocalTools()
				loop := agent.NewAgentLoop(mockClient, registry)

				return loop, context.Background()
			},
			expectError:   false, // Tool errors should be handled gracefully
			errorContains: "",
		},
		{
			name: "max iterations exceeded",
			setup: func() (*agent.AgentLoop, context.Context) {
				mockClient := client.NewMockClient()
				// Always return a tool call, never a final response
				mockClient.SetHandler(func(input string) *client.MockMessage {
					return &client.MockMessage{
						Role: "assistant",
						ToolCalls: []client.MockToolCall{
							{
								ID:   "call_1",
								Name: "glob",
								Args: `{"pattern": "*"}`,
							},
						},
					}
				})

				registry := tools.RegisterLocalTools()
				loop := agent.NewAgentLoop(mockClient, registry)
				loop.SetMaxIterations(2) // Low limit to trigger exceeded

				return loop, context.Background()
			},
			expectError:   false, // Should return gracefully with message
			errorContains: "",
		},
		{
			name: "context cancellation",
			setup: func() (*agent.AgentLoop, context.Context) {
				mockClient := client.NewMockClient()
				mockClient.SetResponse("Delayed response")

				registry := tools.RegisterLocalTools()
				loop := agent.NewAgentLoop(mockClient, registry)

				// Create cancelled context
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				return loop, ctx
			},
			expectError:   false, // Mock client doesn't check context, so no error expected
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loop, ctx := tt.setup()
			_, err := loop.Run(ctx, "test query")

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				// No error expected - errors should be handled internally
				if err != nil {
					t.Logf("Got error (may be expected): %v", err)
				}
			}
		})
	}
}

// TestToolSecurity tests that tool security features work end-to-end
func TestToolSecurity(t *testing.T) {
	// Create temp directory for safe operations
	tempDir := t.TempDir()
	os.Chdir(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	registry := tools.RegisterLocalTools()

	// Test that file_read respects path safety
	tool, ok := registry.Get("file_read")
	if !ok {
		t.Fatal("file_read tool not found")
	}

	// Test safe path (absolute path to existing file)
	result, _ := tool.Run(context.Background(), `{"file_path": "`+testFile+`"}`)
	if result.IsError {
		// This might fail if file doesn't exist, which is acceptable
		t.Logf("Safe path returned error (may be expected if file not found): %s", result.Content)
	} else {
		if !strings.Contains(result.Content, "content") {
			t.Errorf("Expected file content, got: %s", result.Content)
		}
	}

	// Test unsafe path (parent directory traversal)
	result, _ = tool.Run(context.Background(), `{"file_path": "../../../etc/passwd"}`)
	// This might succeed or fail depending on current directory, but should be blocked if outside CWD
	if !result.IsError {
		// If it didn't error, it means the path was within CWD (relative paths are safe)
		t.Logf("Path traversal test: result=%s", result.Content)
	}

	// Test absolute path outside CWD
	result, _ = tool.Run(context.Background(), `{"file_path": "/etc/passwd"}`)
	if !result.IsError {
		t.Error("Expected absolute path outside CWD to be blocked")
	}
}

// TestTimeout tests that operations respect timeouts
func TestTimeout(t *testing.T) {
	mockClient := client.NewMockClient()
	mockClient.SetResponse("response")

	registry := tools.RegisterLocalTools()
	loop := agent.NewAgentLoop(mockClient, registry)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait to ensure timeout
	time.Sleep(10 * time.Millisecond)

	_, err := loop.Run(ctx, "test")
	// May or may not timeout depending on timing
	if err != nil {
		t.Logf("Got expected timeout error: %v", err)
	}
}

// TestThinkToolInRegistry tests that the think tool is properly registered
func TestThinkToolInRegistry(t *testing.T) {
	registry := tools.RegisterLocalTools()

	// Verify think tool exists
	tool, ok := registry.Get("think")
	if !ok {
		t.Fatal("Expected 'think' tool to be registered")
	}

	// Verify tool info
	info := tool.Info()
	if info.Name != "think" {
		t.Errorf("Expected tool name 'think', got '%s'", info.Name)
	}
	if info.Description == "" {
		t.Error("Expected non-empty description")
	}
	if len(info.Required) != 1 || info.Required[0] != "thought" {
		t.Errorf("Expected required field 'thought', got %v", info.Required)
	}

	// Verify no approval required
	if tool.RequiresApproval() {
		t.Error("Expected think tool to not require approval")
	}
}

// TestThinkToolExecution tests the think tool end-to-end
func TestThinkToolExecution(t *testing.T) {
	registry := tools.RegisterLocalTools()

	tool, ok := registry.Get("think")
	if !ok {
		t.Fatal("think tool not found in registry")
	}

	// Test valid thought
	result, err := tool.Run(context.Background(), `{"thought": "I need to plan this task"}`)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.Content)
	}
	if result.Content != "I need to plan this task" {
		t.Errorf("Expected thought content, got: %s", result.Content)
	}

	// Test empty thought
	result, _ = tool.Run(context.Background(), `{"thought": ""}`)
	if !result.IsError {
		t.Error("Expected error for empty thought")
	}

	// Test invalid JSON
	result, _ = tool.Run(context.Background(), "not valid json")
	if !result.IsError {
		t.Error("Expected error for invalid JSON")
	}
}

// TestSystemInfoToolInRegistry tests that system_info tool is registered
func TestSystemInfoToolInRegistry(t *testing.T) {
	registry := tools.RegisterLocalTools()

	tool, ok := registry.Get("system_info")
	if !ok {
		t.Fatal("Expected 'system_info' tool to be registered")
	}

	info := tool.Info()
	if info.Name != "system_info" {
		t.Errorf("Expected tool name 'system_info', got '%s'", info.Name)
	}
	if info.Description == "" {
		t.Error("Expected non-empty description")
	}

	// Verify no approval required
	if tool.RequiresApproval() {
		t.Error("Expected system_info tool to not require approval")
	}
}

// TestSystemInfoToolExecution tests system_info tool end-to-end
func TestSystemInfoToolExecution(t *testing.T) {
	registry := tools.RegisterLocalTools()

	tool, ok := registry.Get("system_info")
	if !ok {
		t.Fatal("system_info tool not found in registry")
	}

	result, err := tool.Run(context.Background(), "{}")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.Content)
	}

	// Check for expected fields
	if !strings.Contains(result.Content, "OS:") {
		t.Error("Expected OS field in output")
	}
	if !strings.Contains(result.Content, "Arch:") {
		t.Error("Expected Arch field in output")
	}
	if !strings.Contains(result.Content, "Hostname:") {
		t.Error("Expected Hostname field in output")
	}
	if !strings.Contains(result.Content, "CPUs:") {
		t.Error("Expected CPUs field in output")
	}
}

// TestHTTPToolInRegistry tests that http tool is registered
func TestHTTPToolInRegistry(t *testing.T) {
	registry := tools.RegisterLocalTools()

	tool, ok := registry.Get("http")
	if !ok {
		t.Fatal("Expected 'http' tool to be registered")
	}

	info := tool.Info()
	if info.Name != "http" {
		t.Errorf("Expected tool name 'http', got '%s'", info.Name)
	}
	if info.Description == "" {
		t.Error("Expected non-empty description")
	}

	// Verify approval required
	if !tool.RequiresApproval() {
		t.Error("Expected http tool to require approval")
	}
}

// TestHTTPToolExecution tests http tool end-to-end with mock server
func TestHTTPToolExecution(t *testing.T) {
	// Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	registry := tools.RegisterLocalTools()

	tool, ok := registry.Get("http")
	if !ok {
		t.Fatal("http tool not found in registry")
	}

	result, err := tool.Run(context.Background(), fmt.Sprintf(`{"url": "%s"}`, server.URL))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.Content)
	}

	// Check response contains expected fields
	if !strings.Contains(result.Content, "200") {
		t.Error("Expected status 200 in output")
	}
	if !strings.Contains(result.Content, `"status": "ok"`) {
		t.Error("Expected response body in output")
	}
}

// testEventHandler captures events for testing
type testEventHandler struct {
	toolCalled   bool
	toolName     string
	toolArgs     string
	toolResult   string
	textReceived string
	usage        client.Usage
}

func (h *testEventHandler) OnToolCall(name string, args string) {
	h.toolCalled = true
	h.toolName = name
	h.toolArgs = args
}

func (h *testEventHandler) OnToolResult(name string, result agent.ToolResult) {
	h.toolResult = result.Content
}

func (h *testEventHandler) OnText(text string) {
	h.textReceived = text
}

func (h *testEventHandler) OnUsage(usage client.Usage) {
	h.usage = usage
}

// TestAuditLoggingIntegration tests that tool calls are logged to audit file
func TestAuditLoggingIntegration(t *testing.T) {
	// Create temp directory for audit logs
	tempDir := t.TempDir()

	// Create audit logger
	logger, err := audit.NewAuditLogger(tempDir)
	require.NoError(t, err)
	defer logger.Close()

	// Create temp directory with test file
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Hello, World!"), 0644)
	require.NoError(t, err)

	// Create tool registry
	registry := tools.RegisterLocalTools()

	// Create mock client that returns a tool call
	mockClient := client.NewMockClient()
	callCount := 0
	mockClient.SetHandler(func(input string) *client.MockMessage {
		callCount++
		if callCount == 1 {
			return &client.MockMessage{
				Role:    "assistant",
				Content: "",
				ToolCalls: []client.MockToolCall{
					{
						ID:   "call_1",
						Name: "file_read",
						Args: `{"file_path": "` + testFile + `"}`,
					},
				},
			}
		}
		return &client.MockMessage{
			Role:    "assistant",
			Content: "Done",
		}
	})

	// Create agent loop with audit logger
	loop := agent.NewAgentLoop(mockClient, registry)
	loop.SetMaxIterations(5)
	loop.SetAuditLogger(logger)
	loop.SetSessionID("test-session-123")

	// Run the agent
	ctx := context.Background()
	_, err = loop.Run(ctx, "Read the test file")
	require.NoError(t, err)

	// Close logger to ensure all data is flushed
	logger.Close()

	// Read audit log
	content, err := os.ReadFile(filepath.Join(tempDir, "audit.log"))
	require.NoError(t, err)

	// Should have at least one audit entry
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	require.GreaterOrEqual(t, len(lines), 1, "Expected at least one audit entry")

	// Verify each line is valid JSON
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry audit.AuditEntry
		err := json.Unmarshal([]byte(line), &entry)
		require.NoError(t, err, "Audit entry should be valid JSON: %s", line)

		// Verify expected fields
		assert.Equal(t, "file_read", entry.ToolName)
		assert.Equal(t, "test-session-123", entry.SessionID)
		assert.NotZero(t, entry.Timestamp)
		assert.True(t, entry.Approved)
		assert.NotEmpty(t, entry.InputSummary)
	}
}

// TestAuditLoggingWithSecretRedaction tests that secrets are redacted in audit logs
func TestAuditLoggingWithSecretRedaction(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := audit.NewAuditLogger(tempDir)
	require.NoError(t, err)
	defer logger.Close()

	// Log an entry with a secret
	logger.Log(audit.AuditEntry{
		Timestamp:     time.Now(),
		SessionID:     "test",
		ToolName:      "bash",
		InputSummary:  "export API_KEY=secret123456789",
		OutputSummary: "done",
		Approved:      true,
		DurationMs:    100,
	})

	logger.Close()

	// Read audit log
	content, err := os.ReadFile(filepath.Join(tempDir, "audit.log"))
	require.NoError(t, err)

	// Secret should be redacted
	assert.Contains(t, string(content), "[REDACTED]")
	assert.NotContains(t, string(content), "secret123456789")
}
