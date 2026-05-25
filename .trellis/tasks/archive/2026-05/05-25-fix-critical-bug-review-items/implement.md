# Implementation Plan

## Scope

Fix only the 5 severe findings from `BUG_REVIEW.md` and add focused regression coverage.

## Checklist

1. Inspect current implementations and tests for:
   - `internal/client/mock.go`
   - `internal/agent/loop.go`
   - `internal/context/sanitize.go`
   - `internal/context/window.go`
   - `internal/daemon/checkpoint.go`
2. Implement minimal fixes:
   - Add synchronization and defensive copies to `MockClient`.
   - Return the streamed response path cleanly after successful `StreamChat`.
   - Merge consecutive same-role content instead of replacing it.
   - Count completed assistant/user pairs correctly in tool-result compression.
   - Sanitize checkpoint IDs by rejecting/neutralizing traversal and path separators.
3. Add regression tests near each source file.
4. Run focused package tests:
   - `go test ./internal/client ./internal/agent ./internal/context ./internal/daemon`
5. Run full test suite:
   - `go test ./...`
6. Run race test where practical:
   - `go test -race ./internal/client ./internal/agent ./internal/context ./internal/daemon`

## Risk Notes

- `agent/loop.go` may have streaming response contract assumptions; inspect existing streaming tests before changing control flow.
- Context message merging may affect prompt shape; keep separators simple and deterministic.
- Checkpoint ID sanitization must preserve existing valid IDs while making traversal impossible.
