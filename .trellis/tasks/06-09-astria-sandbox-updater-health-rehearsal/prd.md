# Astria sandbox updater health rehearsal

## Goal

Add a sandbox-only post-update health rehearsal for fake Astria.app fixtures.
The rehearsal should simulate the health gates a future updater must pass after
staging a candidate fixture, without touching the real app, daemon, user
Application Support, network, or Apple release credentials.

## Requirements

- Operate only under a temporary sandbox directory created by the validation
  script.
- Reuse disposable fake `Astria.app` fixture directories rather than a real app
  bundle.
- Simulate post-update health gates for:
  - app launch marker
  - daemon health marker
  - Desktop RPC capabilities marker
  - Web UI readiness marker
- Require all health marker paths and fixture paths to stay under the sandbox.
- Fail deterministically when a required post-update health marker is absent.
- Keep real replacement disabled and avoid network, signing, notarization,
  stapling, or updater private material.
- Include the health rehearsal in local Astria release validation.

## Acceptance Criteria

- [ ] Dedicated `--sandbox-updater-health-rehearsal-smoke` passes with a fully
      healthy fake candidate fixture.
- [ ] The smoke fails a negative case with a missing required health marker.
- [ ] The smoke rejects any health marker path outside the sandbox.
- [ ] `--npm-only --astria-local` runs the health rehearsal.
- [ ] Astria release README and backend directory spec document the health
      rehearsal boundary.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
