# Phase 5 secret leakage regression

## Goal

Add cross-surface regression coverage proving sensitive prompt, tool, provider, token, and secret-like data does not leak through observability, trace, summary, support, or handoff-style surfaces.

## Requirements

- Cover known risky fields: prompt text, assistant text, tool args, provider request/response bodies, API keys, bearer tokens, passwords, and secret-looking scalar values.
- Validate metrics, structured trace export/read, run summary metadata, support info, and Web UI trace/recovery output where applicable.
- Prefer reusable fixtures/helpers so future surfaces can be added to the same leak test set.

## Acceptance Criteria

- [x] Regression fixture injects prompt text, tool args, request/response-like payloads, and secret-like values.
- [x] Metrics and trace exports do not contain forbidden values or raw keys.
- [x] Run summary/recovery metadata remains aggregate-safe.
- [x] Web UI trace/recovery asset or rendered-smoke checks do not introduce unsafe raw payload display.
- [x] Targeted tests and full `go test ./...` pass.

## Non-Goals

- No prompt archive.
- No removal of intentional run detail Prompt/Result sections unless separately scoped.
- No external secret scanning service.
