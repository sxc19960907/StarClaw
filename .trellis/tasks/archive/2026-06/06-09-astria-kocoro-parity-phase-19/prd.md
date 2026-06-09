# Astria Kocoro parity phase 19: verified updater release execution

## Goal

Close the next Kocoro parity gap after Phase18 by moving from updater/release
metadata boundaries toward verified updater dry-run execution, release
compatibility manifests, and user-approved local OS crash artifact collection.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase18 completed local crash summaries, notification readiness, and signed
  updater metadata boundary validation.
- Phase18 final gap review estimates Astria is roughly 94-96% aligned with
  Kocoro for local-first desktop platform behavior.
- Remaining gaps are verified updater execution, release compatibility
  manifests, and richer local crash artifact collection.

## Child Plan

1. `astria-signed-updater-dry-run`: verify signed updater metadata and produce
   a no-replacement dry-run decision.
2. `astria-release-compatibility-manifest`: generate/check release
   compatibility metadata for Astria app plus bundled daemon.
3. `astria-local-os-crash-artifact-collection`: collect local OS crash artifacts
   only through explicit user-approved/export flows.

## Requirements

- Keep updater behavior no-replacement until verification, compatibility, and
  transaction safety are implemented.
- Keep release validation credential-free for local development.
- Do not commit signing/notarization/updater private material.
- Do not upload crash artifacts automatically.
- Preserve existing daemon/browser/Astria fallback paths.

## Acceptance Criteria

- [ ] Each child task has independent planning artifacts and testable
      acceptance criteria before implementation.
- [ ] Updater metadata verification supports a no-replacement dry-run decision.
- [ ] Release compatibility manifests cover app and bundled daemon versions.
- [ ] Local OS crash artifact collection is user-triggered and redacted.
- [ ] Final gap review updates Kocoro parity and remaining production release
      gaps.

## Out of Scope

- Real app replacement.
- Public signed/notarized release publication.
- Remote crash ingestion.
