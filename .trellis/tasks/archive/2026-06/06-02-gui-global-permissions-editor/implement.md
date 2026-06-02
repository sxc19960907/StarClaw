# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Extend backend config PATCH with permissions payload.
- [x] Add backend tests for permission save, refresh, and clear.
- [x] Convert GUI permissions panel into editor + preview.
- [x] Extend Web UI smoke for permissions save/clear.
- [x] Run syntax, Go, vet, smoke, and diff whitespace checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/daemon-webui-smoke.png'
```

## Rollback

Revert `internal/daemon/config_api.go`, permissions GUI changes, and smoke updates if permission persistence regresses.
