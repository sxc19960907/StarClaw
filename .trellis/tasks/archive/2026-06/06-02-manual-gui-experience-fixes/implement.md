# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Fix chat readiness copy.
- [x] Disable unsupported update checks.
- [x] Improve permissions preview layout.
- [x] Improve agent editor/Test Runner layout.
- [x] Clear toast on navigation.
- [x] Run targeted checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
scripts/smoke_webui_core.sh
scripts/smoke_webui_permissions.sh
scripts/smoke_webui_agents.sh
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*' ':(exclude)output/manual-gui/*'
```
