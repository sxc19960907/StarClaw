# Runtime pause resume support implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Inspect agent loop model/tool call boundaries, daemon active run tracking, cancel, replay, and run control tests.
3. Add `PauseController` interface and setter in `internal/agent/loop.go`.
4. Call `WaitIfPaused` before `chatWithRetry` and before `executeTool`.
5. Add agent unit/blackbox tests proving pause blocks before a model call and cancellation exits while paused.
6. Add daemon runtime handle/controller in `internal/daemon/server.go` or a small new daemon file.
7. Replace `running` map values from raw cancel func to runtime handle.
8. Wire pause controller into `RunAgentRequest`/`RunAgentWithApproval` or server execution path without changing public API JSON.
9. Update pause/resume/cancel handlers and tests.
10. Add backend quality spec for cooperative pause/resume.
11. Run validation:
    - `gofmt -w internal/agent/loop.go internal/agent/*pause*_test.go internal/daemon/server.go internal/daemon/server_test.go`
    - `go test ./internal/agent -run 'TestAgentLoop_.*Pause|TestAgentLoop_.*Cancel' -count=1`
    - `go test ./internal/daemon -run 'TestHandleRunControlPauseResume|TestHandleCancel|TestHandleRunControlCancel|TestHandleRunControlReplay|TestHandleMetrics' -count=1`
    - `go test ./...`
    - `git diff --check`
12. Update PRD acceptance criteria, commit, archive child task, and record journal.

## Risk Files

- `internal/agent/loop.go`
- `internal/agent/loop_blackbox_test.go`
- `internal/daemon/server.go`
- `internal/daemon/server_test.go`
- `.trellis/spec/backend/quality-guidelines.md`

## Non-Goals

- No preemption of currently executing tools.
- No UI changes.
- No deterministic replay.
- No persisted process resurrection after OS kill.
