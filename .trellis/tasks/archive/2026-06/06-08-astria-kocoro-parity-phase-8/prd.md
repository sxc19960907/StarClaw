# Astria Kocoro parity phase 8: local API streaming and runtime compatibility

## Goal

Close the next local-first Kocoro parity gaps after Phase 7 by improving local API streaming compatibility and runtime client behavior without introducing external cloud credentials or real off-machine channel transport.

## Current Evidence

- Local Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3c`.
- Kocoro documents and exercises daemon SSE behavior for `POST /message`.
- StarClaw already supports `POST /message` SSE and internal LLM streaming events.
- StarClaw's OpenAI-compatible gateway now supports the happy path for `stream=true` in `internal/daemon/openai_api.go`.
- User explicitly called out that project replies did not support streaming output; the remaining risk is edge-case correctness and Kocoro-style stream robustness.
- Latest local comparison is recorded in `.trellis/research/kocoro-local-comparison-phase8-plan.md`.

## Child Plan

1. `openai-compatible-streaming-deltas`
   - Add OpenAI-compatible SSE chunk responses for `POST /v1/chat/completions` with `stream=true`.
   - Keep tool/function call request fields unsupported in this slice.

2. `ws-controller-cloud-lifecycle`
   - Add local lifecycle control boundary for future cloud WebSocket start/stop/restart.
   - Keep cloud credentials and real connection disabled unless explicitly approved.

3. `streaming-runtime-edge-cases`
   - Improve stream error/fallback semantics, final event consistency, duplicate final text suppression, and partial text handling.

4. `sse-reconnect-idle-watchdog`
   - Add Kocoro-style SSE idle timeout, bounded reconnect, and `Last-Event-ID` replay behavior where it fits StarClaw's local client/server API.

5. `native-tool-parity-followup`
   - Recompare Kocoro native tool depth after local API streaming is closed.
   - Produce the Phase 9 plan for native Desktop tools and local integration depth.

## Deferred Beyond Phase 8

- OpenAI-compatible tool/function call streaming deltas remain deferred. StarClaw currently keeps local daemon tools as the first-class tool path and rejects OpenAI `tools` / `functions` fields.
- Real cloud credentials, external channel delivery, sync, and keychain behavior remain out of scope until explicitly approved.
- Native Desktop RPC, calendar tools, browser handoff/leasing, Ghostty helpers, and image tooling should be grouped under a Phase 9 parent after streaming runtime compatibility is closed.

## Acceptance Criteria

- [x] Child tasks are planned, implemented, validated, committed, and archived independently.
- [x] Phase 8 remains local-first by default.
- [x] Real cloud credentials/transports are not enabled without explicit approval.
- [x] Phase 8 ends with a fresh Kocoro comparison and a concrete Phase 9 native tool/local integration plan.
