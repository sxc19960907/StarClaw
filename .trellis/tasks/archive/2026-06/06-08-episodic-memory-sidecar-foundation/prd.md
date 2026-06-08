# Episodic memory sidecar foundation

## Goal

Implement the next Phase6 Kocoro parity slice: add local-first episodic memory sidecar foundations so StarClaw can report memory provider readiness, read local bundle metadata, recall matching memory facts through a safe API, and inject private recall context into a run without persisting recalled content into transcripts or structured telemetry.

## Requirements

- Use local Kocoro evidence from `/Users/timmy/PycharmProjects/Kocoro/internal/memory/*` and README lines around memory sidecar, bundle, and implicit episodic preflight.
- Keep this child local-first:
  - no Kocoro/Shannon Cloud calls
  - no session upload/sync
  - no account or tenant API
  - no external telemetry
- Add a memory provider/status abstraction that can represent disabled, ready, unavailable, and degraded states.
- Add local bundle metadata discovery under StarClaw's memory directory. A valid local bundle must be represented by `memory/bundles/<version>/manifest.json` and `memory/current` or an equivalent safe current pointer.
- Add a recall API that uses local MEMORY.md/taxonomy and local bundle metadata as a fallback-friendly foundation. It must not require a real sidecar binary in this slice.
- Add a content-free preflight audit/trace signal for memory recall attempts: attempted, provider status, results count, context injected, outcome, and error class. Do not log query text or recalled content.
- Add an agent-loop preflight hook/interface that can inject a `<private_memory>` block into the in-memory user message sent to the model while preserving session persistence without that block.
- Recalled content must not be stored in session transcripts, run prompt fields, structured events, metrics, trace export, or compaction summaries.
- Preserve existing `GET /memory`, `POST /memory`, `DELETE /memory/{name}`, `memory` tool, and `memory_append` tool behavior.

## Acceptance Criteria

- [ ] `GET /memory/status` returns provider status, bundle readiness, current bundle metadata, and local fallback availability without exposing secrets or recalled content.
- [ ] `POST /memory/recall` accepts a query and returns safe local recall results from MEMORY.md/taxonomy when available.
- [ ] `POST /memory/recall` handles disabled/no-bundle/no-memory states with explicit reason codes rather than HTTP 500.
- [ ] Agent loop supports a memory preflight provider that injects `<private_memory>` into the model-facing message only.
- [ ] Session persistence tests prove `<private_memory>` and recalled content are not saved into session messages.
- [ ] Run-store/structured event tests prove memory preflight telemetry is content-free.
- [ ] Existing memory API/tool tests and full daemon tests remain compatible.
- [ ] Backend quality spec is updated with the memory sidecar/preflight contract.

## Notes

- This is a foundation slice. Real `tlm` process supervision, bundle pull, tenant wipe, and cloud sync belong to later tasks after the local privacy contract is proven.
