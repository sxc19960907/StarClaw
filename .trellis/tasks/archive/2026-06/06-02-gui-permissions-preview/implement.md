# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add pending permissions preview markup.
- [x] Render permissions preview and risk hints from form payload.
- [x] Refresh preview on load, edits, clear, and save.
- [x] Extend agent permission preview warnings.
- [x] Update Web UI smoke coverage.
- [x] Run targeted checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
scripts/smoke_webui_permissions.sh
scripts/smoke_webui_agents.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*'
```
