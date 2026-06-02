# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add Run detail action toolbar.
- [x] Add copy/open/rerun click handlers.
- [x] Replace raw timeline rendering with grouped timeline entries.
- [x] Extend runs smoke coverage for actions and grouped timeline.
- [x] Run JS syntax check, targeted Go tests, Web UI smoke, and diff check.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
scripts/smoke_webui_runs.sh
scripts/smoke_webui_agents.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*.png'
```
