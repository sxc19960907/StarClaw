# Stop wait tool timer

## Goal

Make the `wait` tool release timer resources promptly when its context is cancelled.

This keeps the agent's preferred waiting primitive lightweight even when waits are cancelled before their configured duration.

## Requirements

- Replace `time.After(duration)` in `WaitTool.Run` with a stoppable `time.Timer`.
- Preserve existing validation, cancellation result, and success messages.
- Preserve the 30 second maximum wait.
- Existing wait tool tests must continue passing.

## Acceptance Criteria

- [ ] `WaitTool.Run` stops/drains the timer when cancellation wins the select.
- [ ] `TestWaitTool_ContextCancel` continues passing.
- [ ] `go test ./internal/tools -run TestWaitTool` passes.
- [ ] `go test ./...` passes.

## Notes

Out of scope:

- Changing wait duration defaults.
- Changing loop detection behavior.
