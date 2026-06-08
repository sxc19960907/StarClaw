# Runtime recovery UI

## Goal

Surface recovered durable runs, restart state, replay approval state, pause/resume state, and trace availability in Astria Mission Control.

## Requirements

- Preserve the embedded static Web UI architecture under `internal/daemon/webui/`.
- Keep StarClaw CLI/module/package names; Astria is the product-facing UI name.
- Show recovery-oriented signals in the Runs / Mission Control view without requiring a new frontend build pipeline.
- Identify runs restored from durable storage after daemon start.
- Highlight pending replay approvals and approved replay links.
- Show pause/resume/cancel control history and workflow step state for a selected run.
- Provide trace visibility from the run detail view using the existing local trace endpoint.
- Do not expose prompt text, tool arguments, provider payloads, API keys, tokens, or secrets through new trace/recovery UI elements.
- Keep runtime behavior unchanged; this task is UI and read-only display except for existing run control actions.

## Acceptance Criteria

- [x] Mission Control includes a recovery summary card/count for restored durable runs.
- [x] Run rows show compact recovery, replay, pause/resume, and trace badges when those states exist.
- [x] Run detail shows a runtime recovery section with restart/recovered status, workflow steps, and control history.
- [x] Run detail exposes a trace section populated from `GET /runs/{id}/trace`.
- [x] Trace display uses sanitized structured event fields and does not render prompt/tool args/provider payloads/secrets.
- [x] Existing run list/detail behavior remains compatible when the trace endpoint is unavailable or returns no data.
- [x] Web UI smoke/static tests and daemon tests continue to pass.

## Non-Goals

- No new runtime recovery semantics.
- No replay execution changes.
- No external telemetry collector or cloud trace upload.
- No frontend build pipeline.

## Goal

TBD.

## Requirements

- TBD

## Acceptance Criteria

- [ ] TBD

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
