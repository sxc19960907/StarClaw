# Run follow-up suggestions

## Goal

Add run-completion follow-up suggestion actions to Chat run summaries and Mission Control run details.

## Requirements

- Reuse existing Chat run summary and Mission Control run detail state; do not add backend endpoints.
- Add a follow-up suggestion action to completed Chat run summaries.
- Add the same follow-up suggestion action to Mission Control run detail actions.
- The action should seed the Home mission composer with a concrete next prompt based on run id, prompt, status, session, result, and usage when available.
- Keep existing copy, rerun, open run, and open session behavior unchanged.

## Acceptance Criteria

- [x] Chat run summaries render a follow-up suggestion action.
- [x] Mission Control run details render a follow-up suggestion action when a run has a prompt.
- [x] Clicking either action seeds the Home mission composer with a run-derived next prompt.
- [x] Existing run summary and run detail actions still work.
- [x] Smoke or targeted tests verify at least one follow-up action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new suggestion API.
- No automatic second run.
- No replacement for Re-run; this action drafts a next prompt instead of repeating the same prompt.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
