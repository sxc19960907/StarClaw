# Add GUI Tool Call Smoke

## Problem

The GUI has smoke coverage for chat, streaming, run history, and mocked run-detail tool events, but it does not yet verify the full real agent loop where a model requests a tool, the daemon executes it, the GUI displays the tool call/result, and Run detail persists the timeline.

## Scope

- Extend the local fake OpenAI-compatible provider used by Web UI smoke tests with a deterministic tool-call scenario.
- Use a safe built-in tool (`version`) so the test does not require approvals, filesystem writes, or external services.
- Add a targeted Web UI smoke script/mode for the real tool-call chain.
- Do not add this path to the default CI core smoke unless explicitly decided later.

## Acceptance Criteria

- [x] A targeted Web UI smoke mode/script triggers a real model tool call through daemon `/message`.
- [x] The daemon executes the `version` tool and returns the result to the provider loop.
- [x] The GUI displays the tool call and tool result while the run completes.
- [x] Run detail timeline groups `tool_call` and `tool_result` for the real run.
- [x] Session history can be opened and includes the final answer.
- [x] `scripts/smoke_webui_tool_call.sh` passes locally.
- [x] Existing `scripts/smoke_webui_core.sh` still passes locally.
- [x] `go test ./...` and `go vet ./...` pass locally.
