# Prompt suggestion dock

## Goal

Add a Home prompt suggestion dock inspired by Kocoro prompt suggestions, deriving next prompts from current workspace state.

## Requirements

- Reuse existing Web UI state only; do not add backend endpoints or LLM suggestion calls.
- Add a Home "Prompt Suggestion Dock" that proposes next prompts from runs, sessions, approvals, diagnostics, memory, MCP, inbox, file intake, and selected workflow state.
- Each suggestion should show a compact label, title, reason, and a direct action that seeds the Home mission composer.
- Include an explicit empty/default suggestion state when no recent workspace context exists.
- Refresh when existing Home state changes.
- Keep styling dense and operational with Astria's subtle celestial identity.

## Acceptance Criteria

- [x] Home renders a Prompt Suggestion Dock.
- [x] Suggestions derive from existing state without a backend call.
- [x] At least one suggestion action seeds the Home mission composer.
- [x] Empty/default state is explicit.
- [x] Core smoke verifies rendering and seed action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new suggestion API.
- No paid model call or external research.
- No replacement for workflow recipes or strategy matrix.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
- Kocoro reference: README documents prompt suggestion as a Desktop/TUI feature that can suggest the next prompt after a main turn.
- Astria implementation should be deterministic and state-derived for now.
