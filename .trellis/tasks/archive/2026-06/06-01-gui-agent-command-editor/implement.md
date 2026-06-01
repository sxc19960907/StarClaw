# Implementation Plan

## Checklist

- [x] Add optional command payload handling and safe command-name validation.
- [x] Add backend tests for create/update/delete/preserve/invalid command names.
- [x] Add command editor controls to the Agents panel.
- [x] Wire frontend load/save/delete for commands.
- [x] Update smoke coverage for command round trip.
- [x] Run validation commands.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/daemon-webui-smoke.png'
```
