# Implementation Plan

## Checklist

- [x] Start the Trellis task after planning artifacts are complete.
- [x] Add embedded web UI handlers and route registration.
- [x] Build static HTML shell.
- [x] Build CSS for Codex-inspired workbench layout and responsive behavior.
- [x] Build browser JS client for status, chat, agents, skills, sessions, schedules.
- [x] Add daemon route tests.
- [x] Run targeted tests.
- [x] Run full `go test ./...` and `go vet ./...`.
- [x] Run local browser screenshot verification.

## Validation Commands

```bash
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
```

## Rollback Points

- Revert `internal/daemon/webui.go`, `internal/daemon/webui/`, router changes, and route tests if embedded GUI serving conflicts with a future standalone app architecture.
