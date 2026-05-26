# MCP Serve Hardening Implementation Plan

## Checklist

- [x] Load config in `mcpServeCmd`.
- [x] Pass `cfg.Tools` to `RegisterLocalTools`.
- [x] Pass MCP server timeout and expose list into `NewMCPServer`.
- [x] Track exposed MCP tool count in `MCPServer`.
- [x] Wrap MCP tool handler execution in `context.WithTimeout` when configured.
- [x] Add tests for expose filtering, exposed server info count, and timeout behavior.
- [x] Run targeted tests: `go test ./internal/tools ./cmd`.
- [x] Run full tests: `go test ./...`.
- [x] Run `git diff --check`.

## Risk Points

- `mcpserver.ServeStdio` owns stdio serving, so command-level tests should not invoke `mcp serve` directly unless the serve path is injectable.
- Timeout tests should use a mock tool that observes context cancellation to avoid slow sleeps.
- Keep MCP handler errors as tool results to preserve existing behavior.
