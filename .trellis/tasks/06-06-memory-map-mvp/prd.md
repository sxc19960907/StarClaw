# Memory Map MVP

## Goal

Create a reviewed memory workflow for Astria so useful facts from sessions can become durable context only after user approval. The product metaphor is a "Memory Map": connected stars of reusable knowledge, grounded in real source sessions.

## Requirements

- Identify memory candidates from session/run content using conservative extraction rules.
- Show source session/run references for every candidate.
- Let users approve, edit, delete, and search memories.
- Keep memory scope explicit: project, user, or global if supported.
- Make memory injection into future sessions auditable and reversible.
- Avoid silently storing sensitive content.

## Acceptance Criteria

- [x] A candidate memory can be generated from a completed session/run.
- [x] User can approve/edit/delete a candidate before it becomes active memory.
- [x] Active memories show source metadata.
- [x] Future prompt/session context can include approved memory through a documented path.
- [x] Tests cover candidate lifecycle and deletion.

## Non-Goals

- No automatic unreviewed long-term memory.
- No vector database requirement unless justified in design.
- No cross-device sync in MVP.
- No complex graph visualization in MVP; list/map hybrid is enough.

## Dependencies

- Requires clear session storage/indexing contracts.
- Should be planned with `design.md` before implementation due to data retention and privacy implications.
