# MCP Config Editor

## Goal

Allow users to add, edit, disable, and test MCP server configuration from Astria while preserving safe config handling and secret redaction.

## Requirements

- Support editing core MCP fields already represented in config.
- Preserve existing secrets unless explicitly replaced.
- Validate required fields before saving.
- Allow disabling rather than deleting for rollback.
- Keep connection testing available after edits.

## Acceptance Criteria

- [x] User can add a stdio MCP server from the UI.
- [x] User can edit and disable an existing MCP server.
- [x] Blank secret/env values are preserved where applicable.
- [x] Invalid config returns actionable UI errors.
- [x] Server tests cover config patching and redaction.
- [x] Web UI smoke or targeted test covers add/edit/disable.

## Non-Goals

- No marketplace/discovery catalog.
- No remote secret manager.
- No automatic tool execution.

## Dependencies

- Depends on MCP Starport UI and config API patterns.
