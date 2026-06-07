# MCP capability inspector

## Goal

Add a Home MCP capability inspector that summarizes tool docks, transports, env keys, and readiness actions.

## Requirements

- Reuse existing Web UI config/MCP state only; do not add backend endpoints.
- Add a Home "Tool Dock Inspector" / MCP capability console that summarizes configured MCP docks.
- Surface total docks, enabled/disabled state, transport types, env key counts, keep-alive/context flags, and no-dock state.
- Each item should route to MCP Starport or existing MCP edit/test actions where available.
- Refresh when config/MCP data changes.
- Keep styling dense and operational with Astria's subtle celestial identity.

## Acceptance Criteria

- [x] Home renders a Tool Dock Inspector / MCP capability console.
- [x] Console derives items from existing `state.config.mcp_servers`.
- [x] Items navigate to MCP Starport or existing MCP actions.
- [x] Empty/no-dock state is explicit.
- [x] Core smoke verifies rendering and one navigation action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new MCP backend API.
- No new transport implementation.
- No automatic external tool execution.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
