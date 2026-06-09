# Astria release compatibility manifest

## Goal

Generate and validate a release compatibility manifest for Astria app plus
bundled daemon versions.

## Requirements

- Capture app version/build and bundled daemon version inputs.
- Keep release manifest generation credential-free.
- Reject mismatched or missing version fields in release-candidate validation.

## Acceptance Criteria

- [ ] Manifest shape is documented and validated.
- [ ] Smoke/tests cover matching and mismatched app/daemon versions.
