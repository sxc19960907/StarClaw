# Safe replay execution boundary implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Inspect run control, approval requester, run-store persistence, workflow step state, and daemon tests.
3. Add replay helper functions in daemon server code:
   - build redacted replay plan
   - clone source request for replay
   - generate replay request id
   - execute approved replay
4. Update `handleRunControl` replay branch:
   - approved false -> existing approval-required response
   - approved true -> approved control decision + replay launch
5. Add workflow step state for replay approval/launch boundaries.
6. Ensure replay runs through `s.runAgent` and recording handler so tool approval and run events remain normal.
7. Add tests for:
   - unapproved replay still does not launch a run
   - approved replay launches a new run and links source/replay ids
   - replay response redacts source prompt
   - metrics do not leak source prompt
   - control/step metadata records approval boundary
8. Update backend quality spec with approved replay boundary rules.
9. Run validation:
   - `gofmt -w internal/daemon/server.go internal/daemon/server_test.go`
   - `go test ./internal/daemon -run 'TestHandleRunControlReplay|TestHandleRunControlValidation|TestHandleMetrics|TestWorkflowStep|TestPersistentRunStore' -count=1`
   - `go test ./...`
   - `git diff --check`
10. Update PRD acceptance criteria, commit, archive child task, and record journal.

## Risk Files

- `internal/daemon/server.go`
- `internal/daemon/server_test.go`
- `.trellis/spec/backend/quality-guidelines.md`

## Non-Goals

- No deterministic replay engine.
- No execution of recorded tool outputs.
- No frontend replay UI.
- No real pause/resume.
