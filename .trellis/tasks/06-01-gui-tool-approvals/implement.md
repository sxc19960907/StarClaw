# Implementation Plan

## Checklist

- [x] Add an agent-loop approval requester interface and setter.
- [x] Wire permission `ask` and `RequiresApproval()` into the requester.
- [x] Add daemon approval requester that publishes events and waits on `ApprovalBroker`.
- [x] Inject daemon approval requester in HTTP/SSE runs.
- [x] Add backend tests for allow, deny, timeout/cancel-safe behavior where practical, and event payloads.
- [x] Add Web UI event subscription and approval card rendering.
- [x] Add Web UI decision calls to `POST /approval`.
- [x] Update backend spec if the approval contract changed.
- [x] Run validation commands.

## Validation Commands

```bash
node --check internal/daemon/webui/assets/app.js
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
```

## Risky Files

- `internal/agent/loop.go`
- `internal/daemon/runner.go`
- `internal/daemon/server.go`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`

## Rollback Points

- Revert approval requester interface and daemon adapter if it affects non-daemon agent execution.
- Revert Web UI `/events` subscription independently if backend approval behavior works but UI handling regresses.
