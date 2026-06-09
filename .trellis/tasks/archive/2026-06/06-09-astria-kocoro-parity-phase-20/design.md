# Astria Kocoro parity phase 20 design

## Architecture Boundary

Phase20 introduces production updater transaction safety artifacts while keeping
runtime replacement disabled. A transaction plan may explain what would happen
after metadata, compatibility, rollback, and health gates pass, but it must not
replace the app or bundled daemon.

## Native/Release Areas

- Staged transaction plan: combine verified updater metadata and compatibility
  manifest inputs into a deterministic no-replacement local plan.
- Rollback/health gates: declare the rollback inputs and post-update health
  checks required before future replacement can be enabled.
- Release acceptance gates: require production release metadata to advertise
  signing/notarization/compatibility posture without local private credentials.

## Compatibility

- Existing CLI self-update remains separate from Astria app update behavior.
- Missing updater metadata remains unavailable-safe.
- Local release validation remains credential-free.

## Rollout

1. Staged updater transaction plan.
2. Rollback and health gates.
3. Release acceptance gates.
4. Final Kocoro gap review and next-scope decision.
