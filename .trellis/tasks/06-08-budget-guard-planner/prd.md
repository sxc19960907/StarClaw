# Budget guard planner

## Goal

Add an Astria Budget Guard planning surface to the embedded daemon Web UI so users can plan local token, model, fallback, complexity, and stop-rule boundaries before launching expensive or long-running work.

## Requirements

- Expose Budget Guard from the sidebar and Manage hub.
- Summarize budget plans across model routing, token caps, context trimming, fallback, long-run stop rules, schedule limits, and evidence-cost tradeoffs.
- Each budget card must show budget shape, trigger, guardrail, fallback route, and review boundary.
- Each card must draft a Chat prompt that asks Astria to plan within budget limits.
- Each card must route back to a relevant existing panel such as Chat, Runs, Prompt Lab, Agents, Schedules, Snapshot, or Diagnostics.
- Keep this as static embedded daemon Web UI only. Do not add backend accounting, billing, cloud sync, or a frontend build pipeline.

## Acceptance Criteria

- [x] Sidebar and Manage hub expose Budget Guard with live counts.
- [x] Budget Guard panel renders multiple budget planning cards and a selected detail brief.
- [x] Detail brief includes budget shape, trigger, guardrail, fallback, and review boundary.
- [x] Draft actions populate Chat with a budget-aware planning prompt.
- [x] Route actions open relevant source panels.
- [x] Web UI smoke coverage verifies panel render, card selection, draft behavior, and routing.

## Notes

- Product-facing copy should use Astria. Internal repo/module naming stays StarClaw.
