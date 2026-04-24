// Package mcp provides MCP (Model Context Protocol) client functionality.
// This is a simplified implementation without external dependencies.
// For full functionality, integrate with github.com/mark3labs/mcp-go
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MCPServerConfig describes how to connect to an MCP server.
type MCPServerConfig struct {
	Command   string            `mapstructure:"command" yaml:"command" json:"command"`
	Args      []string          `mapstructure:"args,omitempty" yaml:"args,omitempty" json:"args,omitempty"`
	Env       map[string]string `mapstructure:"env,omitempty" yaml:"env,omitempty" json:"env,omitempty"`
	Type      string            `mapstructure:"type,omitempty" yaml:"type,omitempty" json:"type,omitempty"`        // "stdio" (default) or "http"
	URL       string            `mapstructure:"url,omitempty" yaml:"url,omitempty" json:"url,omitempty"`         // for http type
	Disabled  bool              `mapstructure:"disabled,omitempty" yaml:"disabled,omitempty" json:"disabled,omitempty"`    // skip this server
	Context   string            `mapstructure:"context,omitempty" yaml:"context,omitempty" json:"context,omitempty"`     // LLM context injected into system prompt
	KeepAlive bool              `mapstructure:"keep_alive,omitempty" yaml:"keep_alive,omitempty" json:"keep_alive,omitempty"` // stay connected between turns
}

// RemoteTool represents a tool discovered from an MCP server.
type RemoteTool struct {
	ServerName string
	Name       string
	Description string
	InputSchema json.RawMessage
}

// ToolCallResult represents the result of calling an MCP tool.
type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a block of content in the tool result.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCPClient interface abstracts the MCP client for testing.
type MCPClient interface {
	CallTool(ctx context.Context, toolName string, args json.RawMessage) (*ToolCallResult, error)
	ListTools(ctx context.Context) ([]RemoteTool, error)
	Close() error
}

// ClientManager manages connections to multiple MCP servers.
type ClientManager struct {
	mu          sync.Mutex
	clients     map[string]MCPClient
	configs     map[string]MCPServerConfig
	toolCache   map[string][]RemoteTool
	reconnectMu map[string]*sync.Mutex
}

// NewClientManager creates a new MCP client manager.
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:     make(map[string]MCPClient),
		configs:     make(map[string]MCPServerConfig),
		toolCache:   make(map[string][]RemoteTool),
		reconnectMu: make(map[string]*sync.Mutex),
	}
}

// ConnectAll connects to all configured MCP servers in parallel and returns discovered tools.
func (m *ClientManager) ConnectAll(ctx context.Context, servers map[string]MCPServerConfig) ([]RemoteTool, error) {
	type result struct {
		tools []RemoteTool
		err   error
		name  string
	}

	var wg sync.WaitGroup
	results := make(chan result, len(servers))

	for name, cfg := range servers {
		if cfg.Disabled {
			continue
		}
		wg.Add(1)
		go func(name string, cfg MCPServerConfig) {
			defer wg.Done()
			serverCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			tools, err := m.connect(serverCtx, name, cfg)
			results <- result{tools: tools, err: err, name: name}
		}(name, cfg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allTools []RemoteTool
	var hasErrors bool
	for r := range results {
		if r.err != nil {
			hasErrors = true
			// Log error but continue with other servers
			continue
		}
		allTools = append(allTools, r.tools...)
	}

	if hasErrors && len(allTools) == 0 {
		return nil, fmt.Errorf("failed to connect to any MCP server")
	}

	return allTools, nil
}

// connect establishes a connection to a single MCP server.
func (m *ClientManager) connect(ctx context.Context, name string, cfg MCPServerConfig) ([]RemoteTool, error) {
	m.mu.Lock()
	m.configs[name] = cfg
	m.mu.Unlock()

	// TODO: Implement actual MCP client connection
	// For now, return empty tools list (placeholder implementation)
	// In full implementation:
	// 1. Create stdio or HTTP client based on cfg.Type
	// 2. Initialize connection with timeout
	// 3. List available tools
	// 4. Cache tools and client

	return []RemoteTool{}, nil
}

// IsConnected returns true if connected to the given server.
func (m *ClientManager) IsConnected(serverName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.clients[serverName]
	return ok
}

// ConnectedServers returns a list of connected server names.
func (m *ClientManager) ConnectedServers() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	return names
}

// CallTool calls a tool on a specific MCP server.
func (m *ClientManager) CallTool(ctx context.Context, serverName, toolName string, args json.RawMessage) (*ToolCallResult, error) {
	m.mu.Lock()
	client, ok := m.clients[serverName]
	m.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("not connected to MCP server: %s", serverName)
	}

	// TODO: Implement actual tool call
	// return client.CallTool(ctx, toolName, args)
	_ = client
	return &ToolCallResult{
		Content: []ContentBlock{{Type: "text", Text: "MCP tool call not yet implemented"}},
	}, nil
}

// GetCachedTools returns the cached tools for a server.
func (m *ClientManager) GetCachedTools(serverName string) []RemoteTool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tools, ok := m.toolCache[serverName]; ok {
		return tools
	}
	return []RemoteTool{}
}

// Close closes all MCP connections.
func (m *ClientManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, client := range m.clients {
		if client != nil {
			client.Close()
		}
		delete(m.clients, name)
	}
	return nil
}
