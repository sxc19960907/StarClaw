# Implementation Plan

## Checklist

1. Inspect existing Web UI stream rendering, run summary, cancel, and run detail
   code paths.
2. Add a compact live runtime status component to the Chat panel.
3. Update the per-run SSE renderer to maintain live status for:
   - stream state,
   - run id/session id,
   - usage,
   - latest tool/control event,
   - cancel/error/final completion.
4. Keep final run summary rendering compatible with current smoke tests.
5. Extend Web UI streaming smoke to assert the live status appears during a
   streaming run and updates with usage/final state.
6. Re-run:
   - `scripts/smoke_webui_streaming.sh`
   - `scripts/smoke_webui_core.sh`
   - `go test ./internal/daemon -count=1`
   - `go test ./...`
   - `git diff --check`
   - Trellis task validate.

## Risky Files

- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `internal/daemon/webui/index.html`
- `scripts/lib/webui_smoke_common.sh`

## Rollback Points

- If live status rendering causes smoke flakiness, revert the live status
  component and keep existing streaming summary behavior.
- If daemon contracts need changes, stop and update `design.md` before editing
  backend handlers.
