# Knowledge curation console

## Goal

Add an Astria knowledge curation console that turns runs, sessions, memory facts, and warnings into reviewable long-term context actions.

## Requirements

- Reuse existing Web UI state only: sessions, runs, memory entries/facts/warnings/categories, and recent run/session metadata.
- Add a Home "Knowledge Curation" console that makes long-term memory work visible and actionable.
- Surface memory warnings, classified facts, source entries, favorite/recent sessions, recent completed runs, and missing/low-context states.
- Each item should include a concise status, review label, short context, and direct navigation action to Memory, Chat, or Runs when appropriate.
- Include an explicit empty/low-context state that guides the user toward Memory Capture or recent work review.
- Refresh when memory, sessions, or runs update.
- Keep styling dense and operational with subtle Astria celestial identity; no new frontend dependencies or backend endpoints.

## Acceptance Criteria

- [x] Home renders a Knowledge Curation console.
- [x] Console derives items from at least four existing memory/session/run state sources.
- [x] Items navigate to Memory, Chat/session, or Runs where available.
- [x] Empty/low-context state is explicit.
- [x] Core smoke verifies rendering and one navigation action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new memory write backend.
- No automatic durable memory writes.
- No replacement for the Memory Map panel.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
