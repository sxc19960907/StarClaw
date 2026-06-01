# Implementation Plan

## Checklist

- [x] Add YAML-aware config read/patch helpers in daemon code.
- [x] Redact API keys from `GET /config` and expose key-present booleans.
- [x] Validate and merge provider-level config patch fields.
- [x] Reload `ServerDeps.Config` after successful patch.
- [x] Add backend tests for YAML get, patch, in-memory refresh, and secret preservation.
- [x] Add Web UI Config nav/panel and provider setup form.
- [x] Add Diagnostics path/action to open Config.
- [x] Update smoke script to verify config repair render/save.
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

- `internal/daemon/server.go`
- `internal/daemon/server_test.go`
- `internal/daemon/webui/index.html`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `scripts/smoke_webui.sh`

## Rollback Points

- Revert backend `/config` behavior if YAML handling causes regressions.
- Revert GUI config panel independently if backend tests pass but UI flow needs redesign.
