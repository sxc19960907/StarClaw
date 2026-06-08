# EventBus replay SSE resilience

## Goal

Align StarClaw's local `/events` SSE replay contract with the Kocoro baseline enough that Astria clients can reconnect without missing or misordering critical daemon events.

## Confirmed Facts

- Kocoro's `/events` handler accepts `last_event_id` query and `Last-Event-ID` header, then uses `SubscribeWithReplay(lastID)` so replay and live subscription share one locked EventBus operation.
- StarClaw's `/events` handler already accepts both cursor sources, emits `id:` frames, sends 30-second keepalives, and has tests for query/header replay.
- StarClaw currently subscribes first and then calls `EventsSince(lastEventID)`, which can blur the boundary between replayed and live events under concurrent publish timing.
- Astria Web UI currently opens `new EventSource("/events")`, handles approval events, and sets a reconnecting pill on `onerror`, but it does not track last seen event ID or expose a recovered state.

## Requirements

- Add an atomic replay+subscribe EventBus API for `/events`.
- Preserve existing `Subscribe(id)`, `Unsubscribe(id)`, `Publish`, and `EventsSince` behavior for existing tests and callers.
- Make `/events` use the atomic replay+subscribe path when a valid cursor is present.
- Keep invalid cursor behavior compatible: replay all buffered history rather than failing the stream.
- Keep keepalive behavior and existing SSE framing.
- Add focused backend tests for atomic replay order and `/events` cursor behavior.
- Add Web UI contract coverage that Astria tracks event stream IDs/reconnect state without requiring a browser runtime change beyond existing app JS.

## Acceptance Criteria

- [ ] `EventBus` exposes an atomic subscribe-with-replay method.
- [ ] `/events?last_event_id=N` and `Last-Event-ID: N` replay missed events before live events without duplicate or out-of-order delivery in tested scenarios.
- [ ] Invalid cursors continue to replay buffered history.
- [ ] Keepalive comments remain supported.
- [ ] Astria Web UI records last seen event IDs and can mark the event stream as recovered after reconnect.
- [ ] `go test ./internal/daemon -run 'TestEventBus|TestHandleEvents|TestWebUI' -count=1 -timeout=90s` passes.
- [ ] `go test ./internal/daemon -count=1 -timeout=90s` and `go test ./...` pass before commit.

## Out of Scope

- Persisting the full event replay ring across daemon restarts.
- Adding cloud event transport, Shannon Cloud auth, or Desktop RPC behavior.
- Reworking run-store structured event persistence.
