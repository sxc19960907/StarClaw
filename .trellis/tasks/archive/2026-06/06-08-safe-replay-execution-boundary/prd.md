# Safe replay execution boundary

## Goal

Convert replay plans into approved replay launches while preserving destructive/external approval gates.

## Requirements

- Preserve the current safe default: `POST /runs/{id}/control` with `action="replay"` and no approval must return an approval-required plan only.
- Support explicitly approved replay launches via `approved=true`.
- Approved replay must create a new run record with a new replay run id, link it to the source run through control metadata and step state, and execute through the existing daemon run path.
- Approved replay must preserve tool/external approval gates by using the existing daemon approval requester and agent loop permissions.
- Replay response must redact the source prompt in planning responses and must not expose tool args, provider payloads, or external side effects.
- Replay launches must not mutate the source run's terminal status.
- Replay must remain local-first and must not introduce cloud sync, database state, or frontend build tooling.

## Acceptance Criteria

- [x] Unapproved replay still returns `approval_required`, redacted source request metadata, and no new run.
- [x] Approved replay launches a new run through the normal daemon execution path.
- [x] Approved replay response includes source run id and replay run id.
- [x] Source run records both the approval-required/approved control decision and an immutable link to the replay run.
- [x] Replay run detail records source/replay metadata without leaking prompt text through control response, metrics, or structured events.
- [x] Existing tool approval behavior is preserved because replay uses the same `s.runAgent` path.
- [x] Validation covers missing run, missing action, unsupported actions, and unapproved vs approved replay.
- [x] Existing run/control/metrics and full daemon tests continue to pass.
- [x] No cloud sync, database, or frontend build pipeline is introduced.

## Notes

- Parent task: `06-08-astria-phase-4-runtime-durability-replay`.
- This child implements the explicit replay execution boundary, not full deterministic side-effect replay.
