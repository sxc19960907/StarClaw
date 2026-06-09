# Astria staged updater transaction plan

## Goal

Add a deterministic, local-only Astria updater transaction planning boundary
that combines verified updater metadata and release compatibility inputs into a
no-replacement staged plan.

## Requirements

- Accept only metadata that keeps app replacement disabled.
- Require checksum, signature, public-key identity, app/daemon compatibility,
  rollback gate, and post-update health gate declarations before a plan is
  considered ready.
- Produce local-only, no-replacement transaction plans for smoke/release
  validation.
- Do not perform app replacement, daemon replacement, download, installation, or
  rollback.
- Do not require Apple credentials or private signing material.

## Acceptance Criteria

- [ ] Valid updater metadata plus compatibility and safety gate fields produces
      a local no-replacement transaction plan.
- [ ] Replacement-enabled metadata is rejected before any transaction plan can
      be considered ready.
- [ ] Missing rollback or post-update health gates are rejected.
- [ ] Smoke coverage is integrated into release validation.
