# Astria Product Roadmap Inspired by Kocoro

## Goal

Plan Astria's next product phase as a set of independently deliverable Trellis child tasks. The roadmap should keep the Kocoro-style desktop-app feel the user prefers, add Astria's own subtle constellation/star identity, and close the highest-value StarClaw capability gaps without turning the product into a dashboard-heavy admin UI.

## Confirmed Facts

- The product-facing brand should be Astria.
- Current implementation work already created an Astria Web UI home launcher task: `06-06-astria-webui-home`.
- Kocoro Desktop GUI is the main UI reference: macOS-native layout, light sidebar, large central task input, clean panels, and practical local-agent controls.
- Kocoro feature inspiration includes local tools and permissions, channel messaging, MCP-native workflows, schedules, memory/session sync, document/archive tools, and `/swarm` multi-agent workflows.
- Existing StarClaw/ShanClaw gap research identifies major gaps in browser/GUI automation, agent streaming/thinking, daemon mode, MCP server capability, skill system, document-like tool breadth, and scheduling/channel workflows.

## Requirements

- Maintain a parent/child Trellis task tree where the parent owns roadmap intent, sequencing, cross-child design constraints, and final integration review.
- Track child tasks for the next product phase:
  - Astria Home Actions and live activity center.
  - Document and archive tools pack.
  - MCP Starport management UI.
  - Memory Map MVP.
  - Agent Council / swarm workflows.
  - Channel Inbox MVP.
- Preserve the existing product boundary unless a child explicitly expands it: do not rename CLI/module/package/release artifacts from StarClaw to Astria as part of this roadmap.
- Preserve the current Web UI static embedded architecture unless a child explicitly justifies a build step.
- Keep Astria's visual direction close to Kocoro's calm native-app surface, with subtle star/constellation identity rather than dark sci-fi styling.
- Each child must be independently testable, shippable, and archivable.

## Acceptance Criteria

- [x] Parent task lists all planned child tasks and the intended sequencing.
- [x] Each child task has a PRD with goals, requirements, acceptance criteria, non-goals, and dependencies.
- [x] Complex child tasks identify whether `design.md` and `implement.md` are required before development starts.
- [x] The roadmap explicitly separates immediate UI/product work from deeper backend/tooling work.
- [x] The roadmap names integration checks that must pass before the parent can be considered complete.

## Non-Goals

- No implementation work in this parent task.
- No repo-wide rename from StarClaw to Astria.
- No new frontend dependency or build pipeline decision in the parent task.
- No commitment to a specific external messaging provider until the Channel Inbox child task is planned.

## Child Task Map

| Priority | Child Task | Purpose | Depends On |
|---|---|---|---|
| P1 | `06-06-astria-home-actions` | Make Astria Home operational, not decorative. | `06-06-astria-webui-home` completion |
| P1 | `06-06-document-archive-tools` | Add high-value local file understanding tools. | Tool registry patterns |
| P1 | `06-06-mcp-starport-ui` | Make MCP setup visible and manageable from UI. | Existing MCP client/config APIs |
| P2 | `06-06-memory-map-mvp` | Add reviewed reusable memory from sessions. | Session storage/indexing clarity |
| P2 | `06-06-agent-council-swarm` | Add multi-agent planning/review workflows. | Agent/run orchestration contracts |
| P3 | `06-06-channel-inbox-mvp` | Add first external inbox and approval handoff. | Daemon/background execution maturity |

## Integration Acceptance

- Astria Home can expose real entry points for implemented children without visual redesign each time.
- New capabilities appear as first-class product surfaces, not only hidden CLI commands.
- Existing smoke tests for Web UI and daemon flows remain stable.
- Feature flags or disabled states are used for incomplete children instead of broken navigation.
