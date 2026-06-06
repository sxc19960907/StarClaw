# Astria Product Polish and External Channels

## Goal

Advance Astria from MVP feature surfaces into a cohesive local-agent product phase. This phase should polish the current embedded Web UI, connect the guarded inbox to real external channels, and turn Council, Memory, MCP, and document tools into workflows that feel connected instead of separate panels.

## Confirmed Facts

- The first Astria roadmap is complete and committed in `3142d9c`.
- Product-facing name remains Astria, while CLI/module/package/release artifacts remain StarClaw.
- The Web UI is a static embedded daemon app; this phase must not add React/Vite or external runtime assets unless a later child explicitly justifies it.
- The preferred design direction remains Kocoro-like: calm native desktop app, light sidebar, central task entry, practical panels, subtle celestial identity.
- Existing implemented surfaces include Home Actions, MCP Starport, document/archive tools, Memory Map, Agent Council, and Channel Inbox.

## Requirements

- Maintain a parent/child Trellis task tree for the second Astria phase.
- Prioritize product coherence before adding provider complexity.
- Keep all external-channel execution guarded by Inbox approval; no unattended execution by default.
- Keep new UI controls useful and dense, not marketing-like or dashboard-heavy.
- Preserve static embedded Web UI architecture.
- Each child must remain independently testable and shippable.

## Child Task Map

| Priority | Child Task | Purpose | Notes |
|---|---|---|---|
| P1 | `06-06-astria-ui-polish` | Smooth visual/copy/state consistency across the Astria shell. | First implementation target. |
| P1 | `06-06-real-channel-provider` | Connect Inbox to one real provider. | Provider choice should depend on available credentials and lowest setup risk. |
| P2 | `06-06-council-workflow-handoff` | Turn Council synthesis into runnable tasks or handoffs. | Builds on existing Council MVP. |
| P2 | `06-06-memory-taxonomy` | Add memory categories and conflict/review governance. | Keeps long-term memory useful. |
| P2 | `06-06-mcp-config-editor` | Let users edit MCP server config from Astria. | Needs careful validation and secret handling. |
| P3 | `06-06-file-intake-ui` | Add file intake UI for document/archive tools. | Uses existing local tools first. |

## Acceptance Criteria

- [x] Parent task lists planned children and priority order.
- [x] Each child task has a PRD with goals, requirements, acceptance criteria, non-goals, and dependencies.
- [x] Complex children have `design.md` and `implement.md` before implementation starts.
- [x] UI polish child is implemented and verified as the first slice.
- [x] Integration checks include `go test ./...`, JS syntax check, diff whitespace check, and Web UI smoke.

## Non-Goals

- No repo-wide rename to Astria.
- No hosted cloud relay.
- No unapproved external-channel execution.
- No frontend build pipeline in the parent task.
