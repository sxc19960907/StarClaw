# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add search Clear button in `index.html`.
- [x] Add debounced session search input handling.
- [x] Add `Copy ID` session row action and delegated handler.
- [x] Extend smoke for copy id and clear search.
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

Revert the session search form, row action, handlers, and smoke additions if the sidebar behavior regresses.
