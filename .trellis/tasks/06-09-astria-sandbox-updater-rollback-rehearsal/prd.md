# Astria sandbox updater rollback rehearsal

## Goal

Add a sandbox-only failed-replacement rollback rehearsal for fake Astria.app
fixtures. The rehearsal should prove that a staged replacement failure restores
the previous fixture and leaves no partially installed candidate as the active
install, without touching real Astria paths.

## Requirements

- Operate only under a temporary sandbox directory created by the validation
  script.
- Use disposable fake `Astria.app` fixture directories only.
- Simulate a failed staged replacement after the previous install fixture has
  been snapshotted.
- Restore the previous install fixture as the active install after failure.
- Verify the failed candidate version is not left active.
- Record or validate rollback/failure state inside the sandbox.
- Reject all touched paths outside the sandbox.
- Keep real replacement disabled and avoid network, signing, notarization,
  stapling, updater private material, real app bundles, and real daemons.
- Include the rollback rehearsal in local Astria release validation.

## Acceptance Criteria

- [ ] Dedicated `--sandbox-updater-rollback-rehearsal-smoke` simulates a failed
      staged replacement and restores the previous fixture.
- [ ] The smoke verifies the active install version is the pre-update fixture
      after rollback.
- [ ] The smoke verifies the failed candidate version is not left active.
- [ ] The smoke rejects outside-sandbox rollback/touched paths.
- [ ] `--npm-only --astria-local` runs the rollback rehearsal.
- [ ] Astria release README and backend directory spec document the failed
      replacement rollback rehearsal boundary.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
