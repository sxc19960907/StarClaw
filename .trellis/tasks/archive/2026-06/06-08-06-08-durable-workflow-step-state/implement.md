# Durable workflow step state implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Inspect run-store persistence, structured event redaction, metrics, and run control tests.
3. Add step status constants and `WorkflowStepState` in `internal/daemon/run_store.go`.
4. Add `Steps []WorkflowStepState` to `RunRecord` and defensive copying in `Get`.
5. Implement `UpsertStep` and `TransitionStep` with timestamp/status behavior and persistence hooks.
6. Emit structured `workflow_step` events using the existing redaction path.
7. Add tests covering:
   - upsert and transition behavior
   - terminal step timestamps without run terminal mutation
   - persistent step recovery
   - metadata redaction in structured events/metrics
   - existing corrupt-file recovery remains safe
8. Update backend quality spec with durable workflow step-state rules.
9. Run validation:
   - `gofmt -w internal/daemon/run_store.go internal/daemon/run_store_persistence_test.go internal/daemon/workflow_step_state_test.go`
   - `go test ./internal/daemon -run 'TestWorkflowStep|TestPersistentRunStore|TestRunStore|TestHandleRunControl|TestHandleMetrics' -count=1`
   - `go test ./...`
   - `git diff --check`
10. Update PRD acceptance criteria, commit, archive child task, and record journal.

## Risk Files

- `internal/daemon/run_store.go`
- `internal/daemon/workflow_step_state_test.go`
- `internal/daemon/run_store_persistence_test.go`
- `.trellis/spec/backend/quality-guidelines.md`

## Non-Goals

- No workflow graph executor.
- No real pause/resume support.
- No replay launch.
- No UI changes.
