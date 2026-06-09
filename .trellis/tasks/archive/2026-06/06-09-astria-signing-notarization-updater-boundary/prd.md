# Astria signing notarization updater boundary

## Goal

Harden the Astria release/distribution boundary for signed/notarized macOS app
builds and future updater metadata without requiring private credentials in
local development.

## Requirements

- Document signing, Hardened Runtime, notarization, stapling, and updater
  metadata expectations.
- Keep unsigned local builds and smoke tests usable.
- Do not commit signing identities, keychain profiles, Apple credentials,
  updater private keys, or notarization secrets.
- Validate release artifacts where credentials are absent by checking structure
  and explicit unavailable-safe behavior.

## Acceptance Criteria

- [ ] Release validation covers Astria local artifacts and credential-free
      boundaries.
- [ ] Docs describe signed/notarized release prerequisites and updater metadata
      safety checks.
- [ ] Local validation does not require private Apple credentials.

## Notes

- Actually producing a signed public release is out of scope for this task.
