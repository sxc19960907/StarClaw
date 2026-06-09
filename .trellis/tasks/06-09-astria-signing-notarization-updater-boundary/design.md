# Astria signing notarization updater boundary design

## Current Shape

`scripts/build_macos_astria_shell.sh` builds an unsigned local `Astria.app`.
`scripts/validate_release_artifacts.sh --npm-only --astria-local` runs npm
package dry-run checks and the Astria shell smoke. Docs state that unsigned
development builds are not notarized release artifacts.

## Proposed Shape

Keep local development credential-free, but make release boundaries explicit
and machine-checkable:

- Add an Astria distribution boundary check to
  `scripts/validate_release_artifacts.sh --astria-local`.
- Verify local `Astria.app` bundle structure and bundled daemon smoke through
  the existing smoke script.
- Verify that credential files or private updater material are not present in
  the repository.
- Verify that updater metadata is unavailable-safe unless a future task adds
  signed metadata validation.
- Document signing/notarization prerequisites and updater safety checks in
  `docs/INSTALL.md`.

## Test Strategy

Run:

- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./...`

The validation must pass without Apple credentials.
