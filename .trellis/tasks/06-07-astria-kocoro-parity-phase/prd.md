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
| P1 | `06-07-focus-brief-current-mission` | Add a Home Focus Brief that summarizes current mission context and next action. |
| P1 | `06-07-workspace-health-strip` | Add a Home readiness strip for diagnostics, permissions, MCP, and memory health. |
| P1 | `06-07-06-07-review-queue-next-actions` | Add a Home review queue that turns scattered workspace risks into direct next actions. |
| P1 | `06-07-06-07-strategy-matrix-workflow-modes` | Add an Astria Strategy Matrix for choosing Kocoro/Shannon-style execution modes before launch. |
| P1 | `06-07-06-07-run-timeline-time-travel` | Add a Mission Control Time Travel timeline for replay-like run review. |
| P1 | `06-07-06-07-approval-center-control-console` | Add an Approval Center for human-in-the-loop risks, recovery, and policy actions. |
| P1 | `06-07-06-07-knowledge-curation-console` | Add a Knowledge Curation console for reviewable long-term memory candidates. |
| P1 | `06-07-06-07-mcp-capability-inspector` | Add a Tool Dock Inspector for MCP transports, env keys, readiness flags, and no-dock recovery. |
| P1 | `06-07-06-07-prompt-suggestion-dock` | Add Kocoro-style next prompt suggestions derived from current Astria workspace state. |
| P1 | `06-07-06-07-run-follow-up-suggestions` | Add run-completion follow-up suggestion actions to Chat summaries and Mission Control details. |
| P1 | `06-08-06-07-agent-capability-roster` | Add an Agents capability roster for model, tools, memory, approval, heartbeat, and commands. |
| P1 | `06-08-06-08-agent-launch-actions` | Add direct Chat, Test, and Council launch actions from each Agent capability roster card. |
| P1 | `06-08-06-08-agent-command-launcher` | Add slash-command chips on Agent roster cards that draft saved command bodies into Chat. |
| P1 | `06-08-06-08-agent-continuity-digest` | Add per-agent continuity digest for recent runs, memory posture, commands, and next actions. |
| P1 | `06-08-council-stage-workflow` | Turn Agent Council detail into a staged planner/researcher/reviewer/synthesis/handoff workflow. |
| P1 | `06-08-comparison-workbench` | Add a side-by-side workbench for comparing runs, agents, memory, and council evidence. |
| P1 | `06-08-proactive-delivery-board` | Add a proactive delivery board for schedules, outbound runs, channels, and recovery readiness. |
| P1 | `06-08-prompt-experiment-lab` | Add a Prompt Lab for direct, evidence-first, council, and delivery-ready prompt variants. |
| P1 | `06-08-knowledge-source-registry` | Add a Source Registry for freshness, reliability, and maintenance of Astria knowledge sources. |

## Acceptance Criteria

- [x] Each child has testable PRD acceptance criteria.
- [x] Each implemented child passes Web UI smoke or targeted tests.
- [x] The phase improves Kocoro parity by making existing capabilities easier to launch as workflows.

## Non-Goals

- No frontend build pipeline.
- No cloud relay.
- No repo/package rename to Astria.
