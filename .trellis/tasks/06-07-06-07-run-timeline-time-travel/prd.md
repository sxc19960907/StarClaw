# Run timeline time travel

## Goal

Add a Mission Control run timeline that makes executions reviewable and replay-like from existing run detail data.

## Requirements

- Reuse existing runs, current run detail, tool event, approval, and session state; do not add backend endpoints.
- Add a Run Timeline / Time Travel surface inside the Runs panel or selected run detail area.
- Timeline should show major milestones: queued/started status, prompt/agent, tool events, approval checkpoints, completion/failure, and linked session when available.
- The selected run should expose a compact timeline even when detailed events are sparse.
- Timeline items should use direct actions where existing UI supports them, such as opening the linked session or run detail.
- Keep styling dense and operational, aligned with Astria Mission Control.
- Preserve StarClaw internal names and embedded static Web UI architecture.

## Acceptance Criteria

- [x] Runs panel renders a Run Timeline for the selected or latest run.
- [x] Timeline derives from at least three existing state sources or run fields.
- [x] Timeline has an explicit sparse/empty state.
- [x] Timeline actions navigate to existing run/session views where available.
- [x] Core smoke verifies timeline rendering and one action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new run event persistence model.
- No backend replay endpoint.
- No destructive rollback/time rewind operation.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
