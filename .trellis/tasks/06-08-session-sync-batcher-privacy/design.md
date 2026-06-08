# Session sync batcher privacy design

## Architecture

Extend `internal/sync` with:

- `scanner.go`: StarClaw session directory discovery and candidate selection.
- `batcher.go`: load/strip/size-pack candidates into local `BatchRequest` values.
- `strip_thinking.go`: structural JSON transform for assistant content blocks.

This task will not call `DryRunUploader.Send`. It prepares batches for a future runner.

## StarClaw Session Layout

StarClaw stores sessions as JSON files:

- default: `<starclawDir>/sessions/<session-id>.json`
- named agent: `<starclawDir>/sessions/<agent>/<session-id>.json`

This differs from Kocoro's `agents/<name>/sessions` layout and SQL index. StarClaw discovery should match StarClaw's existing daemon layout instead of copying Kocoro's directory shape.

## Candidate Contract

`Candidate` fields:

- `Dir`
- `AgentName`
- `SessionID`
- `UpdatedAt`
- `Source`

Discovery should:

- parse session JSON enough to read `id` and `updated_at`;
- skip invalid/unreadable session files and return a skipped count;
- include files with `UpdatedAt.After(marker.LastSyncAt)`;
- skip permanent failed entries unless `UpdatedAt.After(LastObservedUpdatedAt)`;
- union due transient failures from `marker.Failed`;
- dedupe by `SessionID`, keeping the freshest candidate;
- apply agent/source excludes after dedupe.

## Batching Contract

`BuildBatches(ctx, candidates, loader, cfg, marker, now)`:

- checks context per candidate;
- loads raw session JSON;
- records `load_error` as a permanent local failure;
- strips thinking blocks before size checks;
- records `size_limit_exceeded` as a permanent local failure;
- splits by `BatchMaxSessions` and `BatchMaxBytes`;
- returns `[]BatchRequest`.

`SessionEnvelope.ID` should be the candidate id, and `Session` should hold post-strip JSON.

## Privacy

Strip only structured content blocks with:

- `"type": "thinking"`
- `"type": "redacted_thinking"`

The helper must target assistant messages and leave all other message content unchanged. If the session JSON does not use content arrays or cannot be parsed, the helper should return the original bytes and an error only for parse failure. The batcher should continue with original bytes on strip parse errors because a corrupt local file should not block unrelated sessions.

## Local-Only Boundary

No network imports, no cloud client types, no daemon registration, and no uploader response handling in this task.
