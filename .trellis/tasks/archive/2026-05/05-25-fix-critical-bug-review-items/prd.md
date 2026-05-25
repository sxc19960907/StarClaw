# Fix critical bug review findings

## Goal

Fix the 5 severe findings from `BUG_REVIEW.md` so the highest-risk correctness, safety, and cost issues are covered by regression tests.

## Requirements

- Fix `internal/client/mock.go` data races by making `MockClient` safe for concurrent tests and callers.
- Fix `internal/agent/loop.go` streaming behavior so a successful stream does not issue a second non-streaming request.
- Fix `internal/context/sanitize.go` so consecutive `user` or `assistant` messages are merged without silently dropping earlier content.
- Fix `internal/context/window.go` so old tool-result compression counts conversation turns correctly and does not over-truncate.
- Fix `internal/daemon/checkpoint.go` ID sanitization so checkpoint paths cannot escape the checkpoint directory.
- Add or update focused regression tests for each fix.
- Keep behavior changes narrowly scoped to the severe findings.

## Acceptance Criteria

- [ ] All 5 severe findings have code fixes.
- [ ] Regression tests fail against the buggy behavior and pass after the fix.
- [ ] Relevant package tests pass.
- [ ] Full `go test ./...` is attempted and the result is reported.
- [ ] No existing unrelated user/Trellis changes are reverted.

## Notes

- Source list: `BUG_REVIEW.md`, severe findings 1-5.
