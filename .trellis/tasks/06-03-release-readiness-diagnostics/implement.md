# Release Readiness Diagnostics Implementation

## Checklist

- [x] Add `cmd/doctor.go` with top-level Cobra command.
- [x] Register `doctorCmd` in `cmd/root.go`.
- [x] Add command unit tests for daemon unavailable/reachable output.
- [x] Extend `scripts/smoke_cli.sh` to validate `starclaw doctor`.
- [x] Update user docs where support/debug commands are listed.
- [x] Run formatting and targeted validation.

## Validation

```bash
go test ./cmd ./internal/tui ./internal/daemon
go test ./...
go vet ./...
scripts/smoke_cli.sh
```

## Risk / Rollback

- Risk is low because this adds a read-only CLI command and reuses existing checks.
- Rollback is removing `cmd/doctor.go`, the root registration, related tests, smoke assertions, and docs.
