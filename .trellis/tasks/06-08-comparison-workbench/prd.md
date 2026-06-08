# Comparison workbench

## Goal

Add a Kocoro-style comparison workbench where Astria can compare recent runs, agent profiles, memory context, and council output side by side before the operator chooses the next path.

## Requirements

- Keep StarClaw internal naming and the embedded static Web UI architecture.
- Do not add a backend API or frontend build pipeline.
- Add an Astria Comparison Workbench panel reachable from navigation and Manage.
- The workbench must derive comparison candidates from existing in-memory Web UI state: runs, agents, memory, and council runs.
- Each candidate must show source, evidence count, recency/readiness, tradeoffs, and why it may be the better next move.
- The panel must provide direct actions to draft a comparison prompt in Chat and to open the source panel.
- Empty state must remain useful by pointing users to Chat, Runs, Agents, Memory, and Council.

## Acceptance Criteria

- [x] Comparison Workbench is reachable from sidebar navigation and Manage.
- [x] The panel renders at least three side-by-side lanes when the smoke fixture has runs, agents, memory, and council data.
- [x] Each lane shows evidence, tradeoff, and recommendation text.
- [x] A lane can draft a comparison prompt into Chat.
- [x] A lane can open its source panel.
- [x] Web UI smoke verifies panel reachability, lane rendering, Chat draft, and source routing.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
