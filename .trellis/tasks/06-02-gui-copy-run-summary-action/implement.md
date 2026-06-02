# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add a small clipboard helper using existing toast feedback.
- [x] Render `Copy summary` in `renderRunSummary`.
- [x] Handle `data-run-summary-copy` in the delegated click handler.
- [x] Reuse existing run summary actions styling.
- [x] Extend Web UI smoke to validate copy behavior.
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

Revert changes to the scoped Web UI assets and smoke assertion if clipboard behavior is not reliable in the browser smoke environment.
