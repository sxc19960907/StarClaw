# Design

## Architecture

This task tightens existing launch/status surfaces:

- CLI readiness: `cmd/daemon.go`
- Doctor/runtime report: `cmd/doctor.go`
- Daemon runtime JSON: `internal/daemon/server.go`
- GUI Version page: `internal/daemon/webui/assets/app.js`
- Smoke/docs: `scripts/`, `README.md`

No new daemon process model is required.

## Data Flow

Launch/runtime context should be consistent across surfaces:

`daemon constants/config paths` -> CLI `app --check` output
`daemon server deps/version` -> `/version` and `/diagnostics`
`/version` and `/diagnostics` -> GUI Version/Diagnostics panels
`scripts/*smoke*` -> regression coverage

## Contracts

### CLI `app --check`

Text output should include:

- `Launch:        starclaw app`
- `Daemon:        running|not running`
- `Web UI:        http://127.0.0.1:7533/app/`
- `Health:        http://127.0.0.1:7533/health`
- `Status API:    http://127.0.0.1:7533/status`
- `Diagnostics:   http://127.0.0.1:7533/diagnostics`
- `Data:          <starclaw dir>`
- `Config:        <starclaw dir>/config.yaml`

### Runtime JSON

`/version` and `/diagnostics` should expose matching launch/runtime fields. Existing consumers must remain compatible because fields are additive.

## Compatibility

- Existing smoke selectors and JSON keys stay valid.
- Existing command behavior remains unchanged except additional readiness lines.
- No secrets are exposed.

## Rollback

Revert changed CLI output lines, JSON fields, GUI rendering rows, docs, and smoke assertions. No data migration is involved.
