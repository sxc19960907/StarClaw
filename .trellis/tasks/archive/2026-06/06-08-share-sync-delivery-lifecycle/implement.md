# Implementation Plan

## Scope

Add local share artifact manifest tracking plus list/retract tools.

## Steps

1. Read backend/tool quality guidelines.
2. Add `internal/share` store:
   - artifact types
   - load missing/corrupt manifest safely
   - atomic write
   - create/list/retract helpers
3. Update `publish_to_web`:
   - write manifest record
   - include artifact id in result
4. Add `list_published_files` and `retract_published_file` tools.
5. Register new tools.
6. Add unit tests:
   - store create/list/retract
   - publish writes manifest
   - list defaults to active
   - list can include retracted
   - retract missing/retracted cases
7. Run:
   - `go test ./internal/share`
   - `go test ./internal/tools`
   - `go test ./...`
   - `git diff --check`
8. Commit and archive child task.

## Review Gates

- Manifest contains metadata only; no file content.
- Retraction removes only the artifact directory for the given id.
- Manifest writes are atomic.
- Existing `publish_to_web` response compatibility is preserved.

## Completion Notes

- Added `internal/share` manifest store for local published artifact metadata.
- Updated `publish_to_web` to write active manifest records and report artifact ids.
- Added `list_published_files` and `retract_published_file` local tools.
- Registered the new tools.
- Added store and tool tests for create/list/retract, active filtering, and idempotent retract.

## Validation

- `go test ./internal/share ./internal/tools` — passed.
- `go test ./...` — passed.
- `git diff --check` — passed.
- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-share-sync-delivery-lifecycle` — passed.
