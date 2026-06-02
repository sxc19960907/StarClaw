# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add Chat input keydown handler for submit and cancel shortcuts.
- [x] Refocus the Chat input in `submitChat` cleanup.
- [x] Update smoke to cover keyboard submit and post-run focus.
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

Revert the input keydown handler, focus change, and smoke updates if keyboard handling conflicts with textarea behavior.
