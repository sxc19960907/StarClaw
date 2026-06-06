# Memory Taxonomy Design

## Storage Compatibility

Keep the existing `MEMORY.md` file format. Taxonomy is inferred from markdown text instead of introducing a database.

## Categories

Recognized categories:

- preferences
- decisions
- commands
- architecture
- people
- risks
- uncategorized

Accepted line forms:

- `- [decision] Use embedded Web UI assets.`
- `- decision: Use embedded Web UI assets.`
- `## Decisions` section headings apply to following bullets.

## API Additions

`GET /memory` adds:

- `categories`: counts per category.
- `facts`: parsed memory facts with category, text, source entry, and line number.
- `warnings`: duplicate/conflict signals.

No existing fields are removed.

## Conflict / Duplicate MVP

- Duplicate: normalized fact text appears more than once.
- Conflict: same normalized subject appears with different fact text. Subject is the first stable phrase before `:` or ` is ` / ` should ` / ` uses `.

This is intentionally conservative and transparent; it is a warning system, not automatic rewriting.

## UI

Memory Map adds:

- Category filter.
- Taxonomy overview counts.
- Warnings panel.
- Candidate preview classification based on current textarea content.
