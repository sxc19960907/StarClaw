# Astria Kocoro parity phase 9: native desktop tools and local integration depth

## Goal

Close the highest-value local native integration gaps left after Phase 8, using local Kocoro commit `74cdb3c` as the parity baseline while preserving StarClaw's local-first defaults.

## Requirements

- Use `.trellis/research/kocoro-native-tool-parity-phase9-plan.md` as the source plan.
- Implement Phase 9 through independently verifiable child tasks.
- Keep real cloud credentials, external sync, off-machine delivery, and OS credential writes disabled unless explicitly approved.
- Keep StarClaw naming in code/package/task artifacts.
- Start with Desktop RPC calendar protocol depth before calendar tools.

## Child Plan

1. `desktop-rpc-calendar-protocol`
   - Expand Desktop RPC from system-only to calendar v1 protocol constants, payload types, error codes, and helper behavior.
   - Keep transport local over Unix sockets.

2. `calendar-native-tool-boundary`
   - Add Desktop-RPC-backed calendar tools and registration boundaries.
   - Do not access EventKit directly from the daemon.

3. `browser-handoff-lease-depth`
   - Add per-run browser lease tracking and reload handoff cleanup semantics.

4. `terminal-workspace-tool-depth`
   - Add a local terminal workspace helper, Ghostty-compatible where available.

5. `image-tool-provider-boundary`
   - Add provider-gated image generation/editing tools with approval and disabled-by-default credentials.

6. `keychain-sync-migration-discovery`
   - Plan keychain, sync, migration, upload/share, and privacy boundaries before implementation.

## Acceptance Criteria

- [ ] Child tasks are planned, implemented, validated, committed, and archived independently.
- [ ] Phase 9 remains local-first by default.
- [ ] No real cloud credentials, external sync, off-machine delivery, or OS keychain writes are enabled without explicit approval.
- [ ] StarClaw closes the calendar/Desktop RPC foundation gap before adding higher-risk provider/sync work.

## Out of Scope

- Implementing all Phase 9 children in a single task.
- Copying Kocoro branding or package paths into StarClaw runtime code.
- Enabling Shannon/Kocoro cloud credentials by default.
