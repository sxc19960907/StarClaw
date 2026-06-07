# Approval center control console

## Goal

Add an Astria Approval Center that centralizes human-in-the-loop approvals, permission risk, diagnostics, inbox, and failed run recovery.

## Requirements

- Reuse existing Web UI state only: approvals, permissions, diagnostics, inbox, runs, and config/MCP state.
- Add an "Approval Center" / human-in-the-loop console that is reachable from Home and fits the Astria operational workspace.
- Surface pending approval count, permission policy risk, diagnostics readiness, failed/attention runs, failed/pending inbox, and MCP/tooling gaps.
- Each item should include a concise status, risk label, short context, and direct navigation action to the relevant existing panel.
- The console should include a clear empty/low-risk state.
- Refresh whenever underlying state sources update.
- Keep styling dense, work-focused, and aligned with Kocoro/Shannon human approval patterns; do not create a marketing card.

## Acceptance Criteria

- [x] Home renders an Approval Center / human-in-the-loop console.
- [x] Console derives items from at least five existing state sources.
- [x] Items navigate to relevant existing panels or run/session views.
- [x] Empty/low-risk state is explicit.
- [x] Core smoke verifies rendering and one navigation action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new approval backend.
- No persistent dismiss/snooze workflow.
- No replacement for Permissions, Diagnostics, Inbox, Runs, or MCP panels.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
