# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add release readiness card to Version render.
- [x] Update core smoke assertions.
- [x] Run targeted checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
scripts/smoke_webui_core.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*'
```
