# Current Web UI recovery evidence

## StarClaw evidence

- `internal/daemon/webui/assets/app.js` has `connectEventStream()` with
  `new EventSource("/events")`.
- Current EventSource handlers consume `approval_needed` and
  `approval_resolved`.
- `state.eventStream` tracks `lastEventID`, `status`, `reconnects`, and
  `reconnecting`.
- `source.onerror` marks the daemon pill as reconnecting.
- `source.onopen` marks recovered/running, but does not refresh `/runs`.
- `loadRuns()` already refreshes `state.runs`, Mission Control, home activity,
  quality, comparison, reuse, result library, and related run-dependent panels.
- `renderMissionControl()` already has a `recovered` filter using
  `isRecoveredRun(run)`.

## Backend evidence

- `/events` supports replay through `SubscribeWithReplay`.
- `RunStore` now publishes replayable `run_started`, `run_completed`, and
  `run_error` events.

## Decision

Use EventSource lifecycle events for optimistic run summary updates and a
guarded `/runs` refresh after reconnect for durable convergence. Do not add a
backend route or standalone desktop app in this child.
