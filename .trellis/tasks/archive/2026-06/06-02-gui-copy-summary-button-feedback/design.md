# Design

## Boundary

This task changes the daemon Web UI assets and smoke coverage only:

- `internal/daemon/webui/assets/app.js`
- `scripts/smoke_webui.sh`

No backend contract changes are required.

## UI Behavior

The existing delegated `data-run-summary-copy` handler will pass the clicked button to a small feedback helper. After clipboard success, the helper stores the original label, changes the button text to `Copied`, disables it briefly, then restores the original label and enabled state.

Toast feedback remains unchanged.

## Compatibility

Clipboard failure behavior remains the existing toast error path. The button label is changed only after `navigator.clipboard.writeText` succeeds.

## Test Strategy

The Web UI smoke test already clicks `Copy summary` and verifies clipboard contents. Extend it to assert the button shows `Copied` and then returns to `Copy summary`.
