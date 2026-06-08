# Phase 5 secret leakage regression implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Inspect existing redaction helpers, observability tests, support/diagnostic tests, run summary/recovery tests, and Web UI trace/replay rendering.
3. Add a reusable forbidden-value assertion helper/fixture in daemon tests.
4. Add or extend backend tests for metrics, trace read/export, run summaries, replay-control, diagnostics/support output, and recovery metadata.
5. Add Web UI static/rendering regression coverage where the existing test harness supports it.
6. Run validation:
   - Targeted daemon tests for the new regression names.
   - `go test ./internal/daemon ./cmd`
   - `go test ./...`
   - `git diff --check`
7. Update PRD acceptance criteria, commit, archive child, and record journal.

## Risk Files

- `internal/daemon/observability_test.go`
- `internal/daemon/server_test.go`
- `internal/daemon/run_store_persistence_test.go`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui_test.go`
- `cmd/doctor_test.go`

## Non-Goals

- No prompt archive.
- No external secret scanning service.
- No broad Web UI redesign.
