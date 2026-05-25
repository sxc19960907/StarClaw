# Build and CLI smoke verification

## Goal

Verify that the hardened codebase builds and the basic CLI surfaces still run after the recent critical/security/concurrency work.

## Requirements

- Inspect available build targets and choose the least destructive release-relevant build commands.
- Run a normal local build.
- Run cross-platform build target if available and practical.
- Run low-risk CLI smoke commands against the built binary.
- Document results, blockers, and any follow-up release validation needed.

## Acceptance Criteria

- [ ] Build command result is recorded.
- [ ] CLI smoke command results are recorded.
- [ ] Any build artifacts created by the verification are identified.
- [ ] No broad new feature work is introduced.
- [ ] Remaining release blockers, if any, are documented.

## Notes

- This is a verification task. Prefer documenting failures over expanding scope unless a small fix is clearly necessary.
