# Implementation Plan

## Checklist

- [x] Add `Test run` action to the agent editor.
- [x] Wire action to select agent and prefill Chat.
- [x] Update smoke coverage.
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
