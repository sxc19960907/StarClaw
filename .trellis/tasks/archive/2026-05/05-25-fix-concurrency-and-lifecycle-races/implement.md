# Implementation Plan

## Scope

Fix concurrency/lifecycle issues in:

- `internal/tools/process.go`
- `internal/agent/registry.go`
- `internal/agent/readtracker.go`
- `internal/heartbeat/heartbeat.go`

## Checklist

1. Inspect implementations and existing tests.
2. Fix `ProcessTool` output capture race without changing user-facing process semantics.
3. Add synchronization to registry/readtracker while preserving ordering behavior.
4. Fix heartbeat nil-dependency, Close-before-Start, and Start/Close races.
5. Run focused tests:
   - `go test ./internal/tools ./internal/agent ./internal/heartbeat`
   - `go test -race ./internal/tools ./internal/agent ./internal/heartbeat`
6. Run full suite:
   - `go test ./...`

## Risk Notes

- Avoid broad refactors; prefer narrow locking and lifecycle guards.
- Be careful with global process-manager state in tests.
- If race testing exposes unrelated issues in other packages, document them rather than expanding scope without need.
