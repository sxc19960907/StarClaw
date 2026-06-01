# Implementation Plan

## Checklist

- [x] Inspect current Agents panel editor state and smoke script.
- [x] Add or refine permission controls in the Web UI.
- [x] Ensure load/save maps existing agent config to the controls.
- [x] Update smoke coverage for permission round trip.
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
