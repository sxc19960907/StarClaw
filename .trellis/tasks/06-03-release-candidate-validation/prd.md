# Release candidate validation

## Goal

Run and record a release-candidate validation pass for the current StarClaw checkout.

## Requirements

- Verify the working tree is clean before validation.
- Run core local validation: full tests, vet, CLI smoke, app launch smoke, local release install smoke.
- Run race tests for the critical packages listed in the release checklist.
- Run build validation: `make build` and `make build-all`.
- Update `RELEASE_CHECKLIST.md` with the current validation results.
- Do not tag or publish a release in this task.

## Acceptance Criteria

- [x] Validation commands complete successfully or failures are recorded clearly.
- [x] `RELEASE_CHECKLIST.md` reflects the current validation pass.
- [x] The task produces a clear release-candidate readiness conclusion.

## Notes

- Lightweight validation task; PRD-only planning is sufficient.
