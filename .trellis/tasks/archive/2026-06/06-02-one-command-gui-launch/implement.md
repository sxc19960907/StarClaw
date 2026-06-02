# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add daemon ensure/start/open helpers.
- [x] Add top-level `app` command and `daemon open --start`.
- [x] Add CLI unit tests.
- [x] Update README/docs and CLI smoke.
- [x] Run targeted tests, full tests, vet, smoke, and diff check.

## Validation Commands

```bash
go test ./cmd
go test ./...
go vet ./...
scripts/smoke_cli.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*.png'
```
