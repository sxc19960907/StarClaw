# Kocoro lifecycle gap notes

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Kocoro Evidence

- `internal/daemon/events.go` defines `EventRunStatus = "run_status"` and an
  EventBus with `SubscribeWithReplay`.
- `internal/daemon/lifecycle.go` defines Cloud IM lifecycle support:
  `CapIMMessageLifecycleV1`, `EventTypeMessageLifecycle =
  "MESSAGE_LIFECYCLE"`, and states `received`, `processing`, `done`,
  `cleared`.
- `internal/daemon/lifecycle_test.go` and
  `internal/daemon/lifecycle_completion_test.go` verify IM lifecycle
  processing/done/cleared behavior.
- `internal/daemon/bus_handler.go` publishes `run_status` from agent
  `OnRunStatus`.

## StarClaw Evidence

- `internal/daemon/run_store.go` persists structured lifecycle events:
  `run_started`, `run_completed`, and `run_error`.
- `internal/daemon/events.go` already classifies these event types and has
  replayable EventBus history.
- `internal/daemon/server.go` uses `SubscribeWithReplay` for `/events`.
- `internal/daemon/run_store.go` does not currently publish run lifecycle
  transitions onto the EventBus.

## Decision

For this child, StarClaw should close the local recovery gap by publishing
safe run lifecycle events on the local EventBus. Kocoro's Cloud IM
`MESSAGE_LIFECYCLE` protocol remains intentionally out of scope because
StarClaw is preserving local-first behavior and no real cloud transport.
