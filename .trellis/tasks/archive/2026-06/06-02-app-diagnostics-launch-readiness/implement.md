# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Extend diagnostics response with launch/path metadata.
- [x] Extend version response with launch command.
- [x] Render Launch readiness in Diagnostics and Version GUI panels.
- [x] Improve app launch failure hint without changing success behavior.
- [x] Extend API/unit and Web UI smoke coverage.
- [x] Run targeted checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
scripts/smoke_cli.sh
scripts/smoke_webui_core.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*.png'
```
