# Implementation Plan

## Checklist

1. Add frontend helpers for copying current run result and tool result payloads.
2. Add `Copy result` to Run detail actions.
3. Add per-tool-card `Copy result` button for grouped tool results.
4. Change Agent Test stream renderer to store stream events and re-render with grouped timeline helpers.
5. Add minimal CSS for tool card action alignment.
6. Extend Web UI smoke:
   - core smoke verifies Run detail `Copy result`.
   - run detail mocked tool event verifies tool result copy.
   - targeted tool-call smoke verifies real tool result copy.
7. Run validation commands.

## Validation Commands

- `scripts/smoke_webui_core.sh`
- `scripts/smoke_webui_tool_call.sh`
- `go test ./...`
- `go vet ./...`
- `git diff --check`

## Risk Points

- `data-*` copy payloads must be escaped because tool results may contain quotes or JSON.
- Agent Test streaming must preserve current text streaming behavior while changing event rendering.
- Existing smoke selectors should not become ambiguous after adding buttons.
