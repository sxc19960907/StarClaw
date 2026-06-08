# Runtime pause resume support design

## Scope

This child replaces the staged pause/resume `409` behavior for active runs with cooperative runtime control. It does not preempt a tool already running and does not add UI.

## Agent Loop Contract

Add a small pause hook to `internal/agent`:

- `PauseController` interface with `WaitIfPaused(ctx context.Context) error`.
- `AgentLoop.SetPauseController(controller PauseController)`.
- `Run` calls `WaitIfPaused` before each model call and before each tool execution.
- If context is cancelled while paused, `Run` returns the context error.

This keeps agent package independent from daemon run-store types.

## Daemon Control

Add a daemon runtime pause controller:

- Starts unpaused.
- `Pause()` marks paused.
- `Resume()` releases all waiters.
- `Cancel()` releases waiters so cancellation can finish promptly.
- `Paused()` reports state for control responses.

Server keeps controllers in `running` alongside cancel funcs. A small runtime handle struct can hold `cancel` and `pause`.

## API Behavior

`POST /runs/{id}/control`:

- `pause` on active run -> HTTP 200, status `paused`, records control decision and step state.
- `resume` on active paused run -> HTTP 200, status `resumed`, records control decision and step state.
- `resume` on active unpaused run -> HTTP 200, status `running` or `not_paused`; no failure.
- known inactive run -> HTTP 409.
- missing run -> HTTP 404.

`POST /cancel` and `cancel` control must call the handle cancel path, which also unblocks paused waiters.

## State & Observability

- Control decisions use existing `RunControlDecision`.
- Structured events continue through `control_decision`.
- Add/update workflow step `runtime-pause` with `paused`, `resumed`, or `cancelled` safe metadata.
- Run terminal status is still owned by cancellation/completion, not pause/resume.

## Compatibility

- Existing `cancel`, replay, metrics, and run history shapes remain compatible.
- Existing tests that expected pause/resume `409` for known active runs must be updated to the real behavior.

## Rollback

Remove `PauseController` wiring and restore pause/resume conflict branch.
