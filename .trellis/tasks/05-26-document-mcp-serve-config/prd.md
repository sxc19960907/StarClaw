# Document MCP serve config

## Goal

Make StarClaw's MCP stdio server mode discoverable and safely configurable from the user-facing docs.

## Requirements

- Document `starclaw mcp serve` in `README.md`.
- Document `tools.server_tool_timeout` and `tools.mcp_expose` in `docs/CONFIGURATION.md`.
- Explain that an empty `mcp_expose` list exposes all registered local tools.
- Include a safer allow-list example for read-mostly MCP server usage.
- Keep the existing MCP client docs intact.

## Acceptance Criteria

- [ ] README includes a command example for serving StarClaw tools over MCP stdio.
- [ ] README includes an example `tools.mcp_expose` allow-list and timeout.
- [ ] Configuration guide explains field behavior and defaults.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Notes

Out of scope:

- Changing MCP server runtime behavior.
- Adding generated Claude Desktop configuration.
