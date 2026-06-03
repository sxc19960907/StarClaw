# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before editing.
- [x] Inspect `starclaw app`, daemon start, health check, and version command code paths.
- [x] Identify current behavior gaps for already-running daemon and startup failures.
- [x] Implement focused CLI/startup improvements.
- [x] Add or update tests/smoke coverage.
- [x] Run targeted checks and full relevant tests.

## Validation Commands

```bash
go test ./cmd ./internal/daemon ./internal/update
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*' ':(exclude)output/manual-gui/*'
```
