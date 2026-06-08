# Agent launch actions

## Goal

Make named agents directly launchable from the capability roster so the operator can move from inspection to execution without hunting through separate controls.

## Requirements

- Add roster actions for chatting with an agent, preparing a test run, and seeding an Agent Council goal with that agent selected.
- Reuse existing Chat, Agent Test Runner, and Council UI flows; do not add new daemon endpoints.
- Preserve existing edit behavior and StarClaw internal naming.
- Keep actions compact and operational inside the current static Web UI.

## Acceptance Criteria

- [x] Each capability roster card exposes Chat, Test, Council, and Edit actions.
- [x] Chat action switches to Chat with the selected agent and a useful prompt draft.
- [x] Test action switches/focuses the Agent Test Runner with the selected agent and a useful prompt draft.
- [x] Council action switches to Agent Council with the selected agent and a useful goal draft.
- [x] Agents smoke verifies launch actions for `smoke-agent`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
