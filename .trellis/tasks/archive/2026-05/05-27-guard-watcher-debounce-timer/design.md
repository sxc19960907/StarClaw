# Design

## Current Behavior

`Watcher.appendEvent` stops any existing per-agent debounce timer and creates a new `time.AfterFunc`. If an older timer callback has already been queued when `Stop` is called, that callback can still call `flush(agent)`.

Because `flush` deletes the current batch for the agent without knowing which timer generation fired, a stale callback can prematurely flush a newer batch.

`Close` stops timers, but a callback already queued before `Stop` can still run. `flush` checks `w.ctx.Err()` when the watcher was started, but manually used or not-yet-started watchers do not have a cancellation token to invalidate stale callbacks.

## Proposed Change

Track a per-agent debounce generation under `w.mu`. Each append increments the agent generation and captures it in the timer callback. The callback calls a generation-aware flush that returns without side effects unless its generation still matches the current agent generation.

On `Close`, increment all agent generations while stopping timers so queued callbacks are invalidated.

## Scope

- `internal/watcher/watcher.go`
- `internal/watcher/watcher_test.go`
