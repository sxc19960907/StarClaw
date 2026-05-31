# Implementation Plan

## Checklist

- [x] Start the Trellis task after planning artifacts are complete.
- [x] Add `/skills` route registration.
- [x] Add read-only skills listing handler using `ServerDeps.SkillsDir`.
- [x] Add unit coverage for `/skills` empty and populated behavior.
- [x] Add daemon API smoke test covering health/status, agents, skills, sessions/search, schedule CRUD, and representative errors.
- [x] Run targeted daemon tests.
- [x] Run full `go test ./...` and `go vet ./...`.

## Validation Commands

```bash
go test ./internal/daemon
go test ./...
go vet ./...
```

## Rollback Points

- Revert `internal/daemon/router.go`, `internal/daemon/server.go`, and daemon tests if the route conflicts with a planned broader daemon skills API.
