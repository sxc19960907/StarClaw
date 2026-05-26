# Stop Approval Timeout Timer Design

## Boundary

Runtime change is limited to `internal/daemon/approval.go`. Existing tests in `internal/daemon/approval_test.go` cover approval resolution, timeout, and cancellation behavior.

## Behavior

`WaitForApproval` should create `timer := time.NewTimer(b.timeout)` before its select. The timeout case waits on `timer.C`. A deferred cleanup stops the timer and drains it non-blockingly when needed.

Return contracts stay unchanged:

- resolved approval: selected decision, nil error
- timeout: `DecisionDeny, nil`
- context cancellation: `DecisionDeny, ctx.Err()`
