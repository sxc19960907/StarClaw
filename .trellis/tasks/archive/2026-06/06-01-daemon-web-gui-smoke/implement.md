# Implementation Plan

## Checklist

- [x] Add `scripts/smoke_webui.sh`.
- [x] Make the script executable.
- [x] Ensure daemon start/stop cleanup is robust.
- [x] Add browser checks for render and schedule CRUD.
- [x] Add browser checks for approval UI behavior that do not need real LLM credentials.
- [x] Save screenshot under `output/playwright/`.
- [x] Document the smoke command.
- [x] Run validation commands.

## Validation Commands

```bash
scripts/smoke_webui.sh
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
```

## Rollback Points

- Revert `scripts/smoke_webui.sh` if the environment assumptions are too brittle.
- Revert docs-only changes independently if the script needs redesign.
