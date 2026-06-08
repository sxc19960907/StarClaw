# Workflow control API implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Add run control metadata to `RunRecord` and copy it defensively from `RunStore.Get`.
3. Add `RunStore.AddControlDecision` and structured `control_decision` event emission.
4. Update `POST /cancel` to record successful cancel decisions while preserving existing response compatibility.
5. Add `POST /runs/{id}/control`:
   - `cancel`: cancel active run and record metadata.
   - `pause` / `resume`: return `409` staged unsupported response and record metadata.
   - `replay`: return approval-required replay plan and record metadata without launching a run.
6. Register the new route and update route tests.
7. Add daemon tests for cancel metadata, route control cancel, staged pause/resume, replay approval boundary, and missing run/action validation.
8. Update backend quality spec with the workflow-control API contract.
9. Mark PRD acceptance criteria after tests pass.
10. Run validation:
    - `gofmt -w internal/daemon/*.go`
    - `go test ./internal/daemon -run 'TestHandleCancel|TestRunControl|TestRouterRegistersRoutes' -count=1`
    - `go test ./...`
    - `git diff --check`
11. Commit, archive task, and record journal.

## Risk files

- `internal/daemon/server.go`
- `internal/daemon/router.go`
- `internal/daemon/run_store.go`
- `internal/daemon/server_test.go`
- `internal/daemon/router_test.go`

## Rollback points

- If cancel behavior regresses, keep `/cancel` as-is and defer route control.
- If replay request shape is unclear, return only `approval_required` with `source_run_id` and omit request details.
- If structured events create compatibility risk, keep control metadata on run detail and defer structured control events.
