# Astria updater rollback health gates

## Goal

Define and validate credential-free Astria rollback and post-update health gate
manifests so future updater replacement cannot be considered ready without
explicit safety gates.

## Requirements

- Define a local JSON manifest shape for rollback and post-update health gates.
- Require rollback gate identity, rollback source, restore target, daemon
  compatibility guard, and manual approval state.
- Require post-update health checks for app launch, daemon health, Desktop RPC
  capability compatibility, and Web UI readiness.
- Keep replacement disabled and local-only.
- Do not perform rollback, installation, download, or app replacement.
- Do not require Apple credentials or private signing material.

## Acceptance Criteria

- [ ] A valid rollback/health gate manifest passes validation.
- [ ] Missing rollback source/target or disabled manual approval fails
      validation.
- [ ] Missing post-update health checks fail validation.
- [ ] Release validation includes a dedicated smoke for rollback/health gates.
