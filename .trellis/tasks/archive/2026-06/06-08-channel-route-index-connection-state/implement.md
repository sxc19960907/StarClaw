# Implementation Plan

## Steps

1. Add `reply_route_index.go` and tests.
2. Add `connection_state_cache.go` and tests.
3. Wire stores into `Server`.
4. Record route index on queue create when `external_id` and `route_key` are present.
5. Add channel diagnostic API handlers/routes.
6. Add handler and queue integration tests.
7. Run:
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
8. Commit and archive child task.

## Review Gates

- No external channel transport.
- No message text in route index or connection state diagnostics.
- Bounded index eviction is deterministic.
- Existing queue behavior remains compatible.

## Completion Notes

- Added bounded `ReplyRouteIndex` for `message_id -> route_key`.
- Added `ConnectionStateCache` for binding, transport, and membership state.
- Wired `Server` with route index and connection-state cache.
- Queue creation now records `external_id -> route_key` for successful new messages.
- Added read-only diagnostics endpoints:
  - `GET /channel/routes/{message_id}`
  - `GET /channel/state?platform=&channel_id=`
- Added tests for route index, connection state, API handlers, and queue integration.

## Validation

- `go test ./internal/daemon` — passed.
- `go test ./...` — passed.
- `git diff --check` — passed.
- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-channel-route-index-connection-state` — passed.
