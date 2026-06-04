# Add GUI Streaming Provider Smoke Harness

## Problem

GUI smoke currently covers mocked successful UI paths and provider-unavailable error runs, but it does not verify that the daemon, OpenAI-compatible streaming client, GUI chat output, session persistence, and run detail views work together with a real HTTP provider endpoint.

The local environment has no running Ollama service and no provider API keys, so the verification needs a deterministic local fake OpenAI-compatible provider.

## Scope

- Add a reusable local fake OpenAI-compatible streaming provider harness for Web UI smoke tests.
- Exercise the GUI through the daemon rather than mocking browser `/message` calls for this path.
- Verify streaming chat output, run summary, session persistence, and run detail data.
- Keep CI bounded; do not add this path to the default CI core smoke unless explicitly chosen later.

## Acceptance Criteria

- [x] A Web UI smoke mode/script starts a local fake OpenAI-compatible provider.
- [x] The daemon uses that fake provider through normal config with `provider: openai`.
- [x] The GUI sends a prompt through the real daemon `/message` route and receives streaming output.
- [x] The smoke verifies session persistence and Runs/detail content for the streamed run.
- [x] The smoke remains isolated from the user's real `~/.starclaw` data.
- [x] Targeted streaming provider smoke passes locally.
- [x] Existing `scripts/smoke_webui_core.sh` still passes locally.
- [x] `go test ./...` and `go vet ./...` pass locally.

## Notes

- Do not require external API keys or a local Ollama installation for this task.
- The fake provider should implement only the OpenAI-compatible endpoints needed for this smoke path.
- Validation completed with `scripts/smoke_webui_streaming.sh`, `scripts/smoke_webui_core.sh`, `go test ./...`, `go vet ./...`, and `git diff --check`.
