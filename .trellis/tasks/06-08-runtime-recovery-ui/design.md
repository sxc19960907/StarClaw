# Runtime recovery UI design

## Scope

This task updates the daemon-hosted static Astria Web UI to make Phase 4 runtime state visible in Mission Control. It consumes existing run records plus `GET /runs/{id}/trace`; it does not add new backend runtime behavior.

## UI Data Model

The UI derives display-only state from:

- `RunSummary` from `GET /runs`.
- `RunRecord` from `GET /runs/{id}`.
- Trace records from `GET /runs/{id}/trace`.

Recovery status is inferred from durable records that are no longer actively running after a daemon restart:

- `run.status === "running"` and no active local request id means "recovered".
- Any persisted `steps`, `control`, replay metadata, or trace events are treated as recovered runtime state for detail display.

Replay status is derived from `run.control[]` and `run.steps[]`:

- `control.action === "replay" && status === "approval_required"` or a `waiting_approval` step means replay approval is pending.
- `control.action === "replay" && status === "approved"` or step metadata with `replay_run_id` shows approved launch linkage.

Pause/resume status is derived from `control.action in ["pause","resume","cancel"]` and `steps[].id === "runtime-pause"`.

Trace availability is derived from the selected run trace response.

## Rendering

- Add Mission Control card: `Recovered`, counting durable running records not tied to the current active run.
- Add filter: `Recovered`.
- Add row badges for recovered, replay approval, paused/resumed, and trace-ready states.
- Add run detail sections:
  - `Runtime Recovery`: status, step count, control count, replay state.
  - `Workflow Steps`: compact rows with step status and metadata.
  - `Control History`: action/status/reason rows.
  - `Trace`: sanitized structured event summary loaded from `/runs/{id}/trace`.

## Safety

Trace UI renders only `TraceExportRecord` fields and sanitized `attributes`. It must not render legacy event args or request/response bodies in the trace section.

The existing Prompt/Result sections remain unchanged because this task is a runtime recovery display improvement, not a privacy refactor for historical run detail.

## Compatibility

If `/runs/{id}/trace` fails, the detail view still renders with a trace error message. If a run has no steps/control/trace, empty states are shown.

## Rollback

Remove Web UI state additions, trace fetch/render helpers, recovery badges/cards, related CSS, and tests.
