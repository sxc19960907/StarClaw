# Observability trace export implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Inspect structured event redaction, run-store copying, metrics, and router patterns.
3. Add trace export record type and conversion helpers.
4. Add run-store methods:
   - trace records for all runs
   - trace records for one run
   - atomic JSONL export for all/one run
5. Add daemon routes:
   - `GET /runs/{id}/trace`
   - `GET /traces/export?path=...`
6. Add tests:
   - JSONL all-run export
   - single-run trace response
   - missing run not-found
   - recursive redaction of prompt/tool args/secrets
   - existing metrics behavior remains prompt-free
7. Update backend quality spec with trace export rules.
8. Run validation:
   - `gofmt -w internal/daemon/*.go`
   - `go test ./internal/daemon -run 'Test.*Trace|TestRunStoreStructuredEvents|TestHandleMetrics|TestPersistentRunStore' -count=1`
   - `go test ./...`
   - `git diff --check`
9. Update PRD acceptance criteria, commit, archive child task, and record journal.

## Risk Files

- `internal/daemon/run_store.go`
- `internal/daemon/events.go`
- `internal/daemon/router.go`
- `internal/daemon/server.go`
- `internal/daemon/observability_test.go`
- `.trellis/spec/backend/quality-guidelines.md`

## Non-Goals

- No external OpenTelemetry SDK.
- No collector upload.
- No frontend UI.
- No prompt archive.
