# Phase 5 API observability smoke design

## Scope

This child adds an integration-style daemon test that exercises API compatibility and observability surfaces together. It does not expand API feature scope.

## Smoke Path

1. Submit `POST /v1/chat/completions` with a deterministic request id.
2. Confirm response envelope includes OpenAI-compatible fields, usage mapping, and `starclaw_run_id`.
3. Confirm `/runs/{id}` has `source=openai-compatible`, structured events, usage, and no broken run-store metadata.
4. Confirm `/metrics` exposes aggregate counts and does not include prompt text.
5. Confirm `/runs/{id}/trace` returns OTel-ready trace records.
6. Confirm `/traces/export?path=...` writes valid local JSONL for all stored runs.
7. Confirm `POST /runs/{id}/control` replay remains approval-gated and records compatible control metadata.

## Safety

The smoke fixture includes prompt text that must not appear in metrics or trace export. Trace/read/export assertions check structural fields and redaction.

## Rollback

Remove the integration test and task artifacts. Runtime code should remain unchanged unless the smoke reveals a real defect.
