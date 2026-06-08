# Knowledge conflict reconciliation

## Goal

Add an Astria Knowledge Reconciliation panel that identifies stale, conflicting, weakly sourced, or privacy-sensitive knowledge before reuse.

## Requirements

- Keep the implementation inside the embedded static daemon Web UI.
- Do not add backend storage, external assets, cloud sync, accounts, or a frontend build pipeline.
- Use existing local state from Source Registry, Memory Map, Citation Planner, Result Library, Playbook Library, runs, sessions, and file intake.
- Present reconciliation items for conflicts, stale sources, weak citations, duplicate/uncategorized memory, missing source coverage, and privacy or approval blockers.
- Each reconciliation item must show risk, evidence, resolution action, route, and a draft prompt for Chat.
- Preserve Astria product naming in UI copy while keeping StarClaw repo/package naming unchanged.

## Acceptance Criteria

- [x] The sidebar and Manage hub expose a Knowledge Reconciliation panel with live counts.
- [x] Knowledge Reconciliation shows cards for source conflict, stale memory, weak citation, duplicate memory, missing source coverage, privacy boundary, and result freshness review.
- [x] Selecting a reconciliation card renders a detail brief with risk, evidence, resolution action, confidence boundary, and route.
- [x] Each card can draft a resolution prompt into Chat and route back to the relevant source panel.
- [x] Web UI smoke covers opening Knowledge Reconciliation, verifying cards/detail content, drafting a prompt, and route buttons.
- [x] Existing validation remains green: `node --check internal/daemon/webui/assets/app.js`, `git diff --check`, `go test ./internal/daemon`, `./scripts/smoke_webui_core.sh`, and `go test ./...`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
