# Implementation Plan

## Checklist

- [x] Add diagnostics response/check types and builder.
- [x] Register `GET /diagnostics`.
- [x] Add backend tests for ready, needs_setup, and route registration.
- [x] Add Web UI diagnostics state, nav item, topbar chip, and detail panel.
- [x] Update Web UI smoke script to assert diagnostics render.
- [x] Update docs if needed.
- [x] Run validation commands.

## Validation Commands

```bash
scripts/smoke_webui.sh
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
```

## Risky Files

- `internal/daemon/router.go`
- `internal/daemon/server.go`
- `internal/daemon/webui/index.html`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `scripts/smoke_webui.sh`

## Rollback Points

- Revert diagnostics endpoint independently if frontend rendering remains usable.
- Revert frontend panel independently if backend endpoint works but UI needs redesign.
