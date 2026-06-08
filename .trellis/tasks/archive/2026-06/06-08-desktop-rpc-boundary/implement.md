# Implementation Plan

## Scope

Add a local Unix-socket Desktop RPC boundary with codec, broker, listener, system methods, fake desktop tests, and status exposure.

## Steps

1. Run `trellis-before-dev`.
2. Add Desktop RPC package tests:
   - codec round trip
   - codec rejects oversized/invalid frame
   - broker not connected
   - broker request/result correlation
   - broker timeout
   - broker cancel on disconnect
   - listener fake desktop smoke
   - desktop-originated system ping/capabilities
3. Implement `internal/daemon/desktop_rpc/types.go`.
4. Implement `codec.go`.
5. Implement `broker.go`.
6. Implement `listener.go`.
7. Wire broker into `Server`.
8. Expose Desktop RPC connected/listening status through `GET /status`.
9. Update backend quality spec with Desktop RPC boundary contract.
10. Run focused validation:
    - `go test ./internal/daemon/desktop_rpc`
    - `go test ./internal/daemon`
11. Run broader validation:
    - `go test ./...`
    - `git diff --check`
12. Commit, archive child task, and leave Phase6 parent active for the next child.

## Files Likely To Change

- `internal/daemon/desktop_rpc/types.go`
- `internal/daemon/desktop_rpc/codec.go`
- `internal/daemon/desktop_rpc/broker.go`
- `internal/daemon/desktop_rpc/listener.go`
- `internal/daemon/desktop_rpc/*_test.go`
- `internal/daemon/server.go`
- `internal/daemon/server_test.go`
- `.trellis/spec/backend/quality-guidelines.md`
- `.trellis/tasks/06-08-desktop-rpc-boundary/*`

## Review Gates

- Pending RPCs cannot hang after disconnect.
- Socket tests use temp dirs and clean up socket files.
- Status payload exposes connected/listening booleans only, not frame payloads.
- Existing HTTP daemon behavior remains compatible.

## Completion Notes

- Added `internal/daemon/desktop_rpc` with protocol types, frame codec, broker, listener, system method handlers, and fake Desktop tests.
- Wired a Desktop RPC broker into `Server` and exposed `desktop_rpc.listening`, `desktop_rpc.connected`, and `desktop_rpc.pending` through `GET /status`.
- Updated daemon client status types and tests for the new status object.
- Updated backend quality guidelines with the Desktop RPC local boundary contract.

## Validation

- `go test ./internal/daemon/desktop_rpc` — passed.
- `go test ./internal/daemon` — passed.
- `go test ./internal/daemon/desktop_rpc ./internal/daemon` — passed after final cleanup.
- `go test ./...` — passed after final cleanup.
- `git diff --check` — passed.
