# MCP Config Editor Implementation Plan

## Steps

1. Activate the task with Trellis.
2. Extend config patch types with MCP server patching and validation.
3. Add backend tests for MCP save semantics and invalid input.
4. Add MCP Starport editor markup.
5. Add frontend state, rendering, save, edit, and disable handlers.
6. Add smoke coverage for add/edit/disable controls.
7. Run validation:
   - `gofmt -w internal/daemon/*.go`
   - `go test ./internal/daemon`
   - `node --check internal/daemon/webui/assets/app.js`
   - `git diff --check`
   - `./scripts/smoke_webui_core.sh`
   - `go test ./...`

## Review Gates

- `/config` response must not include any MCP env value.
- Blank env fields must not erase existing secrets.
- Invalid MCP config must return `400` with an actionable message.
- Existing provider and permissions config patch behavior must still pass tests.
