# Durable workflow step state design

## Scope

This child introduces a durable step-state contract inside daemon run records. It does not execute workflow graphs, resume tools, or launch replay. The goal is to make future runtime controls read/write a stable, persisted step timeline.

## Data Model

- Add `WorkflowStepState` to `internal/daemon/run_store.go`.
- Add `Steps []WorkflowStepState` to `RunRecord` with JSON field `steps`.
- Step fields:
  - `id`: stable step id unique within a run.
  - `title`: user-facing label safe for run detail.
  - `status`: one of `planned`, `running`, `blocked`, `waiting_approval`, `completed`, `failed`, `cancelled`, `skipped`.
  - `sequence`: optional ordering value.
  - `parent_id`: optional hierarchy link.
  - `attempt`: attempt count, defaulting to 1 on first write.
  - `started_at`, `updated_at`, `ended_at`.
  - `metadata`: redacted map for non-secret state.

## Store API

- `UpsertStep(runID string, step WorkflowStepState) bool`
  - Creates or replaces a step by id.
  - Defaults blank status to `planned`.
  - Defaults zero attempt to `1`.
  - Maintains stable ordering by existing position, then `sequence`, then append.
  - Persists after mutation.
- `TransitionStep(runID, stepID, status string, metadata map[string]any) bool`
  - Updates status and timestamps.
  - Sets `started_at` when entering `running`.
  - Sets `ended_at` for terminal statuses.
  - Merges redacted metadata.
  - Emits a structured `workflow_step` event.

## Redaction

Use the same structured event redaction path already used by `RunStore.AddEvent`. Step metadata must not include prompt text, tool args, raw provider response bodies, API keys, tokens, or secret values in structured events or metrics.

## Persistence

The existing `RunRecord` JSON envelope will naturally include `steps`. Recovery must defensively copy steps in `Get`, preserve them on `List`/`Get`, and keep corrupt JSON behavior unchanged.

## Compatibility

- No route shape changes are required in this slice.
- Existing run summaries remain unchanged.
- Existing metrics remain aggregate-only; step data appears only as event counts, not raw metadata.
- Existing `pause`/`resume` staged `409` and replay approval-required behavior remain unchanged.

## Rollback

Remove `WorkflowStepState`, `RunRecord.Steps`, store APIs, and tests. Persistent run records without `steps` remain compatible.
