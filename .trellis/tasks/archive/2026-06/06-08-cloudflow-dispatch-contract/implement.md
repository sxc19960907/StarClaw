# Implementation Plan

## Steps

1. Add `internal/cloudflow/parse.go`, `display.go`, `dispatch.go`.
2. Add tests for parser, display, and local dispatch.
3. Update `internal/daemon/workflow_command.go`:
   - use `cloudflow.ParseSlash`
   - support `/dag`
   - preserve research strategy
4. Update daemon workflow tests.
5. Run:
   - `go test ./internal/cloudflow`
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
6. Commit and archive child task.

## Review Gates

- No network calls.
- No cloud credentials.
- `/research` and `/swarm` existing behavior remains compatible.
- `/dag` has explicit local workflow steps and metadata.

## Completion Notes

- Added `internal/cloudflow` with slash parser, status display helper, provider interface, and local no-network provider.
- Extended daemon workflow parsing to use `cloudflow.ParseSlash`.
- Added `/dag` auto-orchestration workflow with local steps.
- Preserved research strategy metadata and prompt text.
- Applied workflow route hints to response routing metadata so slash workflows keep their intended route.

## Validation

- `go test ./internal/cloudflow ./internal/daemon` — passed.
- `go test ./...` — passed.
- `git diff --check` — passed.
- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-cloudflow-dispatch-contract` — passed.
