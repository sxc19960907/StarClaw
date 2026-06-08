# Run session lifecycle events design

## Boundary

This task connects StarClaw's local run lifecycle source of truth
(`RunStore`) to the local daemon `EventBus`. It does not implement Kocoro's
Cloud IM `MESSAGE_LIFECYCLE` protocol and does not add external transport.

The design is intentionally additive:

- Persisted run records keep their current fields and structured event schema.
- Existing `/events` clients keep receiving current event names.
- New lifecycle events reuse names StarClaw already persists:
  `run_started`, `run_completed`, and `run_error`.

## Architecture

`RunStore` remains responsible for lifecycle transitions. To avoid requiring
every caller of `Start` and `Complete` to remember a separate publish call,
the store gets an optional `EventBus` dependency:

- `NewRunStore(limit)` keeps working with no bus.
- `NewPersistentRunStore(limit, path)` keeps working with no bus.
- `(*RunStore).SetEventBus(bus)` wires a local bus for daemon servers and
  tests.
- `Server` wires its `runStore` to `eventBus` at construction time.

When the store records a lifecycle transition, it publishes a matching
EventBus event after updating the in-memory record and persisted structured
events. Publishing stays best-effort and local; EventBus already drops events
for slow subscribers and stores a bounded in-memory replay history.

## Event Contracts

### `run_started`

Payload:

```json
{
  "schema_version": "2026-06-08",
  "run_id": "req-123",
  "status": "running",
  "agent": "assistant",
  "channel": "http",
  "source": "api",
  "session_id": "sess-optional",
  "started_at": "2026-06-08T..."
}
```

### `run_completed`

Payload:

```json
{
  "schema_version": "2026-06-08",
  "run_id": "req-123",
  "status": "completed",
  "agent": "assistant",
  "channel": "http",
  "source": "api",
  "session_id": "sess-1",
  "started_at": "2026-06-08T...",
  "ended_at": "2026-06-08T...",
  "usage": {"input_tokens": 1, "output_tokens": 2, "total_tokens": 3},
  "budget_status": {"status": "ok", "input_tokens": 1, "output_tokens": 2, "total_tokens": 3},
  "routing": {"complexity": "simple", "route": "direct", "model_tier": "small"},
  "fallback": {"reason": "provider_error", "route": "direct", "model_tier": "small"}
}
```

### `run_error`

Payload is the same terminal summary shape as `run_completed`, with:

```json
{
  "status": "error",
  "error": "short safe summary"
}
```

The payload builder must pass through the existing redaction helpers before
publishing. Prompt text, assistant content, tool args, request bodies, response
bodies, and sensitive tokens are never included in clear text.

## Data Flow

1. Caller invokes `s.runStore.Start(req)`.
2. Store creates/updates the run record and appends `run_started` structured
   event.
3. Store publishes EventBus `run_started` with redacted safe metadata.
4. `/events` live clients receive the event; reconnecting clients can replay it
   from the EventBus ring by `last_event_id`.
5. Caller invokes `s.runStore.Complete(id, response, err)`.
6. Store records budget/routing/fallback structured events, sets terminal
   status, appends `run_completed` or `run_error`, and publishes the matching
   EventBus lifecycle event.

## Compatibility

The EventBus dependency is optional, so unit tests and persistence-only uses
remain valid. Existing run record JSON is unchanged. Event IDs are assigned by
the EventBus as before.

## Rollback

Rollback is straightforward: remove the optional bus wiring and lifecycle
publish calls. Persisted run records remain compatible because no storage
schema changes are required.
