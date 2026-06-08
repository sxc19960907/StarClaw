# Agent command launcher

## Goal

Make named agent custom commands launchable from the capability roster, turning saved command prompts into direct execution drafts.

## Requirements

- Show each agent's command names on the capability roster when commands exist.
- Clicking a command must select that agent in Chat and draft the saved command body without automatically sending it.
- Reuse the existing `/agents/{name}` detail API to load command bodies; the `/agents` list may expose command names but should not expose full command body text.
- Preserve existing command editor behavior.
- Keep the embedded static Web UI architecture with no new dependencies.

## Acceptance Criteria

- [x] `/agents` list summaries include command names for roster display.
- [x] Capability roster renders command launch chips for named agent commands.
- [x] Command launch chips switch to Chat, select the agent, and draft the command body.
- [x] Empty-command agents do not render launcher chips.
- [x] Agents smoke verifies command launch behavior for `smoke-agent`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
