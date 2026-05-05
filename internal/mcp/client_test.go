package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewClientManager(t *testing.T) {
	cm := NewClientManager()
	if cm == nil {
		t.Fatal("NewClientManager() returned nil")
	}
	if cm.clients == nil {
		t.Error("clients map not initialized")
	}
	if cm.configs == nil {
		t.Error("configs map not initialized")
	}
	if cm.toolCache == nil {
		t.Error("toolCache map not initialized")
	}
}

func TestClientManager_IsConnected(t *testing.T) {
	cm := NewClientManager()

	if cm.IsConnected("test-server") {
		t.Error("IsConnected should return false for unknown server")
	}

	cm.clients["test-server"] = &mockMCPClient{}

	if !cm.IsConnected("test-server") {
		t.Error("IsConnected should return true for connected server")
	}
}

func TestClientManager_ConnectedServers(t *testing.T) {
	cm := NewClientManager()

	servers := cm.ConnectedServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(servers))
	}

	cm.clients["server1"] = &mockMCPClient{}
	cm.clients["server2"] = &mockMCPClient{}

	servers = cm.ConnectedServers()
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}
}

func TestClientManager_ConnectAll_DisabledSkip(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	servers := map[string]MCPServerConfig{
		"disabled-server": {
			Command:  "echo",
			Disabled: true,
		},
	}

	tools, err := cm.ConnectAll(ctx, servers)
	if err != nil {
		t.Logf("ConnectAll returned error: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools from disabled server, got %d", len(tools))
	}

	if _, exists := cm.configs["disabled-server"]; exists {
		t.Error("Disabled server should not be in configs")
	}
}

func TestClientManager_ConnectAll_Timeout(t *testing.T) {
	cm := NewClientManager()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	servers := map[string]MCPServerConfig{
		"slow-server": {
			Command: "sleep",
			Args:    []string{"10"},
		},
	}

	_, _ = cm.ConnectAll(ctx, servers)
}

func TestClientManager_Close(t *testing.T) {
	cm := NewClientManager()

	cm.clients["server1"] = &mockMCPClient{}
	cm.clients["server2"] = &mockMCPClient{}

	cm.Close()

	if len(cm.clients) != 0 {
		t.Error("Close() should clear all clients")
	}
}

func TestClientManager_CallTool_NotConnected(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	content, isError, err := cm.CallTool(ctx, "unknown-server", "test-tool", nil)
	if err == nil {
		t.Error("CallTool should return error for unknown server")
	}
	if content != "" {
		t.Error("CallTool should return empty content on error")
	}
	if !isError {
		t.Error("CallTool should return isError=true on error")
	}
}

func TestClientManager_CallTool_Connected(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	// Inject mock client that returns actual content
	mock := &mockToolClient{
		callToolResult: &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "hello world"},
			},
			IsError: false,
		},
	}
	cm.clients["test-server"] = mock
	cm.configs["test-server"] = MCPServerConfig{Command: "echo"}

	content, isError, err := cm.CallTool(ctx, "test-server", "test-tool", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isError {
		t.Error("expected isError=false")
	}
	if content != "hello world" {
		t.Errorf("expected 'hello world', got %q", content)
	}
}

func TestClientManager_CallTool_Reconnect(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	// Inject config so lazy-start can attempt reconnect
	cm.configs["test-server"] = MCPServerConfig{
		Command: "echo",
		Args:    []string{"hello"},
	}

	content, isError, err := cm.CallTool(ctx, "test-server", "test-tool", nil)
	// The reconnect will fail because echo isn't a real MCP server,
	// but it shouldn't panic and should return an error
	if err == nil && !isError {
		t.Log("reconnect succeeded (unexpected for echo command)")
	}
	_ = content
}

func TestClientManager_CachedTools(t *testing.T) {
	cm := NewClientManager()

	tools := cm.CachedTools("test-server")
	if tools != nil {
		t.Error("CachedTools should return nil for unknown server")
	}

	cm.toolCache["test-server"] = []RemoteTool{
		{ServerName: "test-server", Tool: mcp.Tool{Name: "tool1"}},
		{ServerName: "test-server", Tool: mcp.Tool{Name: "tool2"}},
	}

	tools = cm.CachedTools("test-server")
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	// Verify it's a copy
	cm.toolCache["test-server"][0] = RemoteTool{ServerName: "test-server", Tool: mcp.Tool{Name: "modified"}}
	if tools[0].Tool.Name != "tool1" {
		t.Error("CachedTools should return a copy, not the original")
	}
}

func TestClientManager_ConfigFor(t *testing.T) {
	cm := NewClientManager()

	cfg, ok := cm.ConfigFor("unknown")
	if ok {
		t.Error("ConfigFor should return false for unknown server")
	}
	if cfg.Command != "" {
		t.Error("ConfigFor should return zero config for unknown server")
	}

	cm.configs["test-server"] = MCPServerConfig{Command: "npx", Args: []string{"-y", "test"}}
	cfg, ok = cm.ConfigFor("test-server")
	if !ok {
		t.Error("ConfigFor should return true for known server")
	}
	if cfg.Command != "npx" {
		t.Errorf("expected command 'npx', got %q", cfg.Command)
	}
}

func TestClientManager_Disconnect(t *testing.T) {
	cm := NewClientManager()

	mock := &mockMCPClient{}
	cm.clients["test-server"] = mock

	cm.Disconnect("test-server")

	if cm.IsConnected("test-server") {
		t.Error("Disconnect should remove the client")
	}

	// Disconnect should preserve config
	cm.configs["test-server"] = MCPServerConfig{Command: "echo"}
	cm.clients["test-server"] = &mockMCPClient{}
	cm.Disconnect("test-server")
	if _, ok := cm.configs["test-server"]; !ok {
		t.Error("Disconnect should preserve config")
	}
}

func TestClientManager_ProbeTransport(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	// Not connected
	err := cm.ProbeTransport(ctx, "unknown")
	if err == nil {
		t.Error("ProbeTransport should fail for unknown server")
	}

	// Connected with working mock
	mock := &mockToolClient{
		listToolsResult: &mcp.ListToolsResult{},
	}
	cm.clients["test-server"] = mock
	err = cm.ProbeTransport(ctx, "test-server")
	if err != nil {
		t.Errorf("ProbeTransport should succeed: %v", err)
	}

	// Connected with failing mock
	failMock := &mockToolClient{
		listToolsErr: fmt.Errorf("connection lost"),
	}
	cm.clients["fail-server"] = failMock
	err = cm.ProbeTransport(ctx, "fail-server")
	if err == nil {
		t.Error("ProbeTransport should fail when ListTools fails")
	}
}

func TestClientManager_Reconnect(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	// No config
	_, err := cm.Reconnect(ctx, "unknown")
	if err == nil {
		t.Error("Reconnect should fail without config")
	}

	// With config (will fail because not a real MCP server)
	cm.configs["test-server"] = MCPServerConfig{
		Command: "echo",
		Args:    []string{"hello"},
	}
	_, err = cm.Reconnect(ctx, "test-server")
	if err == nil {
		t.Log("reconnect succeeded (unexpected for echo)")
	}
}

func TestClientManager_ConcurrentAccess(t *testing.T) {
	cm := NewClientManager()

	// Seed with mock clients
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("server-%d", i)
		cm.clients[name] = &mockMCPClient{}
		cm.configs[name] = MCPServerConfig{Command: "echo"}
	}

	var wg sync.WaitGroup

	// Concurrent ConnectedServers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cm.ConnectedServers()
		}()
	}

	// Concurrent IsConnected
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cm.IsConnected("server-1")
		}()
	}

	// Concurrent CachedTools
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cm.CachedTools("server-1")
		}()
	}

	// Concurrent ConfigFor
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cm.ConfigFor("server-1")
		}()
	}

	wg.Wait()
	// Should not panic or deadlock
}

func TestBuildContext(t *testing.T) {
	servers := map[string]MCPServerConfig{
		"github": {
			Context: "GitHub API access for PRs and issues",
		},
		"disabled-server": {
			Context:  "Should not appear",
			Disabled: true,
		},
		"empty-server": {
			Context: "",
		},
		"fetch": {
			Context: "Web fetch capability",
		},
	}

	result := BuildContext(servers)
	if result == "" {
		t.Error("BuildContext should return non-empty string")
	}
	if !contains(result, "[github]") {
		t.Error("BuildContext should include github context")
	}
	if !contains(result, "[fetch]") {
		t.Error("BuildContext should include fetch context")
	}
	if contains(result, "disabled-server") {
		t.Error("BuildContext should skip disabled servers")
	}
	if contains(result, "empty-server") {
		t.Error("BuildContext should skip empty context")
	}
}

func TestBuildContext_Empty(t *testing.T) {
	result := BuildContext(nil)
	if result != "" {
		t.Error("BuildContext with nil should return empty string")
	}
	result = BuildContext(map[string]MCPServerConfig{})
	if result != "" {
		t.Error("BuildContext with empty map should return empty string")
	}
}

func TestBuildEnvSlice(t *testing.T) {
	env := map[string]string{
		"TOKEN": "secret",
		"HOME":  "/home/user",
	}
	result := buildEnvSlice(env)
	if len(result) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(result))
	}
	// Results should be key=value format
	for _, s := range result {
		if !containsAny(s, []string{"TOKEN=secret", "HOME=/home/user"}) {
			t.Errorf("unexpected env entry: %s", s)
		}
	}
}

func TestBuildEnvSlice_Empty(t *testing.T) {
	result := buildEnvSlice(nil)
	if result != nil {
		t.Error("buildEnvSlice with nil should return nil")
	}
	result = buildEnvSlice(map[string]string{})
	if result != nil {
		t.Error("buildEnvSlice with empty map should return nil")
	}
}

func TestIsTransportError(t *testing.T) {
	tests := []struct {
		err     error
		isTrans bool
	}{
		{fmt.Errorf("broken pipe"), true},
		{fmt.Errorf("use of closed network connection"), true},
		{fmt.Errorf("read/write on closed pipe"), true},
		{fmt.Errorf("signal: killed"), true},
		{fmt.Errorf("process already finished"), true},
		{fmt.Errorf("invalid tool name"), false},
		{fmt.Errorf("permission denied"), false},
	}
	for _, tt := range tests {
		if got := isTransportError(tt.err); got != tt.isTrans {
			t.Errorf("isTransportError(%q) = %v, want %v", tt.err, got, tt.isTrans)
		}
	}
}

func TestMCPServerConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config MCPServerConfig
	}{
		{
			name: "stdio server",
			config: MCPServerConfig{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			},
		},
		{
			name: "http server without URL",
			config: MCPServerConfig{
				Type:    "http",
				Command: "curl",
			},
		},
		{
			name: "disabled server",
			config: MCPServerConfig{
				Command:  "echo",
				Disabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("failed to marshal config: %v", err)
			}
			var roundtripped MCPServerConfig
			if err := json.Unmarshal(data, &roundtripped); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}
			if roundtripped.Command != tt.config.Command {
				t.Errorf("Command roundtrip failed: %q != %q", roundtripped.Command, tt.config.Command)
			}
		})
	}
}

func TestRemoteTool_JSON(t *testing.T) {
	rt := RemoteTool{
		ServerName: "test-server",
		Tool: mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
			InputSchema: mcp.ToolInputSchema{
				Properties: map[string]interface{}{
					"param1": map[string]interface{}{"type": "string"},
				},
				Required: []string{"param1"},
			},
		},
	}

	data, err := json.Marshal(rt)
	if err != nil {
		t.Fatalf("failed to marshal RemoteTool: %v", err)
	}

	var rt2 RemoteTool
	if err := json.Unmarshal(data, &rt2); err != nil {
		t.Fatalf("failed to unmarshal RemoteTool: %v", err)
	}

	if rt2.ServerName != "test-server" {
		t.Errorf("ServerName mismatch: %q", rt2.ServerName)
	}
	if rt2.Tool.Name != "test-tool" {
		t.Errorf("Tool.Name mismatch: %q", rt2.Tool.Name)
	}
}

// mockMCPClient implements mcpclient.MCPClient for testing.
type mockMCPClient struct {
	closeCalled bool
}

func (m *mockMCPClient) Initialize(ctx context.Context, req mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{}, nil
}
func (m *mockMCPClient) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{}, nil
}
func (m *mockMCPClient) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{}, nil
}
func (m *mockMCPClient) Close() error {
	m.closeCalled = true
	return nil
}
func (m *mockMCPClient) Ping(ctx context.Context) error               { return nil }
func (m *mockMCPClient) ListResources(ctx context.Context, req mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListResourcesByPage(ctx context.Context, req mcp.ListResourcesRequest) (*mcp.ListResourcesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ReadResource(ctx context.Context, req mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return nil, nil
}
func (m *mockMCPClient) Subscribe(ctx context.Context, req mcp.SubscribeRequest) error {
	return nil
}
func (m *mockMCPClient) Unsubscribe(ctx context.Context, req mcp.UnsubscribeRequest) error {
	return nil
}
func (m *mockMCPClient) ListPrompts(ctx context.Context, req mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListPromptsByPage(ctx context.Context, req mcp.ListPromptsRequest) (*mcp.ListPromptsResult, error) {
	return nil, nil
}
func (m *mockMCPClient) GetPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListToolsByPage(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return nil, nil
}
func (m *mockMCPClient) SetLevel(ctx context.Context, req mcp.SetLevelRequest) error {
	return nil
}
func (m *mockMCPClient) ListResourceTemplates(ctx context.Context, req mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) ListResourceTemplatesByPage(ctx context.Context, req mcp.ListResourceTemplatesRequest) (*mcp.ListResourceTemplatesResult, error) {
	return nil, nil
}
func (m *mockMCPClient) Complete(ctx context.Context, req mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	return nil, nil
}
func (m *mockMCPClient) OnNotification(handler func(notification mcp.JSONRPCNotification)) {
}

var _ mcpclient.MCPClient = (*mockMCPClient)(nil)

// mockToolClient is a configurable mock for testing tool execution.
type mockToolClient struct {
	mockMCPClient
	callToolResult  *mcp.CallToolResult
	callToolErr     error
	listToolsResult *mcp.ListToolsResult
	listToolsErr    error
}

func (m *mockToolClient) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.callToolErr != nil {
		return nil, m.callToolErr
	}
	return m.callToolResult, nil
}

func (m *mockToolClient) ListTools(ctx context.Context, req mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	if m.listToolsErr != nil {
		return nil, m.listToolsErr
	}
	return m.listToolsResult, nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsAny(s string, candidates []string) bool {
	for _, c := range candidates {
		if contains(s, c) {
			return true
		}
	}
	return false
}
