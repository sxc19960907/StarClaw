# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add daemon version/update-check handlers.
- [x] Register routes and add backend tests.
- [x] Add Web UI Version panel, state loading, and check button.
- [x] Extend core Web UI smoke coverage.
- [x] Run syntax, Go, vet, smoke, and diff checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd ./internal/update
go test ./...
go vet ./...
scripts/smoke_webui_core.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*.png'
```
