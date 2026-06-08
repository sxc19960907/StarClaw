# Result library report archive

## Goal

Add an Astria Result Library panel that collects completed reports and reusable output briefs into a local, reviewable archive surface.

## Requirements

- Keep the implementation inside the embedded static daemon Web UI.
- Do not add backend storage, external assets, cloud sync, accounts, or a frontend build pipeline.
- Present result archive entries as durable local outcomes that can be reviewed, routed back to source panels, or drafted into Chat as follow-up missions.
- Cover Kocoro-inspired saved results without implying cloud/team sync.
- Preserve Astria product naming in UI copy while keeping StarClaw repo/package naming unchanged.

## Acceptance Criteria

- [x] The sidebar and Manage hub expose a Result Library panel with live counts.
- [x] Result Library shows archived output cards for recent runs, share packs, data insights, citation briefs, reuse assets, and council outcomes when available.
- [x] Selecting a result renders a detail brief with source, evidence, reuse path, freshness or review posture, and next action.
- [x] Each result can draft a follow-up prompt into Chat and can route back to the relevant source panel.
- [x] Web UI smoke covers opening Result Library, selecting/drafting a result, and route buttons.
- [x] Existing validation remains green: `node --check internal/daemon/webui/assets/app.js`, `git diff --check`, `go test ./internal/daemon`, `./scripts/smoke_webui_core.sh`, and `go test ./...`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
