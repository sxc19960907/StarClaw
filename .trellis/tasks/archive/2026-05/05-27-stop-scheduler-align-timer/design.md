# Design

## Current Behavior

`Scheduler.Start` evaluates once on startup, then waits until the next wall-clock minute boundary with `time.After(next.Sub(now))`. The wait is cancellable because the select also listens to `ctx.Done()`, but the underlying timer cannot be stopped when cancellation wins.

## Proposed Change

Use `time.NewTimer(delay)` for the initial alignment wait. If `ctx.Done()` fires before the timer channel, call `Stop`; when `Stop` reports that the timer already fired, perform a non-blocking drain from the timer channel.

After the timer fires normally, continue with the existing minute ticker behavior unchanged.

## Scope

- `internal/daemon/scheduler.go`
- `internal/daemon/scheduler_test.go` if current cancellation coverage needs tightening
