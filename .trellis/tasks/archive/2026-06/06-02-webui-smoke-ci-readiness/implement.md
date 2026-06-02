# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add configurable smoke artifact directory and metadata/log persistence.
- [x] Update GitHub CI to upload artifacts on Web UI core smoke failure.
- [x] Document Web UI smoke layers and CI behavior.
- [x] Run targeted checks.

## Validation Commands

```bash
scripts/smoke_webui_core.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*'
```
