# Runtime pause resume support

## Goal

Replace staged pause/resume 409 behavior with cooperative runtime pause/resume support where the daemon can safely honor it.

## Requirements

- Replace staged `pause`/`resume` `409` responses for active runs with cooperative runtime control.
- Pause must only affect active runs; inactive historical runs should remain conflict responses.
- Runtime pause must wait at safe boundaries before model calls and before tool calls. It must not interrupt a tool call already executing.
- Resume must release a paused active run and allow execution to continue.
- Pause/resume decisions must be recorded in run control metadata, durable step state, and structured events.
- Pause/resume must preserve cancel behavior: cancelling a paused run must unblock it and terminate via the existing cancellation path.
- The implementation must preserve existing approval, replay, metrics, and run-store behavior.
- No frontend changes, cloud sync, database state, or deterministic replay changes in this slice.

## Acceptance Criteria

- [x] Active run pause returns HTTP 200 and records `paused` control metadata.
- [x] Active paused run resume returns HTTP 200 and records `resumed` control metadata.
- [x] Unknown run returns HTTP 404 for pause/resume.
- [x] Known inactive run returns HTTP 409 for pause/resume.
- [x] Agent loop waits at cooperative pause points before model/tool calls and continues after resume.
- [x] Cancel unblocks a paused run and preserves cancelled terminal status.
- [x] Pause/resume structured events and workflow step state are visible on run detail.
- [x] Existing replay approval boundary, run/control/metrics, and full tests continue to pass.

## Notes

- Parent task: `06-08-astria-phase-4-runtime-durability-replay`.
- This child implements cooperative pause/resume only at safe runtime boundaries.
