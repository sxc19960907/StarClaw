# Astria Kocoro parity phase 2

## Goal

Continue Astria toward Kocoro parity after the first workflow-heavy phase by making successful work reusable across future missions.

## Requirements

- Preserve StarClaw CLI/module/package/release names.
- Keep the embedded static daemon Web UI architecture; no frontend build pipeline.
- Use Kocoro's current product direction as inspiration: local agent workspace, reusable agents/prompts/knowledge, saved results, team-style shared assets, and successful work becoming the next starting point.
- Add independently verifiable child tasks that turn existing Astria capabilities into reusable workspace assets rather than disconnected operational panels.

## Child Task Map

| Priority | Child Task | Purpose |
|---|---|---|
| P1 | `06-08-reuse-gallery` | Add an Astria Reuse Gallery that turns prompts, agents, knowledge sources, and run outcomes into reusable launch assets. |
| P1 | `06-08-browser-mission-planner` | Add a Browser Mission Planner for reviewed web inspection, screenshot, extraction, form-check, and monitoring mission starters. |

## Acceptance Criteria

- [x] Each child has testable PRD acceptance criteria.
- [x] Each implemented child passes Web UI smoke or targeted tests.
- [ ] The phase improves Kocoro parity by making successful Astria work directly reusable as future mission starters.

## Non-Goals

- No frontend build pipeline.
- No cloud team sync or account system.
- No repo/package rename to Astria.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
