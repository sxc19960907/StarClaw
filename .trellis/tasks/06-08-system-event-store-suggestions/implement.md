# Implementation Plan

## Steps

1. Inspect StarClaw's current `internal/agent`, `internal/daemon/server.go`, and route/session validation helpers.
2. Add agent system event and suggestion state types with focused tests.
3. Add daemon `SystemEventStore` with focused tests.
4. Wire stores into `Server`.
5. Add suggestion route validation and handlers.
6. Register suggestion routes.
7. Add handler tests for get/accept/consume behavior.
8. Run:
   - `go test ./internal/agent ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-system-event-store-suggestions`
9. Commit and archive the child task.

## Review Gates

- Do not store system events in the broad observability event store.
- Do not expose all routes/suggestions through diagnostics.
- Do not add real external transport.
- Keep route/session validation explicit and path-safe.

## Completion Notes

TBD.
