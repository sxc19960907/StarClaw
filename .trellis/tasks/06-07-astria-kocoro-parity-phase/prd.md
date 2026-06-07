# Astria Kocoro Parity Phase

## Goal

Keep moving Astria toward a Kocoro-like independent agent workspace: workflows should feel launched, guided, and reviewable, not just exposed as disconnected panels.

## Requirements

- Continue preserving StarClaw CLI/module/package/release names.
- Keep the embedded static Web UI architecture.
- Prefer workflow cohesion and practical operator controls over decorative features.
- New work should connect existing capabilities: Home, Chat, Runs, MCP, Memory, Council, Inbox, File Intake, and schedules.

## Child Task Map

| Priority | Child Task | Purpose |
|---|---|---|
| P1 | `06-07-workflow-recipes-launcher` | Add guided Home recipes that prefill common Astria workflows and route users into the right panel or run. |
| P1 | `06-07-mission-control-run-board` | Upgrade Runs into a Mission Control board with status summaries and filters. |
| P1 | `06-07-workflow-brief-context` | Turn Home recipes into visible work briefs with outcome, context, route, and next checks. |
| P1 | `06-07-workspace-session-hub` | Add a Home workspace hub that summarizes session, run, memory, and file context. |
| P1 | `06-07-workflow-stage-continuity` | Add a Home workflow stage rail and keep run status grouping coherent across Home and Mission Control. |
| P1 | `06-07-command-center-palette` | Add an app-level Command Center for workflows, panel jumps, and workspace actions. |
| P1 | `06-07-recent-work-resume-rail` | Make recent sessions and runs resumable from Command Center and Home Workspace Hub. |

## Acceptance Criteria

- [x] Each child has testable PRD acceptance criteria.
- [x] Each implemented child passes Web UI smoke or targeted tests.
- [x] The phase improves Kocoro parity by making existing capabilities easier to launch as workflows.

## Non-Goals

- No frontend build pipeline.
- No cloud relay.
- No repo/package rename to Astria.
