# Astria signed updater dry-run

## Goal

Add verified updater metadata dry-run behavior for Astria that can parse and
validate signed metadata and return a no-replacement decision.

## Requirements

- Validate metadata checksum/signature/public-key and compatibility fields using
  the existing release boundary contract.
- Produce a dry-run decision that explicitly says app replacement is disabled.
- Reject metadata that enables replacement, lacks required fields, uses
  unsupported algorithms, or contains private fields.
- Keep validation credential-free and local.

## Acceptance Criteria

- [ ] Updater dry-run handles missing, valid, invalid, and replacement-enabled
      metadata.
- [ ] Valid metadata produces a no-replacement decision with version and
      compatibility details.
- [ ] Smoke/tests cover failure and success paths.
