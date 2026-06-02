# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add a focused button-feedback helper.
- [x] Update the copy click handler to use the helper after successful copy.
- [x] Extend smoke assertions for transient label behavior.
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

Revert the app.js click-handler/helper change and smoke assertions if transient label timing makes smoke unreliable.
