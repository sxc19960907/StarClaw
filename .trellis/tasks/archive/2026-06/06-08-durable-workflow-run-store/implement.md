# Durable workflow run store implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Inspect `internal/daemon/run_store.go`, server construction, and existing run/control/metrics tests.
3. Add persistence envelope/types and persistent constructor.
4. Implement load behavior:
   - parse JSON envelope
   - ignore nil/malformed records safely
   - clamp to limit
   - rebuild order and event sequence counters
5. Implement save behavior:
   - temp file in same directory
   - `json.Encoder` with stable output
   - rename into place
6. Hook save after store mutations without changing existing public mutation signatures.
7. Add `run_store_persistence_test.go` for:
   - restart recovery
   - control/structured events recovery
   - corrupt file tolerance
   - limit enforcement
8. Update backend quality spec with durable run-store rules.
9. Run validation:
   - `gofmt -w internal/daemon/run_store.go internal/daemon/run_store_persistence_test.go`
   - `go test ./internal/daemon -run 'TestRunStore|TestHandleRunControl|TestHandleMetrics' -count=1`
   - `go test ./...`
   - `git diff --check`
10. Commit, archive child task, record journal.

## Risk Files

- `internal/daemon/run_store.go`
- `internal/daemon/run_store_persistence_test.go`
- `.trellis/spec/backend/quality-guidelines.md`

## Rollback Point

If persistence hooks become invasive, keep this slice to constructor/load/save helpers and defer daemon wiring.
