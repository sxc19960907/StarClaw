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

- [x] A child task adds a production updater decision manifest and smoke gate.
- [x] The manifest validates the current strategy as replacement disabled.
- [x] The smoke rejects enabling real app replacement without required gates.
- [x] `--npm-only --astria-local` includes the decision gate.
- [x] Final review decides whether the next phase is real updater design,
      CLI/npm update hardening, or Kocoro cloud/channel parity.

## Final Review

Phase22 added `astria-production-updater-decision-manifest` as a release
validation gate. The valid current decision is:

- `strategy=cli_npm_only`
- `app_replacement=disabled`
- `local_validation_credential_free=true`

The smoke rejects replacement-enabled decisions, missing future production
gates, and private signing/updater material references. It is included in
`scripts/validate_release_artifacts.sh --npm-only --astria-local`, so real app
replacement cannot quietly slip into local Astria release validation.

Kocoro has a real CLI self-update path through GitHub releases. Astria now has
explicit policy and rehearsal coverage for the desktop app updater boundary,
but still intentionally does not perform real installed app replacement.

Decision: keep Astria on CLI/npm-only executable updates for now. Do not start a
real signed/notarized app updater executor unless the project explicitly scopes
Apple signing, notarization, privileged install edge cases, artifact hosting,
and relaunch UX. The next Kocoro parity phase should pivot back to larger
cloud/channel parity gaps rather than spending more time on app replacement
mechanics.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
