# Implementation Plan

## Steps

1. Add `cloud_lifecycle.go` with controller, status, default runner, and tests.
2. Add `cloud_lifecycle_api.go` with GET/POST handlers.
3. Wire controller into `Server` and `router.go`.
4. Add API tests.
5. Run:
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-ws-controller-cloud-lifecycle`
6. Commit and archive the child task.

## Review Gates

- No real network calls.
- No cloud credentials or tokens.
- Restart waits for the old run to drain before starting again.
- Status remains content-free.

## Completion Notes

TBD.
