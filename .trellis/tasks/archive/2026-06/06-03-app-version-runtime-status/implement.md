# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before editing.
- [x] Inspect Version API and Web UI renderVersion paths.
- [x] Extend Version API response with runtime context fields.
- [x] Update Version page rendering to include runtime context.
- [x] Add/update daemon API tests and Web UI smoke assertions.
- [x] Run targeted checks and smoke.

## Validation Commands

```bash
go test ./internal/daemon
scripts/smoke_webui_core.sh
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*' ':(exclude)output/manual-gui/*'
```
