# Implementation Plan

## Checklist

- [x] Add `PATCH /sessions/{id}` route and handler.
- [x] Add backend tests for session title/favorite patch validation.
- [x] Make `GET /permissions` return real read-only config summary.
- [x] Add backend test for permissions overview.
- [x] Add Permissions nav/panel in Web UI.
- [x] Add session rename/favorite/delete-confirm controls.
- [x] Add active session state display in composer/chat area.
- [x] Improve activity/tool/approval/error styling and labels.
- [x] Add diagnostics action routing to Config/Permissions.
- [x] Update smoke script for the new GUI flows.
- [x] Run validation commands.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/daemon-webui-smoke.png'
```

## Risky Files

- `internal/daemon/router.go`
- `internal/daemon/server.go`
- `internal/daemon/server_test.go`
- `internal/daemon/webui/index.html`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `scripts/smoke_webui.sh`

## Rollback Points

- Revert session PATCH independently if metadata updates regress.
- Revert Permissions panel independently because it is read-only.
- Revert frontend polish without affecting daemon APIs.
