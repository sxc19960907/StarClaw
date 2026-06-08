# Design

## Boundaries

This task introduces Kocoro-style runtime signals without mixing them into StarClaw's observability event stream.

- Observability events remain under daemon run traces, metrics, and exports.
- System events are daemon-authored, route-scoped messages that can be drained for a specific run/route.
- Suggestions are session-scoped UI state and are consumed independently of normal message posting.

## System Event Store

Add `internal/daemon/system_event_store.go`:

- `SystemEventStore` holds `map[string][]agent.SystemEvent`.
- Constructor accepts cap per route and falls back to 20.
- `Enqueue(routeKey string, ev agent.SystemEvent)`:
  - no-op for nil store or empty route key.
  - if `ev.ContextKey` matches the immediately previous queued event, replace the previous event.
  - if over cap, retain the newest cap entries.
- `Drain(routeKey string)`:
  - returns FIFO events and deletes the route queue.
- `Forget(routeKey string)`:
  - deletes without returning events.

The store is process-local for this slice. Durability across daemon restart is deferred because Kocoro's implementation is also daemon-memory scoped.

## Agent System Event Type

Add `internal/agent/system_event.go`:

- `SystemEvent` with text, trust flag, context key, and timestamp.
- Minimal helper to render events for prompt/run injection.

No broad prompt builder integration is required unless a narrow existing run path already has a route-key preflight point. If integration is not obvious, this task should ship the type/store/tests and leave lifecycle injection to `delivery-inject-lifecycle-depth`.

## Suggestions

Add `internal/agent/suggestion.go` and `suggestion_state.go` or local daemon equivalents if StarClaw's existing package layout fits better:

- Suggestion record: text, timestamp, accepted flag/time.
- Store keyed by session id.
- Methods: `Set`, `Get`, `MarkAccepted`, `Clear`.

Add daemon handlers:

- `GET /sessions/{id}/suggestion`
- `POST /sessions/{id}/suggestion/accept`
- If named-agent routes are already supported in StarClaw session routing, also add:
  - `GET /agents/{name}/sessions/{id}/suggestion`
  - `POST /agents/{name}/sessions/{id}/suggestion/accept`

Validation should reject missing or path-traversal session ids.

## Server Wiring

Add fields to `Server`:

- `systemEvents *SystemEventStore`
- `suggestions *agent.SuggestionState` or equivalent.

Initialize in `NewServer`.

Register routes in `router.go` near session/agent routes.

## Privacy and Safety

- Events are scoped by route key, not global session.
- Suggestion retrieval is scoped by session id.
- No external channel credentials or cloud calls are introduced.
- Diagnostic endpoints should not expose unrelated route queues or all suggestions.

## Trade-offs

- Process-local stores are enough for this slice and match the immediate Kocoro evidence, but daemon restart will lose queued events/suggestions.
- Deferring injection into the model loop keeps this task small; delivery injection can become the first real producer/consumer once the store is in place.
