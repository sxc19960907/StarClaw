# Astria Kocoro parity phase 22: production updater decision gate

## Goal

Convert the remaining Astria/Kocoro updater gap into an explicit production
decision gate. Phase21 proved sandbox updater rehearsal safety; Phase22 should
decide and enforce whether Astria may ever perform real installed app
replacement, or whether release/update execution remains CLI/npm-only while
cloud/channel parity resumes.

## Requirements

- Keep all default local validation credential-free.
- Do not implement real app replacement in this phase.
- Do not touch `/Applications`, the installed Astria app, real bundled daemons,
  user Application Support, network downloads, Apple credentials, notarization,
  stapling, or signing identities.
- Add a machine-readable production updater decision artifact that states the
  current updater strategy and the blockers for real replacement.
- Validate that real replacement remains disabled unless the decision artifact
  explicitly records a future approved strategy with required production gates.
- Compare against the Kocoro baseline:
  `/Users/timmy/PycharmProjects/Kocoro` at
  `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Acceptance Criteria

- [ ] A child task adds a production updater decision manifest and smoke gate.
- [ ] The manifest validates the current strategy as replacement disabled.
- [ ] The smoke rejects enabling real app replacement without required gates.
- [ ] `--npm-only --astria-local` includes the decision gate.
- [ ] Final review decides whether the next phase is real updater design,
      CLI/npm update hardening, or Kocoro cloud/channel parity.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
