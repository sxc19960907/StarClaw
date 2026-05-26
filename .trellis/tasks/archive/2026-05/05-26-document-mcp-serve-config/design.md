# Document MCP Serve Config Design

## Boundary

Docs-only task:

- `README.md`
- `docs/CONFIGURATION.md`

## Content Shape

README should stay quick-start oriented:

- Keep existing MCP client section.
- Add a separate MCP server section with `starclaw mcp serve`.
- Include a concise safe expose-list example.

Configuration guide should provide field-level details:

- `tools.server_tool_timeout`: seconds, `0` disables MCP server timeout.
- `tools.mcp_expose`: optional allow-list for tools exposed by `starclaw mcp serve`; empty means all registered local tools.
- Clarify this is distinct from agent-side `tools.allowed` / `tools.denied`.
