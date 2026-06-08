# Browser mission planner

## Goal

Add a Kocoro-style Browser Mission Planner so Astria can turn browser work into reviewed mission starters: web inspection, screenshot evidence, extraction, form checks, and change monitoring should be planned before Chat launches a run.

## Requirements

- Keep StarClaw internal naming and the embedded static Web UI architecture.
- Do not add a backend API, browser execution endpoint, account system, or frontend build pipeline.
- Add a Browser Mission Planner panel reachable from sidebar navigation and Manage.
- Derive deterministic browser mission templates from existing capabilities: browser navigation/title, screenshot/computer evidence, local file intake context, inbox/webhook work, and delivery/readiness checks.
- Let the operator enter a target URL and mission goal; templates should use those fields to build reviewed starter prompts.
- Each mission card must show mission type, evidence/readiness, risk/approval posture, and next action.
- Actions must draft a browser mission starter into Chat or open the relevant source panel.
- Keep the design dense, operational, and subtly celestial; avoid marketing or dashboard-heavy presentation.
- Browser plans must explicitly avoid submitting forms, changing accounts, or performing destructive webpage actions without approval.

## Acceptance Criteria

- [x] Browser Mission Planner is reachable from sidebar navigation and Manage.
- [x] The planner renders browser mission cards for inspect, screenshot, extraction, form-check, and monitor workflows.
- [x] Target URL and goal fields update generated mission prompts.
- [x] Each card shows type, evidence/readiness, risk/approval posture, and next action.
- [x] A mission card can draft a browser starter prompt into Chat.
- [x] A mission card can open its source panel.
- [x] Web UI smoke verifies planner reachability, card rendering, goal/URL prompt generation, Chat draft, and source routing.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
