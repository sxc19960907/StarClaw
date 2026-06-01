# Implementation Plan

## Checklist

- [x] Add heartbeat fields to daemon agent edit request and config builder.
- [x] Extend backend tests for create/update/clear heartbeat.
- [x] Add Heartbeat controls to the Agents editor.
- [x] Wire frontend load/save for heartbeat fields.
- [x] Update smoke coverage for heartbeat round trip.
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
