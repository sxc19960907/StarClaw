# Astria Kocoro parity phase 8: local API streaming and runtime compatibility

## Goal

Close the next local-first Kocoro parity gaps after Phase 7 by improving local API streaming compatibility and runtime client behavior without introducing external cloud credentials or real off-machine channel transport.

## Current Evidence

- Local Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3c`.
- Kocoro documents and exercises daemon SSE behavior for `POST /message`.
- StarClaw already supports `POST /message` SSE and internal LLM streaming events.
- StarClaw's OpenAI-compatible gateway currently rejects `stream=true` in `internal/daemon/openai_api.go`.
- User explicitly called out that project replies do not support streaming output.

## Child Plan

1. `openai-compatible-streaming-deltas`
   - Add OpenAI-compatible SSE chunk responses for `POST /v1/chat/completions` with `stream=true`.
   - Keep tool/function call request fields unsupported in this slice.

2. `ws-controller-cloud-lifecycle`
   - Add local lifecycle control boundary for future cloud WebSocket start/stop/restart.
   - Keep cloud credentials and real connection disabled unless explicitly approved.

3. `streaming-runtime-edge-cases`
   - Improve stream error/fallback semantics, final event consistency, and partial text handling.

4. `native-tool-parity-followup`
   - Recompare Kocoro native tool depth after local API streaming is closed.

## Acceptance Criteria

- [ ] Child tasks are planned, implemented, validated, committed, and archived independently.
- [ ] Phase 8 remains local-first by default.
- [ ] Real cloud credentials/transports are not enabled without explicit approval.
