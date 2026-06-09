# Astria Kocoro parity phase 21 design

## Architecture Boundary

Phase21 is a sandbox rehearsal layer. It may copy and mutate disposable fixture
directories under `mktemp`, but it must not mutate installed app paths or real
runtime support directories.

## Rehearsal Model

- Build a fake current `Astria.app` fixture.
- Build a fake candidate `Astria.app` fixture.
- Stage the candidate into a temporary staging directory.
- Rehearse replacement into a sandbox install directory.
- Rehearse rollback to the original fixture.
- Emit a local JSON result with touched paths and guard decisions.

## Compatibility

- Reuses Phase20 validation script boundaries.
- Keeps production replacement disabled.
- Keeps local validation credential-free.

## Rollout

1. Fixture replacement/rollback rehearsal.
2. Fixture health-gate rehearsal.
3. Failed replacement rollback rehearsal.
4. Final Kocoro gap review.
