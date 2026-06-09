# Phase19 final gap review

## Completed scope

Phase19 closed the next verified updater release-execution slice without
turning Astria into a replacement-capable updater.

- `astria-signed-updater-dry-run` added signed updater metadata validation and a
  dry-run decision that keeps app replacement disabled.
- `astria-release-compatibility-manifest` added credential-free compatibility
  manifest generation/checking for Astria app and bundled daemon versions.
- `astria-local-os-crash-artifact-collection` added a user-triggered local crash
  artifact export boundary with redacted snippets, local-only output, explicit
  approval, and no upload path.

## Validation evidence

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-local-os-crash-artifact-collection`
- `scripts/smoke_macos_astria_shell.sh`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

## Updated Kocoro parity estimate

Astria is now roughly 95-97% aligned with Kocoro for local-first desktop
platform behavior.

The core local platform boundaries are present: standalone macOS shell, daemon
supervision, Desktop RPC handshake/session diagnostics, native menu/window
behavior, local diagnostics and crash exports, notification/permission readiness
guidance, signed updater metadata validation, dry-run updater decisions,
release compatibility manifests, and user-approved local crash artifact
collection.

## Remaining Kocoro gaps

- Real signed/notarized production release execution: Astria still validates
  metadata and manifests locally, but does not yet execute a transactional app
  or bundled-daemon replacement.
- Updater rollback and install safety: no rollback plan, staged replacement, or
  post-update health gate exists yet.
- Public release pipeline: local validation is credential-free, but the project
  still lacks the full external signing/notarization/stapling/public release
  path.
- Deep cloud/channel parity: the local-first desktop platform is close, but
  Kocoro's broader cloud IM, remote delivery, and team/distribution surfaces are
  still outside the local Astria shell scope.

## Recommended next phase

Phase20 should focus on production update transaction safety:

1. Signed/notarized release artifact acceptance gates that remain
   credential-safe in local validation.
2. Staged updater transaction planning with no replacement by default.
3. Rollback/health-check manifest design for future app and bundled-daemon
   replacement.
4. Final decision on whether to pursue cloud/channel parity next or declare the
   Kocoro local-first platform target effectively complete.
