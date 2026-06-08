# Reuse gallery

## Goal

Add a Kocoro-style Reuse Gallery so Astria can turn successful local work into reusable mission starters: prompts, agent profiles, knowledge sources, and run outcomes should be visible in one operational library and draftable into Chat.

## Requirements

- Keep StarClaw internal naming and the embedded static Web UI architecture.
- Do not add a backend API, persistence layer, account system, or frontend build pipeline.
- Add a Reuse Gallery panel reachable from sidebar navigation and Manage.
- Derive reusable asset cards from existing Web UI state:
  - prompt variants from Prompt Lab
  - agent profiles and saved commands
  - knowledge source rows from Source Registry
  - recent runs and council output
- Each asset must show kind, readiness/evidence, reuse value, and a next action.
- Actions must draft a reusable starter prompt into Chat or open the relevant source panel.
- Empty/base state must still show useful starter assets from deterministic Prompt Lab and Source Registry rows.
- Keep the design dense, operational, and subtly celestial; avoid marketing or dashboard-heavy presentation.

## Acceptance Criteria

- [x] Reuse Gallery is reachable from sidebar navigation and Manage.
- [x] The gallery renders reusable asset cards for prompt, agent, knowledge source, run, and council categories when data exists.
- [x] Base state still renders deterministic prompt and knowledge source cards when runs or agents are empty.
- [x] Each card shows kind, readiness/evidence, reuse value, and next action.
- [x] A gallery card can draft a reusable starter prompt into Chat.
- [x] A gallery card can open its source panel.
- [x] Web UI smoke verifies gallery reachability, asset rendering, Chat draft, and source routing.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
