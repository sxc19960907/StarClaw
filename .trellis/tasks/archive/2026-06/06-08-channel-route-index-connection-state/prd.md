# Channel route index connection state

## Goal

Add the next Phase7 Kocoro parity slice: local route indexing and connection-state cache foundations for future channel delivery. This gives StarClaw durable-enough in-memory routing awareness without adding external IM transports.

## Requirements

- Use Kocoro evidence:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/reply_route_index.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/connection_state_cache.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/message_origin.go`
- Add a bounded reply route index:
  - map inbound/external message id to route key
  - update existing entries and evict oldest when over cap
- Add a connection state cache:
  - platform binding/transport state
  - platform+channel membership state
  - deterministic preamble/status rendering
- Wire queue creation to record `external_id -> route_key` when available.
- Expose local diagnostic API for route lookup and connection state, without external transport.

## Acceptance Criteria

- [ ] Route index supports put/get/update/eviction.
- [ ] Queue create records route index for messages with `external_id`.
- [ ] Connection state cache records membership, binding, and transport changes.
- [ ] Connection state rendering is deterministic and content-free.
- [ ] Daemon exposes a read-only diagnostics endpoint for channel state and route lookup.
- [ ] Tests cover route index, connection state cache, API handlers, and queue integration.
- [ ] Full project tests pass.

