# Design

## Current StarClaw State

StarClaw already has:

- `MEMORY.md` file management through `GET /memory`, `POST /memory`, and `DELETE /memory/{name}`.
- Memory taxonomy parsing and warnings in `internal/daemon/memory_api.go`.
- `memory` and `memory_append` tools.
- Agent memory loading into `<agent_memory>` in `internal/agent/loop.go`.

StarClaw does not yet have:

- memory provider status
- local bundle/current metadata
- memory recall API
- sidecar readiness state
- private preflight injection that is sent to the model but not persisted

## Kocoro Evidence

Kocoro's README describes:

- `memory.provider`: disabled/cloud/local.
- local sidecar over Unix socket.
- bundle root/current bundle lifecycle.
- implicit episodic preflight before the first main-model call.
- `<private_memory>` injected into the current user message only.
- content-free `memory_preflight` audit fields.

Kocoro's `internal/memory` package includes sidecar lifecycle, bundle metadata, typed query intents, health payloads, response envelopes, and service status. This StarClaw child mirrors the contracts locally without real cloud pull or `tlm` process management.

## Proposed Architecture

Add daemon memory sidecar foundation:

- `internal/daemon/memory_sidecar.go`
  - provider/status types
  - bundle status discovery
  - local recall service over existing parsed MEMORY.md facts
  - safe private memory block formatter

- `internal/daemon/memory_sidecar_test.go`
  - status, bundle, recall, redaction, no-content telemetry tests

- `internal/daemon/memory_api.go`
  - add `GET /memory/status`
  - add `POST /memory/recall`

- `internal/daemon/router.go`
  - register new memory routes.

- `internal/agent/loop.go`
  - add a `MemoryPreflightProvider` interface and setter.
  - call it once before the first model request.
  - append `<private_memory>` to the in-memory user message only.
  - keep the original user query for session title and persisted session messages.

- `internal/daemon/runner.go`
  - wire daemon memory provider into the loop when available.

## Data Flow

1. Daemon builds a memory sidecar foundation from `s.memoryDir()`.
2. `GET /memory/status` reports local readiness:
   - disabled when memory dir is unavailable or no memory facts/bundle exists
   - ready when MEMORY.md/facts or current bundle metadata is available
   - unavailable/degraded for malformed bundle metadata
3. `POST /memory/recall` parses the query, searches memory taxonomy facts, and returns ranked local matches.
4. `RunAgentWithApproval` receives a preflight provider from daemon deps.
5. Agent loop calls provider once with the original user query.
6. Provider returns a private block and content-free stats.
7. Loop sends the private block to the model but persists only the original user query.

## Privacy Contract

The private memory block may contain recalled memory because it is model input. The following must remain content-free:

- run structured events
- metrics
- trace export
- memory preflight telemetry
- status endpoints

The following must not contain private memory at all:

- session messages saved to disk
- compaction summaries
- run `Prompt` / `Request.Text`

## Rollback

Remove memory sidecar routes and do not set an agent loop memory preflight provider. Existing `MEMORY.md` management and tools continue to work.
