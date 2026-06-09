# Astria production updater decision manifest

## Goal

Add a credential-free production updater decision manifest and smoke gate for
Astria. The manifest should explicitly state that real installed app
replacement remains disabled today and list the gates required before any
future production updater executor can be enabled.

## Requirements

- Add validation support for an Astria production updater decision manifest.
- The valid current decision must keep:
  - `strategy=cli_npm_only`
  - `app_replacement=disabled`
  - `local_validation_credential_free=true`
- The manifest must declare required future gates for any real app replacement:
  signed/notarized/stapled app artifact, signed updater metadata, checksum
  verification, compatibility manifest, transaction plan, rollback/health gate,
  sandbox rehearsal, operator approval, and post-replacement relaunch plan.
- The smoke must reject manifests that enable replacement without all required
  future gates.
- The smoke must reject private material references.
- The smoke must run without Apple credentials, network, downloads, notarization,
  signing identities, real app bundles, or real daemon replacement.
- `--npm-only --astria-local` must run the decision smoke.

## Acceptance Criteria

- [ ] Dedicated `--astria-production-updater-decision-smoke` passes for the
      current disabled/CLI-npm-only strategy.
- [ ] The smoke rejects replacement-enabled manifests missing required gates.
- [ ] The smoke rejects private signing/updater material references.
- [ ] `--npm-only --astria-local` runs the decision smoke.
- [ ] Astria README and backend spec document the production updater decision
      boundary.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
