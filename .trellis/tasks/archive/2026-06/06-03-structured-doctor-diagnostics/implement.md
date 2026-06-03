# Structured Doctor Diagnostics Implementation

## Checklist

- [x] Refactor `cmd/doctor.go` to build a reusable report.
- [x] Add `--json` flag and JSON rendering.
- [x] Update plain-text rendering to use the report.
- [x] Add unit tests for JSON without daemon and with daemon.
- [x] Update CLI, app launch, and release install smoke scripts.
- [x] Update docs for the JSON support/debug mode.
- [x] Run targeted tests, full tests, vet, and smoke scripts.

## Validation

```bash
go test ./cmd ./internal/tui ./internal/daemon
go test ./...
go vet ./...
scripts/smoke_cli.sh
scripts/smoke_app_launch.sh
```

`scripts/smoke_release_install.sh` requires a release archive, so validate it with shell syntax inspection unless an archive is available.
