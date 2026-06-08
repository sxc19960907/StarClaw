# Prompt experiment lab

## Goal

Add a Kocoro-style Prompt Experiment Lab where Astria can turn one goal into multiple prompt variants, compare their agent/context fit, and draft the chosen experiment into Chat.

## Requirements

- Keep StarClaw internal naming and the embedded static Web UI architecture.
- Do not add a backend API or frontend build pipeline.
- Add a Prompt Lab panel reachable from sidebar navigation and Manage.
- The lab must derive variant recommendations from existing Web UI state: agents, memory, runs, council, comparison lanes, and delivery lanes.
- The lab must provide at least four prompt variants: direct execution, evidence-first, council review, and delivery-ready.
- Each variant must show target agent/context/source, risk posture, evaluation criteria, and direct actions.
- Actions must draft the variant prompt into Chat or open the relevant source panel.
- Empty state must remain useful with default variants even when no runs or memory exist.

## Acceptance Criteria

- [x] Prompt Lab is reachable from sidebar navigation and Manage.
- [x] The panel renders direct, evidence-first, council, and delivery variants.
- [x] Each variant shows agent/context/source, risk, and evaluation criteria.
- [x] A variant can draft a prompt into Chat.
- [x] A variant can open its source panel.
- [x] Web UI smoke verifies panel reachability, variant rendering, Chat draft, and source routing.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
