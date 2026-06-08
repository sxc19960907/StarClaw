# Knowledge source registry

## Goal

Add a Kocoro-style Knowledge Source Registry so Astria can inspect where its working knowledge comes from, how fresh it is, and what maintenance action should happen next.

## Requirements

- Keep StarClaw internal naming and the embedded static Web UI architecture.
- Do not add a backend API or frontend build pipeline.
- Add a Source Registry panel reachable from sidebar navigation and Manage.
- Derive source rows from existing Web UI state: memory entries, sessions, runs, file intake, and council runs.
- Each source must show type, evidence count, freshness/recency, reliability posture, and a maintenance action.
- Actions must draft a source maintenance prompt into Chat or open the relevant source panel.
- Empty state must still explain how to seed the first trusted source from Memory, Runs, File Intake, or Council.

## Acceptance Criteria

- [x] Source Registry is reachable from sidebar navigation and Manage.
- [x] The registry renders source rows for memory, runs, sessions, file intake, and council state.
- [x] Each source row shows evidence count, freshness, reliability, and maintenance action.
- [x] A source row can draft a maintenance prompt into Chat.
- [x] A source row can open its source panel.
- [x] Web UI smoke verifies registry reachability, source rendering, Chat draft, and source routing.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
