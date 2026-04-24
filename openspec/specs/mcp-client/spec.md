# Specification: MCP Client Support

## Overview

| Field | Value |
|-------|-------|
| Feature | MCP Client Integration |
| Type | Core Feature |
| Optional | Yes (no MCP servers = no MCP tools) |
| Default | Enabled but empty |

## Purpose

Enable StarClaw to connect to Model Context Protocol (MCP) servers and use their tools alongside built-in tools. This allows extensibility without modifying StarClaw's core code.

## Configuration

### MCPServerConfig Schema

```yaml
mcp_servers:
  <server_name>:
    command: string              # Required: executable to run
    args: [string]               # Optional: arguments
    env: {string: string}        # Optional: environment variables
    type: string                 # Optional: "stdio" (default) or "http"
    url: string                  # Optional: for http type
    disabled: bool               # Optional: skip this server
    context: string              # Optional: context for LLM
    keep_alive: bool             # Optional: stay connected between turns
```

### Example Configuration

```yaml
# ~/.starclaw/config.yaml
mcp_servers:
  github:
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: ${GITHUB_TOKEN}
    keep_alive: true
  
  filesystem:
    command: npx
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/Users/timmy/workspace"]
  
  fetch:
    command: uvx
    args: ["mcp-server-fetch"]
```

### Environment Variable Substitution

Environment variables in `env` values are expanded at runtime:
- `${VAR}` or `$VAR` - substitute variable
- Missing variables are left as-is (no error)

## Tool Priority

Tools are registered in priority order (highest priority wins on name collision):

1. Local tools (built-in)
2. MCP tools (from configured servers)

## ClientManager API

### Methods

```go
// Create new manager
func NewClientManager() *ClientManager

// Connect to all configured servers
func (m *ClientManager) ConnectAll(ctx context.Context, servers map[string]MCPServerConfig) ([]RemoteTool, error)

// Call a tool on a specific server
func (m *ClientManager) CallTool(ctx context.Context, serverName, toolName string, args json.RawMessage) (*ToolCallResult, error)

// Check if connected to server
func (m *ClientManager) IsConnected(serverName string) bool

// Get connected server names
func (m *ClientManager) ConnectedServers() []string

// Close all connections
func (m *ClientManager) Close() error
```

## Error Handling

| Error Type | Behavior |
|------------|----------|
| Server start failure | Log error, continue with other servers |
| Connection lost | Auto-reconnect on next tool call |
| Tool call timeout | Return timeout error after 60s |
| Invalid arguments | Return validation error from server |
| Server crash | Mark disconnected, attempt reconnect |

## Testing Requirements

### Unit Tests

- Mock MCP client for isolated testing
- Test connection management
- Test concurrent tool calls
- Test error scenarios

### Integration Tests

- Test with real MCP servers (filesystem, fetch)
- Test reconnection after server restart
- Test timeout handling

## Security Considerations

- MCP servers run as separate processes with same user permissions
- Environment variables may contain secrets (tokens)
- Tool arguments are logged (with redaction for sensitive data)
- No network sandboxing (MCP server controls its own network access)

## References

- MCP Protocol: https://modelcontextprotocol.io/
- mcp-go SDK: https://github.com/mark3labs/mcp-go
- ShanClaw Reference: `/Users/timmy/PycharmProjects/ShanClaw/internal/mcp/client.go`
