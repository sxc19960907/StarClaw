# Review queue next actions

## Goal

Add a Home review queue that aggregates actionable workspace issues and routes users to the right panel.

## Requirements

- Reuse existing Web UI state only: runs, inbox, memory, permissions, diagnostics, and MCP/config data.
- Render a Home "Review Queue" / "Next Actions" area near the existing Focus Brief and Workspace Health Strip.
- Prioritize actionable items: failed or attention runs, pending or failed inbox items, memory warnings, diagnostics problems, permission/config risks, and MCP dock gaps.
- Each queue item should expose a concise title, status/risk label, short context, and a direct navigation action to the relevant panel.
- Render a calm empty state when no review items need attention.
- Refresh the queue whenever the underlying Home state sources update.
- Keep the visual direction aligned with Astria: dense independent workspace UI with subtle star/celestial styling, no marketing panel, no new frontend dependencies or build pipeline.

## Acceptance Criteria

- [x] Home renders a Review Queue / Next Actions section.
- [x] Queue derives items from at least four existing state sources.
- [x] Queue items navigate to relevant panels or existing run/session actions.
- [x] Empty or low-risk state is explicit and does not look broken.
- [x] Core smoke verifies section rendering and one navigation action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new backend aggregation endpoint.
- No persistent snooze/dismiss state.
- No replacement for the Runs, Inbox, Memory, Diagnostics, Permissions, or MCP panels.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
