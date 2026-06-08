# OpenAI gateway streaming parity implementation plan

## Checklist

1. Read backend specs and current OpenAI gateway tests.
2. Tighten streaming test helpers to count `[DONE]` and decode ordered SSE chunks.
3. Add regression coverage for:
   - exactly one `[DONE]`;
   - initial assistant role chunk before content;
   - incremental deltas from a streaming test client;
   - stop chunk only on success;
   - error frame without `[DONE]`.
4. Update README local runtime API wording.
5. Add a streaming curl example to `docs/EXAMPLES.md`.
6. Run validation:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-openai-gateway-streaming-parity`
   - `go test ./internal/daemon`
   - `go test ./...`
   - `rg -n "non-streaming|streaming.*unsupported|stream:true|text/event-stream" README.md docs/EXAMPLES.md internal/daemon/openai_api.go internal/daemon/openai_api_test.go`

## Risk Points

- Do not broaden OpenAI compatibility beyond the currently supported local gateway fields.
- Avoid tests that only check substring presence; ordered SSE contract matters.
- Do not duplicate final text after streamed deltas.

## Rollback

The implementation should be small and isolated to docs and OpenAI gateway tests. If code changes become necessary, keep them in `internal/daemon/openai_api.go` and validate with existing daemon tests.
