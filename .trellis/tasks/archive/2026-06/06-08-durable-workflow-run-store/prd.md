# Durable workflow run store

## Goal

Persist daemon run records, structured events, and control decisions to local disk so Astria Mission Control can recover recent workflow state after daemon restart.

## Requirements

- Store run records locally under StarClaw's existing config/data root; do not introduce a database or cloud dependency.
- Preserve the existing in-memory `RunStore` API and current `/runs`, `/runs/{id}`, `/metrics`, `/cancel`, and `/runs/{id}/control` behavior.
- Persist enough data to recover run summaries, details, structured events, control decisions, usage, budget, routing, fallback, response, and terminal status.
- Keep prompt/request fields compatible with current run detail behavior, but do not add prompts to metrics or trace exports.
- Use atomic or temp-file write semantics so a partial write does not corrupt the run store.
- Respect `defaultRunStoreLimit` when loading or writing recovered runs.
- Add tests covering restart recovery, corrupt-file tolerance, limit enforcement, control metadata recovery, and current API compatibility.

## Acceptance Criteria

- [x] `RunStore` can be constructed with a local persistence path and recover records after a new store is created.
- [x] Run summaries and run detail include persisted structured events and control decisions after reload.
- [x] Corrupt or unreadable persistence data does not crash store construction; it starts with an empty store and reports a safe error path in tests.
- [x] Store limit is enforced during recovery and subsequent writes.
- [x] Existing daemon run/control/metrics tests continue to pass.
- [x] No database, cloud sync, or frontend build pipeline is introduced.

## Non-Goals

- No step graph execution yet.
- No real pause/resume implementation.
- No replay execution; replay remains approval-required planning only.
- No new Web UI panel in this slice.
