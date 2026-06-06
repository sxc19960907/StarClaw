# Memory Taxonomy

## Goal

Make Astria memory durable and governable by classifying memories into useful categories and surfacing conflicts or stale entries before they pollute project context.

## Requirements

- Add categories such as preferences, decisions, commands, architecture, people, and risks.
- Keep memory writes user-reviewed.
- Surface duplicate or conflicting memory candidates.
- Preserve compatibility with the current memory file approach.
- Make categories visible in the Memory Map UI.

## Acceptance Criteria

- [x] Memory candidates can be categorized.
- [x] Memory list can group or filter by category.
- [x] Duplicate/conflict warnings are visible before approval.
- [x] Existing memory entries continue to load.
- [x] Tests cover category parsing and conflict detection basics.

## Non-Goals

- No vector database.
- No automatic unreviewed memory writes.
- No cross-project sync.

## Dependencies

- Depends on Memory Map MVP.
