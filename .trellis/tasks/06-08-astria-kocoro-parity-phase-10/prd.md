# Astria Kocoro parity phase 10: privacy and data movement foundation

## Goal

Implement the safe local foundations for the remaining Kocoro parity areas: OS credential storage, dry-run sync foundations, migration preview/apply boundaries, and upload/share privacy contracts. Keep StarClaw local-first by default and do not enable off-machine data movement without explicit opt-in work.

## Requirements

- Use `.trellis/research/kocoro-keychain-sync-migration-discovery.md` as the source plan.
- Implement Phase 10 through independently verifiable child tasks.
- Keep `sync.enabled=false` by default for any future sync work.
- Do not silently migrate existing config secrets into OS keychain.
- Do not upload sessions, artifacts, or migration data to cloud endpoints by default.
- Keep migration preview separate from migration apply.
- Preserve StarClaw naming in code/package/task artifacts.

## Acceptance Criteria

- [ ] Child tasks are planned, implemented, validated, committed, and archived independently.
- [ ] StarClaw remains local-first by default.
- [ ] No OS keychain writes, cloud sync, migration apply, or public URL delivery are enabled without explicit future opt-in controls.
- [ ] Keychain and sync foundations land before cloud uploader or migration apply work.

## Child Plan

1. `keychain-store-boundary`
   - Add StarClaw `internal/keychain` with backend interface, memory backend tests, darwin OS backend, and non-darwin unsupported behavior.
   - Do not migrate existing config secrets automatically.

2. `sync-marker-dry-run-foundation`
   - Add local-only sync config defaults, marker, lock, and dry-run outbox.

3. `session-sync-batcher-privacy`
   - Add StarClaw session candidate discovery and privacy-preserving dry-run batching.

4. `claude-code-migration-preview`
   - Add preview-only scanner/planner without apply.

5. `migration-apply-approval-boundary`
   - Add apply with TTL, freshness, conflict, and explicit approval.

6. `cloud-upload-share-privacy-contract`
   - Define and enforce provider-backed data movement warnings, redaction, and retraction contracts.

## Out of Scope

- Implementing all Phase 10 children in one task.
- Enabling real cloud upload/sync by default.
- Copying Kocoro branding or package paths into StarClaw runtime code.
