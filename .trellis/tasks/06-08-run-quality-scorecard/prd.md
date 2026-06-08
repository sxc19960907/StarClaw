# Run quality scorecard

## Goal

Add an Astria Run Quality Scorecard planning surface to the embedded daemon Web UI so users can evaluate recent runs by completion quality, evidence strength, budget posture, risk, and recommended next action.

## Requirements

- Expose Run Quality from the sidebar and Manage hub.
- Summarize recent run quality across completion status, evidence coverage, budget/stop-rule posture, retry risk, and reuse readiness.
- Provide multiple scorecard lanes for latest run, completed outputs, failed runs, evidence quality, budget risk, reusable output, and delivery readiness.
- Each scorecard must show score, signal, risk, review gate, and recommended route.
- Each scorecard must draft a Chat prompt that asks Astria to evaluate or improve the run.
- Each scorecard must route back to a relevant existing panel such as Runs, Compare, Result Library, Budget Guard, Citation Planner, Share Pack, or Delivery.
- Keep this as static embedded daemon Web UI only. Do not add backend scoring storage, telemetry collection, cloud sync, or a frontend build pipeline.

## Acceptance Criteria

- [x] Sidebar and Manage hub expose Run Quality with live counts.
- [x] Run Quality panel renders scorecard lanes and a selected detail brief.
- [x] Detail brief includes score, signal, risk, review gate, and recommended route.
- [x] Draft actions populate Chat with a run-quality evaluation prompt.
- [x] Route actions open relevant source panels.
- [x] Web UI smoke coverage verifies panel render, card selection, draft behavior, and routing.

## Notes

- Product-facing copy should use Astria. Internal repo/module naming stays StarClaw.
