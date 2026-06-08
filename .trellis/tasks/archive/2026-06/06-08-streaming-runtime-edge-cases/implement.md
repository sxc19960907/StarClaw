# Streaming runtime edge cases implementation plan

## Steps

1. Read applicable backend specs and Phase 8 research.
2. Add focused tests in `internal/daemon/openai_api_test.go` for:
   - fallback `OnText` with no stream deltas,
   - duplicate suppression after stream deltas,
   - stream run failure after headers are written,
   - stream result error after headers are written.
3. Adjust `internal/daemon/openai_api.go` only if tests expose ambiguous or missing behavior.
4. Keep successful stream behavior unchanged.
5. Validate with:
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-streaming-runtime-edge-cases`

## Rollback

Revert only this task's edits to `internal/daemon/openai_api.go`, `internal/daemon/openai_api_test.go`, and this task directory if a compatibility issue appears.
