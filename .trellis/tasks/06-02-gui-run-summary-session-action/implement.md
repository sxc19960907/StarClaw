# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add the run summary action markup in `renderRunSummary`.
- [x] Add a delegated click handler that opens the selected summary session.
- [x] Add CSS for the summary action row/button using existing visual patterns.
- [x] Extend `scripts/smoke_webui.sh` to assert the action is visible.
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

Revert changes to the three scoped files and restore the smoke assertions if the action causes session selection regressions.
