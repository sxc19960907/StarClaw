# Phase 5 Kocoro gap audit

Date: 2026-06-08

## Summary

Phase 5 closed the original platform-alignment slices the user prioritized: token budget hard stops, OpenAI-compatible local API, deterministic routing/fallback, local structured observability, trace export, durable run/workflow metadata, pause/resume/cancel/replay workflow control, redaction regression coverage, and Astria runtime UI hardening.

The older May 2026 StarClaw-vs-ShanClaw research is now partly stale. Local StarClaw evidence shows several previously missing areas now exist: desktop/browser tools, MCP stdio server mode, streaming provider clients, thinking/reasoning config, memory APIs, council handoff, and cloud delegation primitives. The remaining gap is less "missing platform foundation" and more "depth, orchestration, and product-grade parity": richer desktop automation, fully compatible OpenAI streaming/tool APIs, true parallel multi-agent workflows, stronger knowledge lifecycle, external delivery connectors, and clearer evidence on Kocoro/Shannon behavior.

## Status matrix

| Area | Status | Evidence | Gap / note |
|---|---|---|---|
| Token budget enforcement and hard stop | aligned | `agent.token_budget` is documented in `README.md:173`; runtime setters and pre-call enforcement are in `internal/agent/loop.go:247`, `internal/agent/loop.go:397`, and `internal/agent/loop.go:512`; budget tests cover hard-stop decisions in `internal/agent/token_budget_test.go`. | Local runtime budget tracking is implemented. Kocoro/Shannon parity beyond this local contract is unknown. |
| OpenAI-compatible local API gateway | partially aligned | Route registration is in `internal/daemon/router.go:70`; handler maps requests through daemon run execution in `internal/daemon/openai_api.go:60`; docs describe `/v1/chat/completions` in `README.md:122`; tests live in `internal/daemon/openai_api_test.go`. | Non-streaming chat completion compatibility is implemented. `stream=true`, tool/function calling, `response_format`, metadata, and `n > 1` remain explicitly unsupported in `internal/daemon/openai_api.go:162`. |
| Runtime complexity routing and fallback | aligned | Deterministic metadata is documented in `README.md:127` and `docs/CONFIGURATION.md:141`; classifier/fallback code is in `internal/agent/routing.go`; daemon attaches metadata in `internal/daemon/runner.go`; tests cover routes and fallback in `internal/agent/routing_test.go` and `internal/daemon/runner_test.go`. | This meets the local deterministic routing slice. It is advisory metadata, not hosted routing or automatic external escalation. |
| Structured events, metrics, and tracing | aligned | Routes are in `internal/daemon/router.go:53`, `internal/daemon/router.go:65`, and `internal/daemon/router.go:67`; structured events live in `internal/daemon/events.go`; trace export is in `internal/daemon/trace_export.go`; docs describe local-only metrics/trace in `README.md:124` and `docs/CONFIGURATION.md:148`; tests are in `internal/daemon/observability_test.go`. | Local aggregate metrics and explicit JSONL export are in place. There is intentionally no external collector or hosted telemetry. |
| Workflow control API: cancel, pause, resume, replay | aligned | Routes are in `internal/daemon/router.go:60` and `internal/daemon/router.go:66`; active runtime handles are cancellable and pausable in `internal/daemon/server.go:158`; control dispatch is in `internal/daemon/server.go:288`; cooperative pause points are in `internal/agent/loop.go:404` and `internal/agent/loop.go:462`; replay tests are in `internal/daemon/server_test.go`. | Phase 5 now supports real cooperative pause/resume for active runs and approval-gated replay launch. Persisted process resurrection after daemon death remains out of scope. |
| Durable run recovery and workflow step state | aligned | Runtime metadata is documented in `README.md:127`; run records persist budget/routing/fallback/control/steps/structured events in `internal/daemon/run_store.go`; persistence tests are in `internal/daemon/run_store_persistence_test.go`; workflow step tests are in `internal/daemon/workflow_step_state_test.go`. | Durable metadata recovery exists. Deterministic replay of historical tool outputs is not implemented and should remain a separate capability. |
| Secret leakage and redaction | aligned | Redaction boundary is documented in `README.md:281` and `docs/CONFIGURATION.md:149`; structured observability sanitization is tested in `internal/daemon/observability_test.go`; replay-control leakage regression is tested in `internal/daemon/server_test.go`; Phase 5 leakage task was archived under `.trellis/tasks/archive/2026-06/06-08-phase5-secret-leakage-regression`. | Observability/support-style surfaces are covered. Operator-facing local Prompt/Result panels remain intentionally detailed review surfaces. |
| Astria runtime Web UI | partially aligned | Current docs describe Astria runtime panels in `README.md:98`; embedded routes are in `internal/daemon/router.go:45`; runtime trace/recovery rendering exists in `internal/daemon/webui/assets/app.js`; Web UI route/assets tests are in `internal/daemon/webui_test.go`; Phase 5 Web UI bug bash is archived under `.trellis/tasks/archive/2026-06/06-08-phase5-webui-bug-bash`. | Runtime surfaces are usable and hardened, but deeper product polish, accessibility, multi-viewport visual QA, and end-user workflow design still trail a mature Kocoro-like workbench. |
| Browser and desktop automation tools | partially aligned | Tool inventory documents desktop/macOS tools in `README.md:136`; implementations exist in `internal/tools/browser.go:14`, `internal/tools/computer.go:15`, `internal/tools/accessibility.go:15`, and `internal/tools/screenshot.go:15`; tests exist beside each tool. | The category is no longer missing, but implementation is simpler than the old ShanClaw baseline: no local evidence of a native Swift AX server, advanced Playwright/CDP orchestration, or robust cross-app visual automation depth. |
| MCP client and server | aligned | Docs describe MCP server mode in `README.md:394` and `docs/CONFIGURATION.md:90`; CLI command is wired in `cmd/root.go:703`; server wrapper is in `internal/tools/mcp_server.go:20`; client/readiness/supervisor code exists under `internal/mcp/`. | Local MCP client plus stdio server exists. Deeper protocol compatibility and production hardening versus Kocoro/Shannon remain unknown without external evidence. |
| Streaming, thinking, reasoning, and model overrides | partially aligned | Streaming interface is in `internal/agent/loop.go:73`; loop enables thinking/reasoning/model options in `internal/agent/loop.go:227`, `internal/agent/loop.go:232`, and `internal/agent/loop.go:237`; client streaming and thinking request fields are in `internal/client/client.go:214` and `internal/client/client.go:289`; config defaults are documented in `README.md:177`. | Provider/daemon `/message` streaming exists, but the OpenAI-compatible local gateway remains non-streaming and does not expose OpenAI-style streaming or tool-call deltas. |
| Multi-agent collaboration / council | partially aligned | Council routes are registered in `internal/daemon/router.go:119`; council planning and handoff are implemented in `internal/daemon/council_api.go:118`; docs list Agent Council in Astria surfaces in `README.md:98`; tests are in `internal/daemon/council_api_test.go`. | Current council is a deterministic planner/synthesis and single-run handoff surface, not a true parallel multi-agent runtime with independent worker lifecycles, conflict resolution, and shared state. |
| Memory / knowledge | partially aligned | Memory routes are in `internal/daemon/router.go:113`; memory API supports list/append/delete in `internal/daemon/memory_api.go:53`; agent memory injection and persistence are in `internal/agent/loop.go:320` and `internal/agent/loop.go:350`; docs list memory tools in `README.md:146` and `README.md:149`. | Durable local memory exists, but knowledge curation is still mostly local file/API driven. Source registry, conflict reconciliation, provenance, and reviewed memory promotion need more product depth. |
| Reuse, delivery, and share workflows | partially aligned | Astria documents reuse assets and local share-pack drafting in `README.md:98`; Web UI surfaces include reuse/delivery/inbox/share concepts in `internal/daemon/webui/assets/app.js`; publish and inbox primitives exist in `internal/tools/publish_to_web.go` and `internal/daemon/inbox_api.go`. | These are local/review-oriented primitives. Real outbound channel delivery, team sharing, hosted artifact publishing, and external workflow connectors are not part of the local-first Phase 5 contract. |
| Cloud delegation / external agents | partially aligned | Cloud client and tool primitives exist in `internal/client/cloud.go` and `internal/tools/cloud_delegate.go`; tests cover streamed cloud delegate responses in `internal/client/cloud_test.go` and `internal/tools/cloud_delegate_test.go`. | Phase 5 intentionally kept local-first behavior. Cloud delegation is a primitive, not an integrated account/sync/team execution model. |
| Kocoro/Shannon current parity | unknown | The only local comparative evidence is the older `.trellis/research/gap-analysis-starclaw-vs-shanclaw.md`. Current Kocoro/Shannon code or docs are not present in this workspace. | Do not claim exact Kocoro/Shannon parity. Treat this audit as StarClaw/Astria capability maturity against locally known platform themes. |

## Completed Phase 5 plan items

The five user-prioritized platform tasks are complete at the local platform level:

- `token-budget-enforcement`: done and integrated into agent loop, daemon response/run metadata, docs, and tests.
- `openai-compatible-api-gateway`: done for non-streaming `/v1/chat/completions`; unsupported OpenAI fields are rejected explicitly.
- `runtime-complexity-routing-fallback`: done as deterministic local metadata and fallback recommendations.
- `structured-events-metrics-tracing`: done for aggregate `/metrics`, structured events, run trace read, and local JSONL export.
- `workflow-control-api`: done for cancel, active-run pause/resume, replay approval, and approved replay launch through daemon execution.

## Remaining gap size

After Phase 5, the gap to a Kocoro/Shannon-like platform is moderate, not foundational. The base runtime platform is now present. The remaining work clusters into four capability bands:

1. OpenAI gateway depth: streaming responses, streamed tool-call deltas, broader parameter compatibility, and client compatibility tests.
2. Agent orchestration depth: true multi-agent execution, worker lifecycle/state, conflict handling, role-specific artifacts, and durable collaboration traces.
3. Desktop/browser automation depth: native accessibility server maturity, browser/CDP/Playwright orchestration, visual state verification, and robust permission UX.
4. Knowledge/delivery product depth: reviewed knowledge lifecycle, source provenance, reusable result packs, external channel delivery, and share/export workflows.

## Phase 6 recommendation

Prioritize **OpenAI-compatible gateway depth plus agent orchestration**, in that order.

Rationale:

- The OpenAI-compatible gateway is the highest-leverage integration surface. It lets existing tools and clients treat StarClaw as a local model/runtime endpoint, but the current non-streaming slice blocks many real clients that expect streaming, tool calls, and richer compatibility.
- Agent orchestration is the highest platform-risk gap after runtime hardening. Astria has council planning and handoff, but not true multi-agent work execution. Kocoro-like parity will feel limited until workers, reviewer/planner roles, shared artifacts, and workflow state are real runtime concepts.
- Desktop/browser automation and knowledge/delivery should follow once gateway/orchestration are stable, because both depend on stronger execution boundaries, replay safety, and traceable multi-step state.

Recommended Phase 6 sequence:

1. `openai-compatible-streaming-tools`: add streaming chat-completions compatibility and a deliberate local contract for OpenAI tool/function fields.
2. `agent-orchestration-runtime`: turn council from planner/handoff into durable multi-agent execution with planner/researcher/implementer/reviewer roles and run-linked artifacts.
3. `desktop-browser-automation-depth`: harden browser/computer/accessibility around visual state, native AX depth, and reproducible tests.
4. `knowledge-delivery-lifecycle`: promote memory/source/reuse/share from local primitives into reviewed workflows with provenance and export/delivery boundaries.
