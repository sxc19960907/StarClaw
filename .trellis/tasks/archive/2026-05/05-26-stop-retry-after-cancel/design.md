# Stop Retry After Cancel Design

## Boundary

The runtime change lives in `internal/agent/loop.go`; regression coverage lives in `internal/agent/loop_blackbox_test.go`.

## Behavior

`retryWait(ctx, attempt, cfg)` currently blocks on `time.After(backoff)` and returns no signal. Change it to return `error`:

- `nil` when the backoff timer fires
- `ctx.Err()` when context cancellation wins

`chatWithRetry` should check the returned error at both streaming and non-streaming retry sites and return it wrapped with operation context.

## Timer Cleanup

Use `time.NewTimer(backoff)` and stop/non-blocking-drain it on cancellation, matching the backend timer guidance.

## Compatibility

Successful retries still wait for the computed backoff and retry. The only changed behavior is cancellation during backoff, which now exits immediately.
