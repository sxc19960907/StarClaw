# Astria signed updater release boundary

## Goal

Harden Astria's signed updater and release metadata boundary so future updater
work has checksum/signature contracts before any app replacement behavior.

## Requirements

- Keep release validation credential-free for local development.
- Reject updater metadata unless checksum/signature verification requirements
  are present.
- Reject committed signing/notarization/updater private material.
- Do not implement app replacement without verified metadata and compatibility
  enforcement.

## Acceptance Criteria

- [ ] Release validation documents and enforces signed updater metadata
      boundary.
- [ ] Unsafe updater metadata fails validation.
- [ ] Missing updater metadata remains unavailable-safe.
