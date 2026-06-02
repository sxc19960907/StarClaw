# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Keep agent test completion in-place.
- [x] Add prompt/request metadata and copy summary action to result card.
- [x] Add contextual agent test error card.
- [x] Update browser smoke assertions.
- [x] Run targeted checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
scripts/smoke_webui_agents.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*'
```
