# Session sync batcher privacy

## Goal

Add StarClaw session candidate discovery and privacy-preserving batching on top of the local sync foundation, so future sync runs can dry-run prepared session batches without exposing assistant thinking blocks or moving data off-machine.

## Requirements

- Add candidate discovery over StarClaw's existing session directories:
  - default sessions under `.starclaw/sessions/*.json`;
  - named-agent sessions under `.starclaw/sessions/<agent>/*.json`.
- Discover sessions whose `updated_at` is after `marker.LastSyncAt`.
- Include eligible failed transient retry entries from `marker.Failed` when `NextAttemptAt` is due.
- Preserve the no-churn rule for permanent failures: do not retry unless the local session was edited after `LastObservedUpdatedAt`.
- Support `Config.ExcludeAgents` and `Config.ExcludeSources`; use `"default"` and `"local"` as canonical empty values.
- Add batching that:
  - loads session JSON by candidate;
  - strips `thinking` and `redacted_thinking` blocks before size checks;
  - enforces `SingleSessionMaxBytes`, `BatchMaxSessions`, and `BatchMaxBytes`;
  - records local load/size failures in `marker.Failed`;
  - returns dry-run `BatchRequest` values using existing local batch shape.
- Keep this task local-only:
  - no cloud uploader;
  - no daemon/CLI runner;
  - no marker watermark advancement;
  - no background sync.

## Acceptance Criteria

- [ ] Candidate discovery finds default and named-agent StarClaw session JSON files.
- [ ] Candidate discovery filters by `LastSyncAt`, excludes configured agents/sources, and dedupes by session id with the freshest `UpdatedAt`.
- [ ] Permanent failures are skipped until a newer local edit appears.
- [ ] Due transient failures are reintroduced deterministically.
- [ ] Batching strips `thinking` and `redacted_thinking` before size checks.
- [ ] Batching records `load_error` and `size_limit_exceeded` failures without aborting the whole batch build.
- [ ] Batch max session and byte limits split batches deterministically.
- [ ] No `internal/sync` cloud/network uploader is added.
- [ ] `go test ./internal/sync ./internal/session ./internal/daemon` passes.

## Out of Scope

- Sending batches to any remote service.
- Advancing marker watermarks from accepted upload responses.
- Daemon scheduler, CLI command, or UI controls.
- Redacting arbitrary user message text; this task strips only thinking block structures.
