# Stop SSE Reconnect Timer Design

## Boundary

The code change is limited to `internal/client/sse.go`; tests live in `internal/client/sse_test.go`.

## Behavior

When `connectOnce` fails, `run` waits for `delay` before reconnecting. Today that wait uses `time.After(delay)`. The replacement should create `timer := time.NewTimer(delay)`, select between `ctx.Done()` and `timer.C`, then stop and drain the timer when cancellation wins.

## Compatibility

No public API changes. The reconnect delay sequence stays the same.
