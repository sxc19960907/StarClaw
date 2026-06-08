# Observability trace export

## Goal

Export structured run traces to a local JSONL/OpenTelemetry-ready artifact without prompt or secret leakage.

## Requirements

- Export structured run events to local JSONL artifacts.
- Support exporting all stored runs and a single run.
- Keep exported records OpenTelemetry-ready: include schema version, run id, event id, event type, phase, timestamp, and attributes.
- Do not export prompt text, tool arguments, raw provider payloads, request/response bodies, API keys, tokens, or secrets.
- Preserve existing `/metrics`, `/runs`, and run detail behavior.
- Keep export local-only; do not introduce cloud upload, accounts, remote telemetry, or a new tracing dependency.
- Export writes must use safe local file semantics and return contextual errors.

## Acceptance Criteria

- [x] `RunStore` can produce JSONL trace output for all runs.
- [x] `RunStore` can produce JSONL trace output for one run.
- [x] Exported JSONL is one valid JSON object per structured event.
- [x] Exported records include OTel-ready fields and aggregate-safe attributes only.
- [x] Export output does not contain prompt text, tool args, provider payloads, or secrets.
- [x] Daemon exposes local trace export/read endpoints without changing existing metrics/runs behavior.
- [x] Missing run export returns a safe not-found error path.
- [x] Existing structured event, metrics, run-store persistence, and full tests continue to pass.

## Notes

- Parent task: `06-08-astria-phase-4-runtime-durability-replay`.
- This is local artifact export only; no external collector integration in this slice.
