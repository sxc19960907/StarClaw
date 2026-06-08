# Share sync delivery lifecycle

## Goal

Implement the final Phase6 Kocoro parity slice by adding a local share artifact lifecycle to StarClaw: published files should be tracked, listable, and retractable. This closes the first local-first gap toward Kocoro's `internal/share`, `internal/sync`, and retractable delivery behavior without introducing cloud sync.

## Requirements

- Use local Kocoro evidence:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/share/*`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/sync/*`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/retract_published_file.go`
- Keep this child local-first:
  - no cloud uploader
  - no remote sync endpoint
  - no account/user identity
  - no external telemetry
- Add local share artifact tracking:
  - manifest stored under StarClaw data dir
  - records include id, filename, source path, local path, URL, size, purpose, status, created/retracted timestamps
  - manifest writes are atomic
- Update `publish_to_web` to create manifest records.
- Add local list and retract tools:
  - `list_published_files`: read-only list of active/retracted artifacts.
  - `retract_published_file`: removes local published directory and marks manifest record retracted.
- Register the new tools.
- Keep existing `publish_to_web` behavior compatible.

## Acceptance Criteria

- [ ] Publishing a file still copies it under `~/.starclaw/web/<id>/...` and returns the existing style URL.
- [ ] Publishing writes a manifest record with active status.
- [ ] Listing returns active artifacts by default and can include retracted entries.
- [ ] Retracting by id removes the local artifact directory and marks the manifest record retracted.
- [ ] Re-retracting a retracted artifact is idempotent and reports already retracted.
- [ ] Tools avoid leaking file contents; manifest stores paths and metadata only.
- [ ] Unit tests cover publish manifest, list, retract, missing id, and idempotent retract.
- [ ] Full project tests pass.

