# Implementation Plan

## Checklist

- [x] Add agent create/update request type and persistence helpers.
- [x] Implement `handleCreateAgent` and `handleUpdateAgent`.
- [x] Add backend tests for create/update/validation.
- [x] Add Agents panel editor form and state.
- [x] Wire create/edit/save/delete actions.
- [x] Update smoke script for agent editor flow.
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
