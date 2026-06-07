# Workflow Recipes Launcher Implementation Plan

## Steps

1. Start task.
2. Add Home recipe markup to `index.html`.
3. Add `workflowRecipes`, render/selection helpers, and click delegation in `app.js`.
4. Add CSS for compact recipe buttons.
5. Extend Web UI smoke.
6. Run:
   - `node --check internal/daemon/webui/assets/app.js`
   - `go test ./internal/daemon`
   - `git diff --check`
   - `./scripts/smoke_webui_core.sh`
   - `go test ./...`

## Review Gates

- No backend or build pipeline changes.
- Existing Home actions continue to work.
- Recipe text is operational, not marketing copy.
