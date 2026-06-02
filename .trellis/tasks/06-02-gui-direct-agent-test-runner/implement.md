# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add direct agent test runner markup.
- [x] Add JS state/functions for running and rendering agent test output.
- [x] Update existing `Test run` editor action to populate the runner.
- [x] Extend Web UI smoke for the direct test runner.
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

Revert Web UI runner markup, JS changes, and smoke updates if the agent editor or chat flows regress.
