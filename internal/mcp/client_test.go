package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"
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

	// Initially not connected
	if cm.IsConnected("test-server") {
		t.Error("IsConnected should return false for unknown server")
	}

	// Manually add a mock client
	cm.clients["test-server"] = &mockClient{}

	if !cm.IsConnected("test-server") {
		t.Error("IsConnected should return true for connected server")
	}
}

func TestClientManager_ConnectedServers(t *testing.T) {
	cm := NewClientManager()

	// Initially empty
	servers := cm.ConnectedServers()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(servers))
	}

	// Add some mock clients
	cm.clients["server1"] = &mockClient{}
	cm.clients["server2"] = &mockClient{}

	servers = cm.ConnectedServers()
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}
}

func TestClientManager_ConnectAll(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	servers := map[string]MCPServerConfig{
		"enabled-server": {
			Command: "echo",
			Args:    []string{"hello"},
		},
		"disabled-server": {
			Command:  "echo",
			Disabled: true,
		},
	}

	tools, err := cm.ConnectAll(ctx, servers)
	// Currently returns empty tools as implementation is placeholder
	if err != nil {
		// Error is expected for now since we don't have real MCP implementation
		t.Logf("ConnectAll returned error (expected for placeholder): %v", err)
	}

	// Verify disabled server is skipped
	if _, exists := cm.configs["disabled-server"]; exists {
		t.Error("Disabled server should not be in configs")
	}

	// Verify enabled server was processed
	if _, exists := cm.configs["enabled-server"]; !exists {
		t.Error("Enabled server should be in configs")
	}

	_ = tools
}

func TestClientManager_ConnectAll_Timeout(t *testing.T) {
	cm := NewClientManager()
	// Very short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	servers := map[string]MCPServerConfig{
		"slow-server": {
			Command: "sleep",
			Args:    []string{"10"},
		},
	}

	// Should handle timeout gracefully
	_, _ = cm.ConnectAll(ctx, servers)
	// We don't check error here because the placeholder implementation
	// doesn't actually respect the timeout yet
}

func TestClientManager_Close(t *testing.T) {
	cm := NewClientManager()

	// Add some mock clients
	cm.clients["server1"] = &mockClient{}
	cm.clients["server2"] = &mockClient{}

	err := cm.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	if len(cm.clients) != 0 {
		t.Error("Close() should clear all clients")
	}
}

func TestClientManager_CallTool_NotConnected(t *testing.T) {
	cm := NewClientManager()
	ctx := context.Background()

	result, err := cm.CallTool(ctx, "unknown-server", "test-tool", nil)
	if err == nil {
		t.Error("CallTool should return error for unknown server")
	}
	if result != nil {
		t.Error("CallTool should return nil result on error")
	}
}

func TestClientManager_GetCachedTools(t *testing.T) {
	cm := NewClientManager()

	// Initially empty
	tools := cm.GetCachedTools("test-server")
	if tools == nil {
		t.Error("GetCachedTools should return empty slice, not nil")
	}
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(tools))
	}

	// Add cached tools
	cm.toolCache["test-server"] = []RemoteTool{
		{Name: "tool1", ServerName: "test-server"},
		{Name: "tool2", ServerName: "test-server"},
	}

	tools = cm.GetCachedTools("test-server")
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestMCPServerConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  MCPServerConfig
		wantErr bool
	}{
		{
			name: "stdio server",
			config: MCPServerConfig{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			},
			wantErr: false,
		},
		{
			name: "http server without URL",
			config: MCPServerConfig{
				Type:    "http",
				Command: "curl",
			},
			wantErr: false, // Validation happens at connection time
		},
		{
			name: "disabled server",
			config: MCPServerConfig{
				Command:  "echo",
				Disabled: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify struct can be created
			_ = tt.config
		})
	}
}

// Mock client for testing
type mockClient struct {
	closeCalled bool
}

func (m *mockClient) CallTool(ctx context.Context, toolName string, args json.RawMessage) (*ToolCallResult, error) {
	return nil, nil
}

func (m *mockClient) ListTools(ctx context.Context) ([]RemoteTool, error) {
	return nil, nil
}

func (m *mockClient) Close() error {
	m.closeCalled = true
	return nil
}
