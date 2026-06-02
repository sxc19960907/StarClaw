# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Run full Web UI smoke.
- [x] Fix any smoke blockers.
- [x] Re-run targeted/full smoke.
- [x] Run diff check.

## Validation Commands

```bash
scripts/smoke_webui.sh
node --check internal/daemon/webui/assets/app.js
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*'
```
