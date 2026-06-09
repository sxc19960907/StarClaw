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

- [x] Each child task has independent planning artifacts and testable
      acceptance criteria before implementation.
- [x] Sandbox fixture rehearsal stages and rolls back disposable app-bundle
      fixtures without touching real Astria paths.
- [x] Simulated health gates run against the fixture.
- [x] Failed staged replacement rehearsal proves rollback.
- [x] Final gap review updates local-first Kocoro parity and decides whether to
      continue updater execution work or pivot to cloud/channel parity.

## Final Gap Review

Phase21 completed the remaining sandbox-only updater rehearsal layer:

- `astria-sandbox-updater-rehearsal-fixture` added disposable fake
  `Astria.app` staging, sandbox replacement, rollback, and outside-sandbox path
  rejection.
- `astria-sandbox-updater-health-rehearsal` added post-update fixture markers
  for app launch, daemon health, Desktop RPC capabilities, and Web UI
  readiness, including missing-marker and outside-sandbox negative cases.
- `astria-sandbox-updater-rollback-rehearsal` added failed staged replacement
  rehearsal that restores the previous fixture, records rolled-back sandbox
  state, and verifies the failed candidate is not left active.

Kocoro baseline at `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245` has a real CLI
self-update path through `cmd/update.go` and `internal/update/selfupdate.go`,
plus npm/GitHub release install/update assets. Astria intentionally does not
perform real app replacement yet. Compared with Kocoro, Astria now has stronger
local desktop updater safety rehearsal than before, but still lacks a
production updater executor for installed app bundle replacement.

Updated parity estimate: Astria is roughly 97-98% aligned with Kocoro for
local-first desktop platform behavior. The remaining local-first delta is the
actual signed/notarized updater execution path, including real installed app
replacement, real artifact download, signature verification over production
metadata, privileged filesystem edge cases, and post-replacement relaunch.

Recommended next direction: do not jump straight to real replacement unless we
want to scope Apple signing/notarization and installer risk. The better next
phase is a narrow production-updater design gate that decides whether Astria
should implement a real Sparkle-style app updater, keep npm/CLI updates as the
only executable update path, or pivot back to Kocoro's larger cloud/channel
parity gaps.

## Out of Scope

- Real installed app replacement.
- Real bundled daemon replacement.
- Public signed/notarized release publication.
- Network update downloads.
- Cloud/channel parity.
