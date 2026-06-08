# Keychain sync migration discovery design

## Scope

This task is a planning and evidence task for the high-risk Kocoro parity areas that remain after Phase 9 native tool work:

- OS credential storage.
- Session sync marker, scanning, batching, and uploader lifecycle.
- Claude Code migration preview/apply architecture.
- Upload/share privacy contracts.
- Local-first defaults and explicit opt-in controls.

The deliverable is a research artifact and next-phase plan. It must not add execution paths that can write secrets, move data off-machine, or mutate user configuration.

## Evidence Sources

Kocoro baseline:

- `/Users/timmy/PycharmProjects/Kocoro/internal/keychain/*`
- `/Users/timmy/PycharmProjects/Kocoro/internal/sync/*`
- `/Users/timmy/PycharmProjects/Kocoro/internal/migrate/claudecode/*`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/migrate_claudecode.go`

StarClaw baseline:

- `internal/config/config.go`
- `internal/config/multilevel.go`
- `internal/share/store.go`
- `internal/uploads/uploads.go`
- `internal/permissions/permissions.go`
- daemon secret-redaction tests around config, run events, replay, diagnostics, and cloud lifecycle.

## Boundary Model

### Keychain

Kocoro has a platform-backed keychain package with macOS-only production support and an in-memory test backend. StarClaw currently stores API keys in config/env paths and redacts responses, but it has no OS credential store abstraction.

Future StarClaw keychain work should start as a disabled-by-default `internal/keychain` package with explicit platform support checks and memory backend tests. It should not silently migrate existing config values into Keychain.

### Sync

Kocoro sync is a full local-to-cloud session pipeline: disabled defaults, marker file, lock file, candidate scanner, privacy-preserving batcher, dry-run outbox, cloud uploader, and audit events.

Future StarClaw sync work should be split so marker/dry-run foundations land before any cloud uploader. The first runtime task should be local-only and must prove idempotence, lock behavior, corrupt marker handling, and dry-run output.

### Migration

Kocoro Claude Code migration has a preview/apply split, plan store TTL, conflict detection, source fingerprints, and TOCTOU freshness checks. It preserves privacy by retaining only MCP env key names, never env values.

Future StarClaw migration work should begin with preview-only scanning and planning. Apply must be a later task with explicit approval, conflict handling, stale-plan rejection, and rollback notes.

### Upload / Share Privacy

StarClaw already has local upload IDs and local share manifests with retract support. Kocoro has richer cloud upload/share depth. Future work should define a single privacy contract before provider-backed delivery expands: what may leave the machine, how public URLs are warned, what metadata is persisted, and how redaction is tested.

## Non-Goals

- No new Go runtime packages in this task.
- No real credential storage.
- No sync daemon.
- No migration preview/apply endpoints.
- No cloud uploader or public artifact delivery.

## Output Contract

Create `.trellis/research/kocoro-keychain-sync-migration-discovery.md` with:

- Local baseline commit and evidence paths.
- Capability-by-capability gap analysis.
- Risk classification.
- Future task sequence.
- Acceptance criteria and validation commands for each future task.
- Updated overall Kocoro gap estimate after Phase 9.
