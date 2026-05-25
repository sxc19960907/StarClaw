# Fix concurrency and lifecycle races

## Goal

Fix high-priority concurrency and lifecycle races identified by `BUG_REVIEW.md` and race testing, starting with the reproducible `ProcessTool` stdout/stderr buffer race.

## Requirements

- Fix `internal/tools/process.go` so process start/status handling does not read stdout/stderr buffers while exec copy goroutines are writing them.
- Fix `internal/agent/registry.go` concurrent access to registered tools and ordering.
- Fix `internal/agent/readtracker.go` concurrent access to read tracking state.
- Fix heartbeat lifecycle races/deadlocks called out in `BUG_REVIEW.md` when practical within this task.
- Add or update focused tests that cover the corrected behavior.
- Keep unrelated safety/path changes untouched.

## Acceptance Criteria

- [ ] `go test -race ./internal/tools` no longer reports the `ProcessTool` buffer race.
- [ ] Focused race tests for changed packages pass.
- [ ] Full `go test ./...` is attempted and reported.
- [ ] Any remaining race failures outside this task are documented as follow-up.

## Notes

- Sources: `BUG_REVIEW.md` high-priority data-race/lifecycle findings and `go test -race ./internal/tools` failure in `ProcessTool`.
