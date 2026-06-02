# Design

## Boundary

This batch changes only Chat composer Web UI behavior and smoke coverage:

- `internal/daemon/webui/assets/app.js`
- `scripts/smoke_webui.sh`

No daemon API changes are required.

## Behavior

Add a `handleChatInputKeydown` handler for `#chat-input`:

- `Meta+Enter` or `Ctrl+Enter`: prevent default and request form submission if no run is active.
- `Escape`: prevent default and call the existing `cancelActiveRun` only when a run is active.

Use `requestSubmit()` so keyboard submit uses the same `submitChat` validation path as the button.

In `submitChat` cleanup, focus the input after controls are re-enabled. This applies to success, cancel, and error paths because they all share the `finally` block.

## Compatibility

Click submit/stop paths stay unchanged. Plain `Enter` remains normal textarea input behavior.

## Test Strategy

Extend Web UI smoke to fill the chat textarea and submit with keyboard shortcut instead of the submit button for the mocked run summary path. Existing approval smoke already exercises `Stop` visibility indirectly; this batch will add direct Escape cancellation only if it can be done without adding brittle timing.
