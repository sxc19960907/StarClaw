# Mission Control Run Board Implementation Plan

## Steps

1. Start task.
2. Add Mission Control markup to Runs panel.
3. Add filtering state and render helpers in Web UI JS.
4. Add CSS for board and filters.
5. Extend smoke coverage.
6. Run:
   - `node --check internal/daemon/webui/assets/app.js`
   - `go test ./internal/daemon`
   - `git diff --check`
   - `./scripts/smoke_webui_core.sh`
   - `go test ./...`
