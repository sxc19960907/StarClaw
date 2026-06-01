# Implementation Plan

## Checklist

- [x] Add clear/cancel command button to the form.
- [x] Allow command rename in frontend state.
- [x] Update smoke coverage for rename and clear behavior.
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
