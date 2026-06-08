# Implementation Plan

## Scope

Add local-first memory sidecar/status/recall and private preflight injection foundations without cloud sync or real sidecar process supervision.

## Checklist

- [x] Run `trellis-before-dev`.
- [x] Add memory sidecar foundation tests for empty memory, MEMORY.md facts, valid bundle manifest, malformed bundle, recall matches, no-data reasons, private block formatting, and content-free telemetry.
- [x] Implement local `memory_sidecar.go`.
- [x] Add HTTP tests for `GET /memory/status` and `POST /memory/recall`.
- [x] Register memory routes and implement handlers.
- [x] Add agent-loop preflight tests proving private memory is sent to LLM input and not persisted to session messages.
- [x] Implement agent loop memory preflight provider and daemon wiring.
- [x] Add run-store structured event test for content-free `memory_preflight`.
- [x] Update backend quality spec with the memory sidecar/preflight contract.
- [x] Run focused and broader validation.

## Completion Notes

- Added local memory provider/status foundation with provider states, local bundle manifest discovery, fallback availability, and reason codes.
- Added `GET /memory/status` and `POST /memory/recall`.
- Added local recall over existing MEMORY.md taxonomy facts.
- Added agent-loop `MemoryPreflightProvider` support that injects `<private_memory>` into model input only.
- Added stripping so `<private_memory>` and recalled content are not saved to session transcripts.
- Added content-free `memory_preflight` structured events through the existing run recorder path.
- No cloud calls, session upload, external telemetry, accounts, or real sidecar process supervision were introduced.

## Validation Commands

```bash
go test ./internal/agent ./internal/daemon -run 'TestAgentLoopMemoryPreflight|TestStripPrivateMemoryBlock|TestMemory|TestHandleMemory|TestRouterRegistersRoutes|TestHandleMessageMemoryPreflight|TestMultiHandlerFanOut'
go test ./internal/agent
go test ./internal/daemon
go test ./...
git diff --check
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-episodic-memory-sidecar-foundation
```

## Review Gates

- No cloud calls or external telemetry.
- No recalled content in structured events, metrics, trace export, or session persistence.
- Existing `MEMORY.md` management remains compatible.
- Preflight injection is optional and disabled when memory status is not ready.
