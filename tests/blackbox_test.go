// Package integration provides end-to-end black box tests for StarClaw
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/audit"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/session"
	"github.com/starclaw/starclaw/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestConfig returns the test configuration from environment variables
func getTestConfig() (apiKey, endpoint, model string) {
	apiKey = os.Getenv("ANTHROPIC_AUTH_TOKEN")
	endpoint = os.Getenv("ANTHROPIC_BASE_URL")
	model = os.Getenv("ANTHROPIC_MODEL")

	if apiKey == "" {
		apiKey = os.Getenv("STARCLAW_API_KEY")
	}
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-4-sonnet-20250514"
	}

	return apiKey, endpoint, model
}

// skipIfNoAPIKey skips the test if no API key is configured
func skipIfNoAPIKey(t *testing.T) (string, string, string) {
	apiKey, endpoint, model := getTestConfig()
	if apiKey == "" {
		t.Skip("Skipping black box test: No API key configured. Set ANTHROPIC_AUTH_TOKEN or STARCLAW_API_KEY")
	}
	return apiKey, endpoint, model
}

// TestAPIConnectivity tests basic API connectivity
func TestAPIConnectivity(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(3)

	ctx := context.Background()
	resp, err := agentLoop.Run(ctx, "Say 'API connectivity test passed' and nothing else.")

	require.NoError(t, err, "API call should succeed")
	require.NotNil(t, resp, "Response should not be nil")
	assert.NotEmpty(t, resp.Content, "Response should contain content")
	assert.Contains(t, strings.ToLower(resp.Content), "passed", "Response should indicate success")

	t.Logf("API Connectivity Test Response: %s", resp.Content)
	t.Logf("Token Usage - Input: %d, Output: %d", resp.Usage.InputTokens, resp.Usage.OutputTokens)
}

// TestSimpleConversation tests a simple back-and-forth conversation
func TestSimpleConversation(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(1)

	ctx := context.Background()
	resp, err := agentLoop.Run(ctx, "What is 2+2? Answer with just the number.")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "4")

	t.Logf("Simple Conversation Response: %s", resp.Content)
}

// TestToolCall_FileRead tests the file_read tool
func TestToolCall_FileRead(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello from black box test!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)

	ctx := context.Background()
	query := fmt.Sprintf("Read the file at %s and tell me its contents.", testFile)
	resp, err := agentLoop.Run(ctx, query)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Content, "Hello from black box test")

	t.Logf("File Read Tool Response: %s", resp.Content)
}

// TestToolCall_Bash tests the bash tool
func TestToolCall_Bash(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)

	ctx := context.Background()
	resp, err := agentLoop.Run(ctx, "Use the bash tool to run 'echo hello from bash test' and show me the output.")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, strings.ToLower(resp.Content), "hello from bash test")

	t.Logf("Bash Tool Response: %s", resp.Content)
}

// TestToolCall_Glob tests the glob tool
func TestToolCall_Glob(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	// Create some test files
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# README"), 0644)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)

	ctx := context.Background()
	query := fmt.Sprintf("Use the glob tool to find all .go files in %s and list them.", tempDir)
	resp, err := agentLoop.Run(ctx, query)

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Just verify the glob tool was used and returned some .go files
	assert.Contains(t, strings.ToLower(resp.Content), ".go")

	t.Logf("Glob Tool Response: %s", resp.Content)
}

// TestToolCall_Think tests the think tool
func TestToolCall_Think(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)

	ctx := context.Background()
	resp, err := agentLoop.Run(ctx, "Use the think tool to plan how to solve: 5+5*2. Show me your thinking, then give the answer.")

	require.NoError(t, err)
	require.NotNil(t, resp)
	// The model should either use the think tool or provide an answer
	assert.NotEmpty(t, resp.Content)

	t.Logf("Think Tool Response: %s", resp.Content)
}

// TestToolCall_SystemInfo tests the system_info tool
func TestToolCall_SystemInfo(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)

	ctx := context.Background()
	resp, err := agentLoop.Run(ctx, "Use the system_info tool to get information about this system and tell me what OS it's running.")

	require.NoError(t, err)
	require.NotNil(t, resp)
	// The response should mention the OS (darwin, linux, or windows)
	content := strings.ToLower(resp.Content)
	assert.True(t,
		strings.Contains(content, "darwin") ||
			strings.Contains(content, "linux") ||
			strings.Contains(content, "windows") ||
			strings.Contains(content, "mac"),
		"Response should contain OS information")

	t.Logf("System Info Tool Response: %s", resp.Content)
}

// TestToolCall_HTTP tests the http tool
func TestToolCall_HTTP(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)

	ctx := context.Background()
	// Use httpbin.org for testing (simple, reliable test API)
	resp, err := agentLoop.Run(ctx, "Use the http tool to GET https://httpbin.org/get and tell me the status code.")

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Should contain 200 status code or mention httpbin
	assert.True(t,
		strings.Contains(resp.Content, "200") ||
			strings.Contains(strings.ToLower(resp.Content), "httpbin"),
		"Response should indicate successful HTTP request")

	t.Logf("HTTP Tool Response: %s", resp.Content)
}

// TestSessionPersistence tests session save and resume
func TestSessionPersistence(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	tempDir := t.TempDir()
	sessionMgr := session.NewManager(tempDir)

	// Create a new session
	sess := sessionMgr.NewSession()
	sess.Title = "Black Box Test Session"

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)
	agentLoop.SetSession(sess)
	agentLoop.SetSessionManager(sessionMgr)

	ctx := context.Background()
	resp, err := agentLoop.Run(ctx, "Remember this number: 42")

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Save session
	err = sessionMgr.Save()
	require.NoError(t, err)

	// Verify session file was created
	sessionFile := filepath.Join(tempDir, sess.ID+".json")
	_, err = os.Stat(sessionFile)
	require.NoError(t, err, "Session file should exist")

	// Create new manager and resume session
	sessionMgr2 := session.NewManager(tempDir)
	resumedSess, err := sessionMgr2.Resume(sess.ID)
	require.NoError(t, err)
	require.NotNil(t, resumedSess)

	// Verify session data
	assert.Equal(t, sess.ID, resumedSess.ID)
	// Title is generated from first user message
	assert.NotEmpty(t, resumedSess.Title)
	assert.True(t, len(resumedSess.Messages) > 0, "Session should have messages")

	t.Logf("Session Persistence Test Passed - Session ID: %s", sess.ID)
}

// TestAuditLogging tests audit log creation
func TestAuditLogging(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	tempDir := t.TempDir()
	logger, err := audit.NewAuditLogger(tempDir)
	require.NoError(t, err)
	defer logger.Close()

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)
	agentLoop.SetAuditLogger(logger)
	agentLoop.SetSessionID("blackbox-test-session")

	ctx := context.Background()
	resp, err := agentLoop.Run(ctx, "Use the bash tool to run 'echo audit test'")

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Close logger to ensure flush
	logger.Close()

	// Give a moment for file write
	time.Sleep(100 * time.Millisecond)

	// Verify audit log was created
	auditFile := filepath.Join(tempDir, "audit.log")
	content, err := os.ReadFile(auditFile)
	require.NoError(t, err, "Audit log should exist")
	assert.NotEmpty(t, content, "Audit log should not be empty")

	t.Logf("Audit Log Content:\n%s", string(content))
}

// TestMultiToolChain tests chaining multiple tools
func TestMultiToolChain(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	// Create test environment
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "numbers.txt")
	os.WriteFile(testFile, []byte("1\n2\n3\n4\n5\n"), 0644)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(10)

	ctx := context.Background()
	query := fmt.Sprintf(
		"Read the file at %s, then use bash to count how many lines it has. "+
			"Finally, tell me the sum of all numbers in the file.",
		testFile)

	resp, err := agentLoop.Run(ctx, query)

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Should mention the sum (15) or line count (5)
	assert.True(t,
		strings.Contains(resp.Content, "15") || strings.Contains(resp.Content, "5"),
		"Response should contain the sum or line count")

	t.Logf("Multi-Tool Chain Response: %s", resp.Content)
}

// TestErrorHandling_BB tests graceful error handling in black box context
func TestErrorHandling_BB(t *testing.T) {
	apiKey, endpoint, model := skipIfNoAPIKey(t)

	llmClient := client.NewLLMClient(apiKey, endpoint, model)
	registry := tools.RegisterLocalTools()
	agentLoop := agent.NewAgentLoop(llmClient, registry)
	agentLoop.SetMaxIterations(5)

	ctx := context.Background()
	// Try to read a non-existent file
	resp, err := agentLoop.Run(ctx, "Read the file at /nonexistent/path/that/does/not/exist.txt")

	// Should not error out - the tool should return an error result that the model can handle
	// The model should respond with something about the file not being found
	require.NoError(t, err, "Agent should handle tool errors gracefully")
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Content)

	t.Logf("Error Handling Response: %s", resp.Content)
}

// TestAllToolsRegistered tests that all expected tools are in the registry
func TestAllToolsRegistered(t *testing.T) {
	registry := tools.RegisterLocalTools()
	require.NotNil(t, registry)

	expectedTools := []string{
		"file_read",
		"file_write",
		"file_edit",
		"glob",
		"directory_list",
		"grep",
		"think",
		"system_info",
		"http",
		"bash",
	}

	for _, toolName := range expectedTools {
		tool, ok := registry.Get(toolName)
		assert.True(t, ok, "Tool %s should be registered", toolName)
		if ok {
			info := tool.Info()
			assert.Equal(t, toolName, info.Name)
			assert.NotEmpty(t, info.Description)
			t.Logf("✓ Tool registered: %s - %s", info.Name, info.Description)
		}
	}
}
