# EventBus replay SSE resilience design

## Current State

`internal/daemon/events.go` stores a bounded in-memory history and exposes `EventsSince(lastID string)`. `internal/daemon/server.go` opens `/events` SSE by subscribing to the bus, then replaying `EventsSince(cursor)`, then reading live events.

This mostly works, but the subscribe/replay boundary is split across calls. A publish between subscribe and replay can be present both in the subscriber channel and in `EventsSince`, creating duplicate risk. Kocoro avoids this with a locked subscribe-with-replay operation.

## Target Contract

- `SubscribeWithReplay(id, lastID)` returns both missed events and the live channel while holding the EventBus lock.
- Replay events are every buffered event with numeric ID greater than `lastID`.
- Invalid or empty `lastID` behaves like cursor `0`.
- The returned live channel receives only events published after the subscription is installed.
- `/events` writes replay events first, flushes, then streams live events and keepalive comments.

## Web UI Contract

Astria's `EventSource` client should keep small local event-stream state:

- `lastEventID`: most recent SSE event `lastEventId`.
- `status`: connecting, running, reconnecting, recovered.
- `reconnects`: count of error-to-open recovery cycles.

The browser's native `EventSource` automatically sends `Last-Event-ID` on reconnect. The UI state is for observability and recovery cues, not for manually constructing the URL.

## Compatibility

- Existing `Subscribe(id)` remains for call sites that do not need replay.
- Existing `/events` route and payloads remain unchanged.
- Legacy Web UI approval event handling remains unchanged.
- The bounded in-memory ring remains local-only and non-durable.

## Rollback

Revert only:

- new EventBus subscribe-with-replay helper,
- `/events` handler call-site change,
- Web UI event-stream state additions,
- associated tests.

Existing `EventsSince` and current `/events` behavior can remain as fallback.
