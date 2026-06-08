# Astria Kocoro parity phase 2

## Goal

Continue Astria toward Kocoro parity after the first workflow-heavy phase by making successful work reusable across future missions.

## Requirements

- Preserve StarClaw CLI/module/package/release names.
- Keep the embedded static daemon Web UI architecture; no frontend build pipeline.
- Use Kocoro's current product direction as inspiration: local agent workspace, reusable agents/prompts/knowledge, saved results, team-style shared assets, and successful work becoming the next starting point.
- Add independently verifiable child tasks that turn existing Astria capabilities into reusable workspace assets rather than disconnected operational panels.

## Child Task Map

| Priority | Child Task | Purpose |
|---|---|---|
| P1 | `06-08-reuse-gallery` | Add an Astria Reuse Gallery that turns prompts, agents, knowledge sources, and run outcomes into reusable launch assets. |
| P1 | `06-08-browser-mission-planner` | Add a Browser Mission Planner for reviewed web inspection, screenshot, extraction, form-check, and monitoring mission starters. |
| P1 | `06-08-data-insight-planner` | Add a Data Insight Planner for reviewed local file, table, metric, and export analysis mission starters. |
| P1 | `06-08-share-pack-builder` | Add a Share Pack Builder for local reviewed handoff packages that make successful work reusable by future sessions or reviewers. |
| P1 | `06-08-starter-kit-launcher` | Add a Starter Kit Launcher for prebuilt local workflows that map common tasks to prompts, routes, evidence gates, and reusable outputs. |
| P1 | `06-08-citation-grounding-planner` | Add a Citation Grounding Planner for source coverage, claim maps, quote capture, freshness checks, and evidence gap escalation. |
| P1 | `06-08-result-library-report-archive` | Add a Result Library that archives completed reports, evidence briefs, insight summaries, citation briefs, reusable outputs, and council synthesis for follow-up. |
| P1 | `06-08-playbook-library-best-practices` | Add a Playbook Library that turns successful local work patterns into reviewed best-practice launch paths. |

## Acceptance Criteria

- [x] Each child has testable PRD acceptance criteria.
- [x] Each implemented child passes Web UI smoke or targeted tests.
- [ ] The phase improves Kocoro parity by making successful Astria work directly reusable as future mission starters, turning local data into reviewed insight briefs, packaging work into local handoff packs, offering prebuilt starter workflows, planning reliable citations, archiving saved results for follow-up, and capturing repeatable best practices as playbooks.

## Non-Goals

- No frontend build pipeline.
- No cloud team sync or account system.
- No repo/package rename to Astria.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
