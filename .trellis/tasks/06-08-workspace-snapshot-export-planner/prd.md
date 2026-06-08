# Workspace snapshot export planner

## Goal

Add an Astria Workspace Snapshot planning surface to the embedded daemon Web UI so a user can review and draft local handoff/resume snapshot packs from the current workspace context.

## Requirements

- Expose the snapshot planner from the sidebar and Manage hub.
- Summarize reusable local context across sessions, runs, memory, sources, result archives, playbooks, share packs, agents, schedules, and risks.
- Provide multiple snapshot pack templates for resume, evidence, source/memory, results, playbooks, delivery, and privacy review.
- Selecting a snapshot pack must show included context, missing pieces, review gate, privacy/redaction boundary, and next route.
- Each snapshot pack must be able to draft a prompt into Chat and route the user to a relevant existing panel.
- Keep this as a static embedded daemon Web UI feature. Do not add backend storage, export endpoints, external assets, or a frontend build pipeline.

## Acceptance Criteria

- [x] Sidebar and Manage hub expose Workspace Snapshot with live counts.
- [x] Workspace Snapshot panel renders snapshot pack cards and a selected detail brief.
- [x] Snapshot detail includes included context, missing pieces, review gate, privacy/redaction boundary, and route action.
- [x] Draft actions populate Chat with a snapshot/export planning prompt.
- [x] Route actions open relevant source panels.
- [x] Web UI smoke coverage verifies panel render, detail selection, draft behavior, and routing.

## Notes

- Product-facing copy should use Astria. Internal repo/module naming stays StarClaw.
