# Phase 5 API observability smoke

## Goal

Validate local daemon API compatibility and observability surfaces together: OpenAI-compatible gateway, structured events, metrics, trace read/export, and workflow control.

## Requirements

- Exercise `POST /v1/chat/completions` as a local compatibility entry point.
- Verify `/metrics`, `/runs/{id}/trace`, `/traces/export`, and `/runs/{id}/control` stay compatible after Phase 4.
- Keep exports local-only and explicit.
- Do not introduce external collectors or remote telemetry.

## Acceptance Criteria

- [x] OpenAI-compatible request produces a valid local completion envelope and run record.
- [x] Metrics include aggregate-safe counts without prompt/provider payload leakage.
- [x] Trace read/export writes valid JSONL records and preserves redaction.
- [x] Workflow control API responses remain compatible with run-store control metadata.
- [x] Targeted daemon tests and full `go test ./...` pass after fixes.

## Non-Goals

- No full OpenAI API feature parity beyond currently scoped gateway behavior.
- No remote trace upload.
- No cloud account integration.
