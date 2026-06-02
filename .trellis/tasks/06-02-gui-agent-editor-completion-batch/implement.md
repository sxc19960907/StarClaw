# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add editor controls and preview/import/export markup.
- [x] Add dirty-state tracking and guarded action helper.
- [x] Add command editor New/Cancel behavior.
- [x] Add agent export and import handlers.
- [x] Add live permission preview rendering.
- [x] Extend Web UI smoke for the module workflow.
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

Revert the Web UI asset changes and smoke additions if the combined editor workflow creates regressions in existing agent create/update/delete behavior.
