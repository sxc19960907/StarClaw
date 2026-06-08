# Phase 5 API observability smoke implementation plan

## Checklist

1. Load backend specs with `trellis-before-dev`.
2. Inspect OpenAI gateway, metrics, trace export, and workflow-control tests.
3. Add one integration-style daemon test for the gateway/observability/control path.
4. Keep assertions aggregate-safe; do not require prompt text in observability output.
5. Run validation:
   - `go test ./internal/daemon -run 'TestPhase5APIObservabilitySmoke|TestHandleOpenAIChatCompletions|TestHandleMetrics|TestHandleRunTrace|TestHandleExportTraces|TestHandleRunControl' -count=1`
   - `go test ./internal/daemon ./cmd`
   - `go test ./...`
   - `git diff --check`
6. Update PRD acceptance criteria, commit, archive child, and record journal.

## Risk Files

- `internal/daemon/openai_api_test.go`
- `internal/daemon/observability_test.go`
- `internal/daemon/server_test.go`
- `internal/daemon/openai_api.go`
- `internal/daemon/trace_export.go`
- `internal/daemon/server.go`

## Non-Goals

- No new API compatibility fields.
- No remote telemetry.
- No frontend changes.
