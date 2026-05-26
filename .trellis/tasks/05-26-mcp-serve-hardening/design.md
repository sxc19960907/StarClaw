# MCP Serve Hardening Design

## Boundaries

- CLI wiring lives in `cmd/root.go` under `mcpServeCmd`.
- MCP server behavior lives in `internal/tools/mcp_server.go`.
- Tool registry construction remains owned by `internal/tools/register.go`.
- No new configuration fields are introduced; this task activates existing fields only.

## Data Flow

1. `starclaw mcp serve` loads config with `config.Load()`.
2. The command registers local tools with `tools.RegisterLocalTools(cfg.Tools)`.
3. The command registers the version tool.
4. The command creates the MCP server with:
   - `ToolTimeout: cfg.Tools.ServerToolTimeout`
   - `ExposeTools: cfg.Tools.MCPExpose`
5. `NewMCPServer` filters registry tools during MCP registration.
6. Each MCP handler marshals MCP arguments into StarClaw tool JSON and executes with the caller context, optionally wrapped in `context.WithTimeout`.

## Contracts

- `MCPServerConfig.ExposeTools == nil` or empty means expose every registered tool.
- `MCPServerConfig.ToolTimeout <= 0` means no extra MCP server timeout.
- Handler errors returned by StarClaw tools become `mcp.CallToolResult{IsError: true}`.
- Context deadline failures are treated the same way: a tool error result with useful text.
- `ServerInfo().tool_count` reports the number of tools actually registered with the MCP server.

## Compatibility

- Default config remains backward-compatible: all tools exposed, no extra MCP timeout unless configured.
- Existing MCP clients still see the same schema for exposed tools.
- No changes are made to local chat/TUI permission behavior.

## Rollback

The change is isolated to MCP serve wiring and MCP server handler construction. Reverting those edits restores prior all-tools/no-timeout behavior.
