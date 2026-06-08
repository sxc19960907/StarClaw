# Playbook library best practices

## Goal

Add an Astria Playbook Library that turns successful local work patterns into reviewed, reusable best-practice cards.

## Requirements

- Keep the implementation inside the embedded static daemon Web UI.
- Do not add backend storage, external assets, cloud sync, accounts, or a frontend build pipeline.
- Present playbooks as local best-practice assets derived from existing Astria surfaces: starter kits, result archive, reuse gallery, citation planner, browser/data planners, share packs, and council review.
- Each playbook must show when to use it, evidence gate, safety boundary, reusable output, and follow-up route.
- Preserve Astria product naming in UI copy while keeping StarClaw repo/package naming unchanged.

## Acceptance Criteria

- [x] The sidebar and Manage hub expose a Playbook Library panel with live counts.
- [x] Playbook Library shows best-practice cards for research, data insight, handoff, citation, agent profile, memory curation, delivery, and browser review patterns.
- [x] Selecting a playbook renders a detail brief with trigger, steps, evidence gate, safety boundary, reusable output, and next route.
- [x] Each playbook can draft a best-practice launch prompt into Chat and route back to the relevant source panel.
- [x] Web UI smoke covers opening Playbook Library, selecting/drafting a playbook, and route buttons.
- [x] Existing validation remains green: `node --check internal/daemon/webui/assets/app.js`, `git diff --check`, `go test ./internal/daemon`, `./scripts/smoke_webui_core.sh`, and `go test ./...`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
