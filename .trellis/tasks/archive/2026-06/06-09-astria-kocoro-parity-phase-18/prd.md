# Astria Kocoro parity phase 18: production native reliability

## Goal

Close the next Kocoro parity gap after Phase17 by improving Astria production
native reliability and release readiness: local crash summaries, notification
readiness, and signed updater/release safety boundaries.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase17 completed native clipboard/file affordances, permission helper
  guidance, and per-window safe route restoration.
- Phase17 final gap review estimates Astria is roughly 92-95% aligned with
  Kocoro for local-first desktop platform behavior.
- Remaining gaps are production-native polish and release execution depth.

## Child Plan

1. `astria-local-crash-reporter-summaries`: add local crash summary discovery
   and redacted export boundaries for Astria support workflows.
2. `astria-native-notification-readiness`: add notification permission/status
   readiness and smoke-tested local notification guidance.
3. `astria-signed-updater-release-boundary`: harden signed updater and release
   metadata boundaries with unavailable-safe checksum/signature contracts.

## Requirements

- Keep all crash/notification/release affordances local-first and
  user-triggered.
- Do not add automatic crash upload, off-machine telemetry, cloud auth, or
  committed Apple credentials.
- Do not expose secrets, raw prompts, Desktop RPC payloads, socket paths,
  pidfile paths, crash file paths, signing identities, keychain profiles, or
  updater private material.
- Preserve existing daemon, browser, and Astria fallback paths.
- Add smoke/test coverage for each production-native contract.

## Acceptance Criteria

- [ ] Each child task has independent planning artifacts and testable
      acceptance criteria before implementation.
- [ ] Astria exposes local crash summary/report affordances without automatic
      upload.
- [ ] Astria defines notification readiness and permission boundaries without
      surprise prompts.
- [ ] Release/updater boundaries reject unsafe metadata and private material.
- [ ] Final gap review updates Kocoro parity and remaining production-native
      gaps.

## Out of Scope

- Actual remote crash ingestion.
- Real public signed/notarized release publication.
- Replacing the daemon-served Web UI with native UI.
