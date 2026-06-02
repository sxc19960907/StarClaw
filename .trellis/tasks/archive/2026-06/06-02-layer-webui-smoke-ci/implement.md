# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add shared Web UI smoke harness.
- [x] Add focused layer scripts.
- [x] Refactor full smoke to use the shared harness.
- [x] Add core Web UI smoke to GitHub CI.
- [x] Run each smoke layer, full smoke, Go tests, vet, and diff check.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
scripts/smoke_webui_core.sh
scripts/smoke_webui_permissions.sh
scripts/smoke_webui_agents.sh
scripts/smoke_webui_runs.sh
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*.png'
```

## Rollback

Revert smoke script refactor and CI workflow changes if the layer scripts are less reliable than the current full smoke entrypoint.
