# Stop approval timeout timer

## Goal

Make daemon approval waits release their timeout timer promptly when an approval resolves or the caller context is cancelled.

The default approval timeout is five minutes, so cancelled approval waits should not leave timer resources alive until that timeout expires.

## Requirements

- Replace `time.After(b.timeout)` in `ApprovalBroker.WaitForApproval` with a stoppable `time.Timer`.
- Preserve approval resolution behavior.
- Preserve timeout behavior: timeout returns `DecisionDeny, nil`.
- Preserve cancellation behavior: cancellation returns `DecisionDeny, ctx.Err()`.
- Existing approval broker tests must continue passing.

## Acceptance Criteria

- [ ] `WaitForApproval` stops/drains the timeout timer on early return.
- [ ] Approval resolve, deny, timeout, and context cancellation tests pass.
- [ ] `go test ./internal/daemon -run ApprovalBroker` passes.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Notes

Out of scope:

- Changing approval timeout duration.
- Changing approval API payloads.
