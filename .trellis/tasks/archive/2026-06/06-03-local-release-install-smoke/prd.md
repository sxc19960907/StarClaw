# Local release install smoke

## Goal

Make release install smoke runnable from a normal development checkout without requiring a pre-existing GoReleaser archive.

## Requirements

- Add a local smoke script that builds a current-platform release-style archive.
- The local archive must contain the `starclaw` binary at a path accepted by `scripts/smoke_release_install.sh`.
- The script must reuse `scripts/smoke_release_install.sh` for the actual extraction/install/runtime checks.
- The script must work on Linux, macOS, and Windows-like environments where bash can run.
- Avoid writing artifacts into tracked paths by default.
- Document the new local smoke entry point in release docs.

## Acceptance Criteria

- [x] A developer can run one script to build a local archive and execute release install smoke.
- [x] The script chooses `.tar.gz` for Unix platforms and `.zip` for Windows.
- [x] The script injects a non-`dev` version into the binary so release-oriented version paths are exercised.
- [x] The script runs successfully on the current platform.
- [x] Release checklist documents the new local smoke command.

## Notes

- Existing `scripts/smoke_release_install.sh` already validates extraction, app startup, runtime routes, and doctor JSON.
