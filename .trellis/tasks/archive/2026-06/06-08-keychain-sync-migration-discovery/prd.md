# Keychain sync migration discovery

## Goal

Compare StarClaw against the local Kocoro baseline for OS keychain storage, session sync, Claude Code migration, upload/share privacy, and related credential boundaries, then produce the next implementation plan without enabling any credentialed or off-machine behavior.

## Requirements

- Use local Kocoro checkout `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245` as the parity baseline.
- Inspect Kocoro's `internal/keychain`, `internal/sync`, `internal/migrate/claudecode`, daemon migration handlers, and relevant config defaults.
- Inspect StarClaw's current config, share, upload, permissions, and secret-redaction boundaries.
- Produce a research artifact that captures:
  - confirmed Kocoro capabilities;
  - current StarClaw coverage and gaps;
  - privacy/security invariants;
  - future child tasks in recommended order;
  - acceptance criteria and validation expectations for each future task.
- Keep this task discovery-only. Do not add runtime keychain, sync, migration, or cloud upload execution paths.
- Preserve StarClaw local-first defaults: no OS keychain writes, no real cloud sync, no migration apply, no off-machine data movement, and no telemetry unless a future task explicitly adds an opt-in flow.

## Acceptance Criteria

- [ ] Research artifact exists under `.trellis/research/` and is based on the local Kocoro checkout.
- [ ] The artifact separates keychain, sync, migration, upload/share privacy, and cloud credential concerns.
- [ ] The artifact recommends independently verifiable future child tasks with acceptance criteria and test expectations.
- [ ] The artifact states non-goals and hard privacy constraints for future work.
- [ ] Trellis manifests reference the relevant backend specs and research artifact.
- [ ] No runtime code is changed as part of this discovery task.

## Out of Scope

- Implementing `internal/keychain`, `internal/sync`, or `internal/migrate`.
- Writing to macOS Keychain or any OS credential store.
- Uploading sessions or artifacts to any cloud service.
- Applying Claude Code migration changes to user files.
- Adding UI surfaces for enabling these capabilities.
