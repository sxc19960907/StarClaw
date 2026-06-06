# MCP Starport Management UI

## Goal

Make MCP servers a first-class Astria product surface. Users should be able to see configured servers, inspect health and tools, test connections, and understand setup issues from the Web UI.

## Requirements

- Add an MCP management panel or upgrade the existing MCP surface.
- Display configured MCP servers, transport type, enabled state, status, and last error.
- Show discovered tools/resources for a server when available.
- Provide a connection test action with clear success/failure feedback.
- Support enable/disable and edit flows only if matching backend/config APIs exist or are added in this task.
- Preserve manual config compatibility; UI should not corrupt existing MCP config.
- Styling should use Astria "Starport" language subtly while staying functional.

## Acceptance Criteria

- [x] MCP server list renders from real config/state.
- [x] Connection test reports success/failure without crashing the daemon UI.
- [x] Tool/resource discovery is visible for at least one configured server in tests or smoke fixtures.
- [x] Invalid server config is displayed with actionable error state.
- [x] Existing MCP CLI/client behavior remains unchanged.
- [x] Web UI and relevant backend tests pass.

## Non-Goals

- No MCP server implementation in this task.
- No marketplace/catalog of third-party MCP servers.
- No secret manager beyond existing config/env patterns.

## Dependencies

- Depends on existing MCP client/config contracts.
- May require a design document before implementation because it crosses UI, daemon API, and config persistence.
