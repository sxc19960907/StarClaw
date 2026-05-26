# Harden MCP serve

## Goal

Make `starclaw mcp serve` usable as a configurable MCP stdio tool server by honoring existing StarClaw tool exposure and server timeout settings.

This lets users expose StarClaw as a general-purpose local agent/tool server without accidentally publishing every local tool or allowing unbounded tool execution.

## Confirmed Facts

- `tools.server_tool_timeout` and `tools.mcp_expose` already exist in `config.ToolsConfig`.
- `tools.NewMCPServer` already accepts `MCPServerConfig{ToolTimeout, ExposeTools}`.
- `mcp serve` currently registers local tools without loading config and passes an empty MCP server config.
- MCP server tool handlers currently delegate directly to `tool.Run(ctx, argsJSON)` without enforcing `ToolTimeout`.
- An empty expose list should preserve the existing all-tools behavior.

## Requirements

- `starclaw mcp serve` must load StarClaw config before registering local tools.
- Local tool registration in `mcp serve` must receive `cfg.Tools` so tool-level config remains consistent with chat/TUI usage.
- `tools.mcp_expose` must restrict which registered tools are added to the MCP server when non-empty.
- `tools.server_tool_timeout` must limit individual MCP tool calls when greater than zero.
- Timeout failures must be reported as MCP tool error results, not process crashes or handler-level errors.
- Existing tool schema, required args, descriptions, and read-only annotations must remain preserved.
- `ServerInfo().tool_count` must reflect exposed MCP tools, not the unfiltered registry size.

## Acceptance Criteria

- [ ] `mcp serve` passes `cfg.Tools.ServerToolTimeout` and `cfg.Tools.MCPExpose` into `tools.NewMCPServer`.
- [ ] MCP server tests cover expose filtering and exposed tool count.
- [ ] MCP server tests cover timeout behavior for a slow tool.
- [ ] Existing MCP server execution and schema tests continue passing.
- [ ] `go test ./internal/tools ./cmd` passes.
- [ ] `go test ./...` passes or any unrelated pre-existing failure is documented.

## Notes

Out of scope:

- MCP authentication or approval prompts.
- HTTP/SSE MCP serving.
- Claude Desktop config generation.
- Changing remote MCP client behavior.
