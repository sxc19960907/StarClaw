# Batch GUI chat composer polish

## Goal

Improve the Chat composer keyboard flow with a single batch of small, related GUI usability changes.

## Requirements

- `Cmd+Enter` on macOS and `Ctrl+Enter` on other keyboards should submit the chat form from the prompt textarea.
- `Escape` should stop an active run when the chat input has focus.
- After a run finishes, fails, or is cancelled, the prompt textarea should regain focus so the next message can be typed immediately.
- Existing click-based `Send` and `Stop` behavior must remain unchanged.
- Existing run summary behavior must remain unchanged.
- No backend API changes.

## Acceptance Criteria

- [ ] `Cmd/Ctrl+Enter` submits the current prompt and renders the run summary.
- [ ] `Escape` while a run is active triggers the existing cancellation flow.
- [ ] After completion/cancellation/error cleanup, the prompt textarea is focused.
- [ ] Existing browser smoke coverage still passes and includes the keyboard submit flow.

## Notes

- This batch intentionally groups related Chat composer polish items to reduce planning and commit overhead.
