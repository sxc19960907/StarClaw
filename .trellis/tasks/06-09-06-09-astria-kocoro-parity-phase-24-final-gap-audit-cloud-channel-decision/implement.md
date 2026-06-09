# Phase 24 implementation plan

## Checklist

1. Inspect current StarClaw task state and archived Phase16-23 reviews.
2. Inspect Kocoro baseline for cloud/channel, IM lifecycle, desktop/native,
   release/updater, runtime streaming, sync/share, and team surfaces.
3. Inspect StarClaw equivalents and local-only boundaries.
4. Write `final-gap-audit.md` with the parity matrix and decision.
5. Validate:
   - `python3 ./.trellis/scripts/task.py validate <task-dir>`
   - `git diff --check`
   - docs/static checks if a doc checker exists.
6. Commit the audit task and archive it if complete.

## Risky Files

- `.trellis/tasks/06-09-06-09-astria-kocoro-parity-phase-24-final-gap-audit-cloud-channel-decision/*`

## Rollback Points

- If the audit uncovers a runtime regression, stop and create a separate
  implementation task instead of mixing code fixes into this audit.
- If the Kocoro baseline is unavailable or dirty in a way that affects the
  audit, record the uncertainty in `final-gap-audit.md` and do not overstate
  parity.
