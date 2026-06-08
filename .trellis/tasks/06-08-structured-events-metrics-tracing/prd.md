# Structured events metrics and tracing

## Goal

Add structured runtime events and metrics/tracing foundations so Astria can move toward Kocoro/Shannon-style observability beyond browser-only run cards.

## Requirements

- Define structured event records for run lifecycle, tool calls, model calls, budget decisions, fallback decisions, and errors.
- Provide a local metrics endpoint or export surface for counters/gauges that can later map to Prometheus/OpenTelemetry.
- Avoid paid provider calls and avoid exporting secrets or prompt bodies by default.
- Preserve existing SSE/Web UI behavior.
- Keep tracing local-first and opt-in for detailed payloads.

## Acceptance Criteria

- [ ] Structured event schema is documented and test-covered.
- [ ] Runtime emits events for run start/end/error and budget/fallback decisions where implemented.
- [ ] Metrics endpoint/export returns stable counters without secrets.
- [ ] Existing event stream and Web UI smoke tests continue to pass.
- [ ] Tests cover redaction and metric shape.

## Notes

- This is a backend/observability slice, not the final UI visual polish.
