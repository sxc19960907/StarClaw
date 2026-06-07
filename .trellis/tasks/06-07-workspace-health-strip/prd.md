# Workspace health strip

## Goal

Add a compact Home Workspace Health Strip that summarizes Astria readiness from diagnostics, permissions, MCP docks, and memory warnings with direct navigation actions.

## Requirements

- Reuse existing diagnostics, permissions, config/MCP, and memory state; do not add backend endpoints.
- Render a compact health strip on Home near Focus Brief / Workspace Hub.
- Show diagnostics readiness, permissions mode, MCP dock status, and memory warning/fact status.
- Each health item should navigate to the relevant panel.
- Refresh the strip when diagnostics, permissions, MCP/config, or memory data changes.
- Keep styling dense, operational, and consistent with Astria.
- Do not add frontend dependencies or a build pipeline.

## Acceptance Criteria

- [x] Home renders a Workspace Health Strip.
- [x] Strip shows diagnostics, permissions, MCP, and memory items.
- [x] Strip updates from existing loaded state.
- [x] Clicking an item navigates to the relevant panel.
- [x] Core smoke verifies strip rendering and one navigation action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No backend readiness aggregation endpoint.
- No persistent health preferences.
- No replacement of Diagnostics, Permissions, MCP, or Memory panels.
