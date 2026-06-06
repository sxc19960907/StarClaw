# File Intake UI Implementation Plan

## Steps

1. Start the Trellis task.
2. Add `intake_api.go` with `POST /intake/file`.
3. Register route and add backend tests.
4. Add File Intake nav/home/manage UI.
5. Add frontend state, render, analyze, and chat prompt helpers.
6. Extend Web UI smoke with a temp text/doc path and archive path check.
7. Run validation:
   - `gofmt -w internal/daemon/*.go`
   - `go test ./internal/daemon`
   - `node --check internal/daemon/webui/assets/app.js`
   - `git diff --check`
   - `./scripts/smoke_webui_core.sh`
   - `go test ./...`

## Review Gates

- The direct intake endpoint is read-only.
- Archive extraction remains routed through normal chat/run approval.
- Missing files and unsupported formats produce visible actionable errors.
- UI remains consistent with the Astria/Kocoro-like shell.
