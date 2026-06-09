# Astria Kocoro parity phase 20: production updater transaction safety

## Goal

Close the next Kocoro parity gap after Phase19 by moving Astria from verified
updater metadata and dry-run decisions toward production-safe updater
transaction planning, rollback/health gates, and release acceptance boundaries.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase19 completed signed updater dry-run validation, release compatibility
  manifests, and user-approved local crash artifact collection.
- Phase19 final gap review estimates Astria is roughly 95-97% aligned with
  Kocoro for local-first desktop platform behavior.
- Remaining local-first gaps are production updater transaction safety,
  rollback/health gates, and signed/notarized release acceptance. Broader
  cloud/channel parity remains a later scope decision.

## Child Plan

1. `astria-staged-updater-transaction-plan`: produce a local, no-replacement
   transaction plan from verified metadata plus compatibility manifest inputs.
2. `astria-updater-rollback-health-gates`: define and validate rollback and
   post-update health gate manifests without replacing the app.
3. `astria-release-acceptance-gates`: strengthen release validation so signed
   or notarized acceptance requirements are explicit while local validation
   stays credential-free.

## Requirements

- Keep all updater behavior no-replacement until a future task explicitly
  implements transactional replacement.
- Do not require Apple credentials for local validation.
- Do not commit signing identities, notary profiles, private keys, provisioning
  profiles, updater private material, or notarization secrets.
- Preserve existing CLI/browser/Astria fallback paths.
- Keep transaction plans local-only, deterministic, and smoke-testable.

## Acceptance Criteria

- [ ] Each child task has independent planning artifacts and testable
      acceptance criteria before implementation.
- [ ] Staged updater transaction planning validates metadata, compatibility,
      replacement disabled state, and required safety gates.
- [ ] Rollback/health gate manifests are defined and validated.
- [ ] Release acceptance gates reject missing or unsafe production release
      metadata without requiring private credentials.
- [ ] Final gap review updates Kocoro parity and decides whether the next phase
      is cloud/channel parity or declaring local-first parity effectively
      complete.

## Out of Scope

- Real app or bundled-daemon replacement.
- Shipping public signed/notarized releases.
- Remote update distribution.
- Cloud IM or team sync implementation.
