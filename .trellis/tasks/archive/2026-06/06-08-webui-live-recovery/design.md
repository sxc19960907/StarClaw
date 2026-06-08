# WebUI live recovery design

## Boundary

This task updates only the embedded daemon Web UI client. The backend event
surface already exists:

- `/events` supports replay and `Last-Event-ID` continuity.
- `RunStore` publishes `run_started`, `run_completed`, and `run_error`.
- `/runs` and `/runs/{id}` expose durable run summaries/details.

The Web UI should use those existing surfaces rather than adding a new route.

## Client State

Add a lightweight recovery marker to `state.eventStream`:

- `lastRecoveredAt`: timestamp of the most recent successful reconnect.
- `refreshingRuns`: guard to avoid overlapping `/runs` refreshes.

Run summaries updated from EventBus lifecycle payloads should include:

- `id`
- `status`
- `agent`
- `channel`
- `source`
- `session_id`
- `started_at`
- `ended_at`
- `usage`
- `recovered: true`
- `recovery_source: "event_stream"`

The UI must not insert prompt text, raw response, request, content, tool args,
or deltas from lifecycle payloads. If future payloads accidentally include
those fields, the front-end mapper omits them.

## Event Flow

1. `connectEventStream()` opens `/events`.
2. Existing approval listeners continue unchanged.
3. New listeners consume:
   - `run_started`
   - `run_completed`
   - `run_error`
4. Each listener:
   - tracks `lastEventId`
   - parses JSON safely
   - maps lifecycle payload to a safe run summary
   - upserts the summary into `state.runs`
   - rerenders run list, Mission Control, home activity, and dependent panels
   - if the active run is affected, optionally reloads run detail/trace.
5. `source.onopen` detects recovered reconnects and schedules a guarded
   `loadRuns()` so durable backend state wins after replay.

## Compatibility

Manual `refreshAll()` and `loadRuns()` remain the authoritative durable
refresh path. EventStream lifecycle updates are optimistic/near-real-time and
additive.

If a lifecycle payload is malformed or missing a run id, it is ignored. This
matches current approval behavior where invalid JSON produces an empty object
and no UI crash.

## Rollback

Rollback is a static asset-only revert: remove lifecycle listeners and helper
functions. The backend EventBus lifecycle events remain harmless for clients
that ignore them.
