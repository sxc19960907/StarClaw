# WS controller cloud lifecycle

## Goal

Add a local cloud WebSocket lifecycle controller boundary so StarClaw can start, stop, restart, and report future cloud runtime loops in a Kocoro-compatible shape without enabling real cloud credentials or external WebSocket transport.

## Confirmed Facts

- Kocoro has `internal/daemon/ws_controller.go` for `Start`, `Stop`, `Restart`, and `IsRunning` around a reconnecting cloud WebSocket client.
- StarClaw currently has no equivalent runtime lifecycle controller.
- StarClaw does have local cloud delegation HTTP client code, but not a long-running WS client.
- Phase8 must remain local-first and must not enable real cloud transport without explicit approval.

## Requirements

- Add a daemon-local lifecycle controller with start/stop/restart/status semantics.
- Accept an injectable runner function so tests can prove goroutine lifecycle without real networking.
- Start must be idempotent while running.
- Stop must be idempotent and cancel the active run.
- Restart must cancel the active run, wait for it to drain, then start a fresh run.
- Status must expose running state, start time, stop time, restart count, and last error in content-free form.
- Wire the controller into `Server`.
- Add read/control daemon APIs for local lifecycle status:
  - `GET /cloud/lifecycle`
  - `POST /cloud/lifecycle` with action `start`, `stop`, or `restart`
- Default controller must be disabled/no-op: no real cloud connection or credential use.

## Acceptance Criteria

- [ ] Unit tests cover start, idempotent start, stop, restart, status, and runner error capture.
- [ ] API tests cover GET status and POST start/stop/restart.
- [ ] API response makes clear real cloud transport is not configured.
- [ ] No real network call is made by default.
- [ ] `go test ./internal/daemon` passes.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Out of Scope

- Real Shannon Cloud WebSocket client.
- Credential/keychain integration.
- Auth sign-in/out state machine.
- External channel delivery transport.

## Evidence

- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/ws_controller.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/ws_controller_test.go`
