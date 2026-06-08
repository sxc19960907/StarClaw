# Durable workflow step state

## Goal

Introduce durable step-level workflow state for launched missions, with status transitions and safe restart semantics.

## Requirements

- Add a local, serializable workflow step state model for daemon run records.
- Track step identity, display title, status, timestamps, attempt count, optional parent/sequence metadata, and redacted metadata.
- Support explicit status transitions for planned, running, blocked, waiting_approval, completed, failed, cancelled, and skipped steps.
- Persist step state through the existing optional `RunStore` persistence path so recovered runs keep their step history.
- Expose step state through existing run detail JSON without changing `/runs`, `/metrics`, `/cancel`, or replay execution behavior.
- Emit structured step events for safe observability while avoiding prompt text, tool arguments, provider payloads, secrets, and external side-effect data.
- Keep this child as state foundation only: no replay execution, no real pause/resume, and no frontend panel.

## Acceptance Criteria

- [x] `RunRecord` includes durable workflow step state in run detail JSON.
- [x] `RunStore` exposes APIs to upsert steps and transition step status without mutating run terminal status incorrectly.
- [x] Step state persists and recovers through `NewPersistentRunStore`.
- [x] Step structured events are redacted and counted in metrics only as aggregate event types.
- [x] Corrupt run-store recovery behavior remains safe after step fields are added.
- [x] Existing run/control/metrics and persistent run-store tests continue to pass.
- [x] No replay execution, cloud sync, database, or frontend build pipeline is introduced.

## Notes

- Parent task: `06-08-astria-phase-4-runtime-durability-replay`.
- This is the backend state contract needed before safe replay and real pause/resume children.
