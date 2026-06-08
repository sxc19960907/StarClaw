# Proactive delivery board

## Goal

Add a Kocoro-style proactive delivery board that makes scheduled Astria work feel like a monitored outbound workflow, not just a cron form. Operators should be able to see what will run, what recently delivered, which channels are ready, and what to draft next.

## Requirements

- Keep StarClaw internal naming and the embedded static Web UI architecture.
- Do not add backend delivery integrations or external network calls in this slice.
- Add a Proactive Delivery Board panel reachable from sidebar navigation and Manage.
- Derive board data from existing Web UI state: schedules, runs, inbox providers/items, diagnostics, and config.
- Show delivery lanes for scheduled work, recent outbound runs, channel readiness, and recovery actions.
- Each lane must show readiness, evidence, risk, and a direct action.
- Actions must draft a delivery/check prompt into Chat or open the relevant source panel.
- Empty state must still help users create their first scheduled delivery plan.

## Acceptance Criteria

- [x] Proactive Delivery Board is reachable from sidebar navigation and Manage.
- [x] Board renders schedule, recent delivery, channel readiness, and recovery lanes.
- [x] Each lane shows readiness, evidence, risk, and next action text.
- [x] A lane can draft a delivery prompt into Chat.
- [x] A lane can open its source panel.
- [x] Web UI smoke verifies board reachability, lane rendering, Chat draft, and source routing.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
