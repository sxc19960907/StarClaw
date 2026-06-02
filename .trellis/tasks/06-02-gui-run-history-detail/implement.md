# Implementation Plan

## Checklist

- [x] Read applicable Trellis specs before code edits.
- [x] Add run record/store types and bounded in-memory storage.
- [x] Wire run recording into `handleMessage` sync and SSE paths.
- [x] Add `GET /runs` and `GET /runs/{id}` routes and tests.
- [x] Add GUI Runs nav/panel/list/detail.
- [x] Add `Open run` action to chat run summary.
- [x] Extend Web UI smoke for run history/detail.
- [x] Run syntax, Go, vet, smoke, and diff whitespace checks.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
scripts/smoke_webui.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/daemon-webui-smoke.png'
```

## Rollback

Revert run store/routes and frontend run panel changes if the API or smoke behavior regresses existing chat/session flows.
