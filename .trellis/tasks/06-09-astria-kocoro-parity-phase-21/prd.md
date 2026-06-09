# Astria Kocoro parity phase 21: sandbox updater rehearsal

## Goal

Close the next local-first updater gap after Phase20 by adding a sandbox-only
updater rehearsal that exercises staged replacement and rollback against a
disposable fixture, without touching the installed Astria app or bundled daemon.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase20 completed local metadata, transaction, rollback/health, and release
  acceptance gates while keeping replacement disabled.
- Phase20 final gap review estimates Astria is roughly 96-98% aligned with
  Kocoro for local-first desktop platform behavior.
- Remaining local-first gap is not safety metadata; it is sandboxed end-to-end
  updater rehearsal before any real replacement implementation exists.

## Child Plan

1. `astria-sandbox-updater-rehearsal-fixture`: add a disposable fixture
   rehearsal that stages a candidate app bundle, verifies expected files, rolls
   back the fixture, and proves real Astria paths are never touched.
2. `astria-sandbox-updater-health-rehearsal`: add simulated post-update health
   gate checks against the fixture.
3. `astria-sandbox-updater-rollback-rehearsal`: deepen rollback checks for
   failed staged replacements.

## Requirements

- Operate only under a temporary sandbox directory created by the validation
  script.
- Never mutate `/Applications`, the real Astria app bundle, user Application
  Support, or the real bundled daemon.
- Keep replacement disabled for real releases.
- Do not require Apple credentials, network downloads, notarization, stapling,
  or signing identities.
- Keep rehearsal deterministic and smoke-testable.

## Acceptance Criteria

- [ ] Each child task has independent planning artifacts and testable
      acceptance criteria before implementation.
- [ ] Sandbox fixture rehearsal stages and rolls back disposable app-bundle
      fixtures without touching real Astria paths.
- [ ] Simulated health gates run against the fixture.
- [ ] Failed staged replacement rehearsal proves rollback.
- [ ] Final gap review updates local-first Kocoro parity and decides whether to
      continue updater execution work or pivot to cloud/channel parity.

## Out of Scope

- Real installed app replacement.
- Real bundled daemon replacement.
- Public signed/notarized release publication.
- Network update downloads.
- Cloud/channel parity.
