# Astria release acceptance gates

## Goal

Strengthen Astria production release acceptance validation so release metadata
must explicitly declare signing, notarization, stapling, updater compatibility,
rollback/health gates, and credential-free local validation posture before a
future production release can be considered acceptable.

## Requirements

- Define a local JSON acceptance manifest shape for Astria production release
  readiness.
- Require Developer ID signing, Hardened Runtime, notarization, stapling,
  checksum/signature metadata, compatibility manifest, rollback/health gate
  manifest, and updater transaction plan references.
- Keep `replacement` disabled and `local_validation_credential_free=true`.
- Reject private credential references or committed private material fields.
- Do not require Apple credentials in local validation.
- Do not publish, notarize, staple, upload, or replace the app.

## Acceptance Criteria

- [ ] A valid acceptance manifest passes validation.
- [ ] Missing signing/notarization/stapling declarations fail validation.
- [ ] Missing updater transaction or rollback/health gate references fail
      validation.
- [ ] Private credential references fail validation.
- [ ] Release validation includes a dedicated acceptance-gates smoke.
