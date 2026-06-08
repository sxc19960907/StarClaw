# Phase 5 runtime E2E smoke

## Goal

Validate the main Astria runtime path end to end across Home/Chat launch, run history, workflow control, replay approval, pause/resume, trace, and recovery display.

## Requirements

- Exercise completed Phase 3/4 platform capabilities together instead of only unit-level checks.
- Prefer deterministic daemon/browser smoke fixtures over paid provider calls.
- Cover Mission Control and run detail states that a user would inspect after long-running work.
- Preserve local-only behavior and existing embedded Web UI architecture.

## Acceptance Criteria

- [ ] Smoke path launches a run through local daemon/UI-compatible API and verifies it appears in `/runs`.
- [ ] Smoke path verifies budget/routing/fallback fields remain renderable when present.
- [ ] Smoke path covers workflow control responses for cancel, pause/resume, and replay approval boundaries.
- [ ] Smoke path verifies trace/recovery UI hooks are present and tolerate recovered run state.
- [ ] Targeted daemon/Web UI tests and full `go test ./...` pass after fixes.

## Non-Goals

- No new runtime semantics.
- No real external LLM or cloud dependency.
- No broad UI redesign.

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
