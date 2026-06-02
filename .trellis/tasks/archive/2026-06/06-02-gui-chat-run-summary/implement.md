# Implementation Plan

## Checklist

- [x] Add run summary rendering helper.
- [x] Call helper after successful chat completion.
- [x] Add CSS for the summary card.
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
