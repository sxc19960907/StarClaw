# Design

## Current State

StarClaw has `publish_to_web`, which copies a single local file into `~/.starclaw/web/<id>/<filename>` and returns a local daemon URL. The publish operation is not tracked, so the agent cannot list or retract previous shares.

Kocoro has richer share and sync packages with rendered session pages, uploader lifecycle, marker state, and retract tools. For StarClaw Phase6, the local equivalent should first make local published artifacts durable and reversible.

## Proposed Architecture

Add `internal/share`:

- `Artifact`
  - manifest record for a local published file
  - metadata only, no file contents

- `Store`
  - JSON manifest at `<starclawDir>/web/manifest.json`
  - atomic writes via temp file + rename
  - create/list/get/retract helpers

Update `internal/tools/publish_to_web.go`:

- After copying the file, write an active artifact manifest record.
- Keep current user-facing response shape and append the artifact id.

Add `internal/tools/published_files.go`:

- `list_published_files`
  - read-only
  - lists active records by default
  - optional `include_retracted`

- `retract_published_file`
  - approval required
  - validates `id`
  - removes `<starclawDir>/web/<id>`
  - marks manifest status as retracted

## Scope Boundaries

- No cloud uploader or sync scheduler.
- No session HTML rendering in this slice.
- No CDN cache semantics.
- No raw content, screenshots, cookies, or prompts in manifest.

## Compatibility

- Existing `publish_to_web` action remains registered and keeps copying files to the same URL shape.
- Existing tests that scan `~/.starclaw/web` for copied files should still pass.

## Rollback

Revert `internal/share`, `published_files.go`, and manifest write integration in `publish_to_web.go`. Existing published files remain on disk but lose manifest management.

