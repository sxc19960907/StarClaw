# Kocoro keychain, sync, migration discovery

Date: 2026-06-08

## Local Kocoro baseline

- Path: `/Users/timmy/PycharmProjects/Kocoro`
- Commit: `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`
- Remote: `https://github.com/Kocoro-lab/Kocoro.git`
- Use this checkout for future comparisons unless the user explicitly asks to refresh from GitHub.

## Scope

This discovery covers the remaining high-risk Kocoro parity areas after Phase 9's native tool work:

- OS keychain storage.
- Session sync marker, scanner, batcher, uploader, and audit lifecycle.
- Claude Code migration preview/apply flow.
- Upload/share privacy and cloud-delivery boundaries.

It is intentionally planning-only. StarClaw should not enable credential writes, sync uploads, migration apply, or off-machine delivery until each area has an explicit opt-in design and independently verified child task.

## Evidence Summary

### Keychain

Kocoro evidence:

- `internal/keychain/keychain.go`
- `internal/keychain/backend_darwin.go`
- `internal/keychain/backend_other.go`
- `internal/keychain/backend_mem.go`
- `internal/keychain/keychain_test.go`

Kocoro capability:

- Wraps OS credential storage behind a `Backend` interface.
- Uses macOS Keychain in production through `zalando/go-keyring`.
- Uses an in-memory backend for tests.
- Returns `ErrUnsupportedPlatform` on non-darwin builds.
- Separates API keys and active-user state:
  - service `ai.kocoro.daemon.api_key`
  - service `ai.kocoro.daemon.state`
  - account `current_user_id`
  - account `legacy`
- Supports active user lookup, set, legacy rename, API key delete, and active user clear.
- Does not treat missing entries as fatal for normal reads.

StarClaw coverage:

- `internal/config/config.go` and `internal/config/multilevel.go` support config/env API keys.
- daemon config APIs redact secrets in responses.
- permissions classify sensitive files such as `.env` and keychain-like paths.
- no `internal/keychain` package exists.

Gap:

- StarClaw has redaction and config layering, but no OS-backed secret store, no active credential pointer, no memory backend for credential tests, and no migration path from config-file secrets to an OS store.

Risk:

- High. A naive implementation could silently move secrets, prompt OS Keychain unexpectedly, break non-darwin behavior, or leak key material through diagnostics/config APIs.

### Session Sync

Kocoro evidence:

- `internal/sync/config.go`
- `internal/sync/marker.go`
- `internal/sync/sync.go`
- `internal/sync/scanner.go`
- `internal/sync/batcher.go`
- `internal/sync/uploader.go`
- `internal/sync/strip_thinking.go`
- sync package tests.

Kocoro capability:

- Default config keeps sync disabled: `sync.enabled=false`.
- Uses local `sync_marker.json` with atomic writes.
- Sidecars corrupt or unknown-version marker files instead of crashing.
- Uses `sync.lock` with flock to serialize daemon and CLI sync callers.
- Scans sessions by updated watermark plus eligible failed retry entries.
- Handles permanent failures with a no-churn rule: retry only after local edits.
- Batches by max session count, max batch bytes, and single-session bytes.
- Strips `thinking` and `redacted_thinking` blocks before upload.
- Supports dry-run uploader that writes local outbox files and synthesizes accepted responses.
- Supports cloud uploader with defensive response normalization.
- Audits outcomes: ok, partial, transport error, noop.

StarClaw coverage:

- local session storage and run/event persistence exist.
- structured events/metrics/tracing and token budget phases are already in place.
- `internal/share` has local manifest and retraction semantics.
- no `internal/sync` package exists.

Gap:

- StarClaw does not yet have a sync marker, sync lock, scanner, batcher, dry-run outbox, upload protocol, retry marker, or privacy-preserving session transformer.

Risk:

- High. This path can move full session history off-machine. It must start with local-only marker/dry-run work and explicit user opt-in before any cloud uploader.

### Claude Code Migration

Kocoro evidence:

- `internal/daemon/migrate_claudecode.go`
- `internal/migrate/claudecode/types.go`
- `internal/migrate/claudecode/planner.go`
- `internal/migrate/claudecode/applier.go`
- scanner/converter/fingerprint/plan-store tests.

Kocoro capability:

- HTTP preview endpoint scans Claude Code source paths and returns a plan.
- HTTP apply endpoint requires a plan id from preview.
- Plan store has TTL and sweep behavior.
- Plan includes conflict detection, source fingerprints, symbolic paths, warnings, and plan hash.
- Apply checks plan freshness to reject stale plans after source changes.
- Migration categories include skills, agents, commands, global rules, and MCP servers.
- Privacy invariant: MCP env values are parsed only transiently; retained/returned/copied/logged data contains env key names only.
- MCP servers with missing env keys or unsupported fields are imported disabled.

StarClaw coverage:

- StarClaw already has agent command editing and MCP config APIs with secret-preserving update behavior.
- daemon config responses redact MCP env values and expose env key names.
- no `internal/migrate` package exists.
- no Claude Code preview/apply endpoint exists.

Gap:

- StarClaw lacks migration scanning, preview planning, plan TTL store, conflict summaries, fingerprint freshness, and apply semantics.

Risk:

- High. Migration touches user-owned configuration and can accidentally copy secrets or overwrite existing StarClaw assets.

### Upload / Share Privacy

Kocoro evidence:

- richer upload/share/cloud paths exist outside this discovery's detailed code read.
- sync and image tool boundaries show explicit provider/cloud warnings and disabled-by-default behavior.

StarClaw coverage:

- `internal/uploads/uploads.go` stores local uploads by generated id.
- `internal/share/store.go` records local share artifacts and supports retraction.
- Phase 8 and Phase 9 added cloud lifecycle and provider-gated image boundaries without default cloud credentials.

Gap:

- StarClaw has local share/upload mechanics but not a unified privacy contract for provider-backed public URLs, sync uploads, migration-generated artifacts, and cross-device cloud delivery.

Risk:

- Medium-high. The implementation surface is smaller than sync/migration, but user trust depends on consistent warnings, redaction, manifest metadata, and retraction semantics.

## Recommended Next Plan

### Phase 10: privacy and data-movement foundation

Goal: land the safe local foundations for secrets, sync, migration preview, and upload/share privacy without enabling real off-machine behavior by default.

Recommended child tasks:

1. `keychain-store-boundary`
   - Add StarClaw `internal/keychain` with backend interface, memory backend tests, darwin OS backend, and non-darwin unsupported behavior.
   - Add explicit service/account constants using StarClaw names.
   - Do not migrate existing config secrets automatically.
   - Acceptance:
     - missing entries read as empty where appropriate;
     - non-darwin returns unsupported;
     - memory backend covers set/get/delete/active-user behavior;
     - no daemon config endpoint leaks stored values.
   - Validation:
     - `go test ./internal/keychain ./internal/config ./internal/daemon`

2. `sync-marker-dry-run-foundation`
   - Add local-only sync config defaults, marker read/write, corrupt sidecars, flock lock, and dry-run outbox writer.
   - Keep `sync.enabled=false` by default.
   - Do not add a cloud uploader.
   - Acceptance:
     - marker missing/corrupt/unknown-version cases are covered;
     - writes are atomic;
     - lock contention is noop, not fatal;
     - dry-run writes 0600 outbox files under `.starclaw`;
     - no network calls exist.
   - Validation:
     - `go test ./internal/sync ./internal/config`

3. `session-sync-batcher-privacy`
   - Add candidate discovery and batching over StarClaw session stores.
   - Strip thinking/redacted-thinking content before dry-run output.
   - Add size caps, failed marker entries, transient/permanent classification, and no-churn permanent retry behavior.
   - Acceptance:
     - privacy stripping happens before size checks;
     - oversized/load-error sessions are recorded without aborting the run;
     - failed retries are deterministic;
     - no secrets appear in audit or dry-run summary output.
   - Validation:
     - `go test ./internal/sync ./internal/session ./internal/daemon`

4. `claude-code-migration-preview`
   - Add preview-only scanner/planner for Claude Code skills, agents, commands, global rules, and MCP server definitions.
   - Return symbolic paths by default.
   - Retain MCP env key names only, never values.
   - Import plans only as previews; no apply endpoint in this task.
   - Acceptance:
     - missing env values are reported as key names;
     - unsupported MCP fields are surfaced;
     - conflicts with existing StarClaw targets are detected;
     - preview results do not contain env values or source file secret contents.
   - Validation:
     - `go test ./internal/migrate/claudecode ./internal/daemon`

5. `migration-apply-approval-boundary`
   - Add apply endpoint after preview exists.
   - Require plan id, TTL validation, fingerprint freshness, conflict safety, and explicit user approval in daemon mode.
   - Acceptance:
     - stale plans reject with conflict;
     - expired/missing plan ids return distinct errors;
     - existing targets are not overwritten;
     - MCP servers with missing env keys are written disabled;
     - apply never copies env values.
   - Validation:
     - `go test ./internal/migrate/claudecode ./internal/daemon`

6. `cloud-upload-share-privacy-contract`
   - Define and enforce one provider-backed data movement contract for sync, image/public artifacts, and share delivery.
   - Keep provider credentials disabled until explicitly configured.
   - Add warnings and manifest metadata for public/permanent URLs.
   - Acceptance:
     - local-only mode remains the default;
     - public URL creation requires explicit approval/configuration;
     - retraction semantics are documented and tested;
     - logs/events/config responses redact credential material.
   - Validation:
     - `go test ./internal/share ./internal/uploads ./internal/tools ./internal/daemon`

### Phase 11: real opt-in cloud and product integration

Start this only after Phase 10 local foundations are complete.

Recommended scope:

- Cloud sync uploader behind explicit `sync.enabled=true` and configured endpoint/API key.
- UI controls for dry-run review, enablement, and last sync status.
- Migration preview/apply UI.
- Credential onboarding flow that can choose config/env/keychain storage explicitly.
- Cloud upload/share provider settings with warnings and retraction controls.

### Phase 12: memory and deeper product parity

Start after cloud/data movement contracts are stable.

Recommended scope:

- Kocoro-style deeper memory package parity.
- Cross-device/session resume if sync is enabled.
- Richer upload/share lifecycle polish.
- Final Astria stellar workbench UI language pass.

## Hard Privacy And Security Invariants

- StarClaw must stay local-first by default.
- Sync must default to disabled.
- Dry-run must precede cloud upload work.
- OS Keychain writes require explicit opt-in and platform support checks.
- Migration preview must precede migration apply.
- Migration must never retain, return, log, hash, or copy MCP env values.
- Config/API responses must expose secret key names only where needed, never values.
- Off-machine data movement requires explicit user approval/configuration.
- Public or permanent artifact URLs require visible warnings and tested retraction semantics.
- Corrupt local state files should be sidecarred and reset, not used unsafely.

## Updated Gap Estimate After Phase 9

Phase 9's first five children close major native desktop/tool gaps: Desktop RPC calendar protocol, calendar tools, browser lease handoff, terminal workspace tool, and provider-gated image tools.

Estimated alignment against local Kocoro `74cdb3c` after Phase 9:

- Local daemon/API/runtime control foundations: 85-90%.
- OpenAI-compatible local API and streaming: 80-90%.
- Native desktop/tool depth: 65-75%.
- Calendar/system integration: 75-85%.
- Browser runtime ownership depth: 70-80%.
- Terminal workspace depth: 55-65%.
- Image generation/editing boundary: 65-75%, with real provider behavior still opt-in.
- Keychain/sync/migration/cloud-product depth: 20-30%.
- Upload/share privacy and cloud delivery: 45-55%.

Overall StarClaw is now much closer on local runtime and native tool depth. The remaining gap is concentrated in data movement and product trust infrastructure: credentials, sync, migration, cloud upload/share, and memory/cloud integration. The next practical move is Phase 10, not UI polish.
