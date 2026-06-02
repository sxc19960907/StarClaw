# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add `app --check` readiness output.
- [x] Add `app --no-open` daemon launch mode.
- [x] Add CLI tests and smoke coverage.
- [x] Update install/usage docs.
- [x] Run targeted checks.

## Validation Commands

```bash
go test ./cmd
scripts/smoke_cli.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*'
```
