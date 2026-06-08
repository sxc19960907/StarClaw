# Astria Phase 5 integrated hardening and E2E validation

## Goal

Validate and harden Astria's completed Kocoro/Shannon parity platform slices as one integrated local product before opening the next capability phase.

Phase 5 is not a new feature-pile phase. It is the acceptance layer for Phase 3 and Phase 4: token budgets, OpenAI-compatible local API, runtime routing/fallback, structured observability, workflow control, durable recovery, replay approval, pause/resume, trace export, and Mission Control UI must work together end to end.

## Requirements

- Preserve StarClaw CLI/module/package/release names; Astria remains product-facing Web UI naming only.
- Keep local-first behavior: no cloud sync, accounts, remote telemetry, or external collectors.
- Keep the embedded daemon Web UI architecture; do not introduce a frontend build pipeline for validation.
- Prefer fixing integration defects over adding new UI panels.
- Validate platform behavior with deterministic tests, browser smoke where appropriate, and documented manual checks when full automation is impractical.
- Keep prompt/tool args/provider payload/API keys/secrets out of metrics, traces, summaries, support info, and handoff surfaces.
- After hardening is complete, run a Kocoro gap audit and use it to decide Phase 6.

## Child Task Map

| Priority | Child Task | Purpose |
|---|---|---|
| P1 | `phase5-runtime-e2e-smoke` | Validate Home/Chat launch through run history, budget state, routing/fallback, workflow control, replay approval, pause/resume, trace, and recovery display. |
| P1 | `phase5-api-observability-smoke` | Validate OpenAI-compatible gateway, `/metrics`, structured events, trace read/export, and workflow-control API compatibility together. |
| P1 | `phase5-secret-leakage-regression` | Add cross-surface regression checks that prompt text, tool args, provider payloads, tokens, and secrets do not leak through observability or handoff surfaces. |
| P2 | `phase5-webui-bug-bash` | Fix integration UX issues across Mission Control, Run detail, Budget, Quality, Reuse, Share, Memory, and Recovery navigation. |
| P2 | `phase5-docs-current-capabilities` | Update user-facing docs to reflect current Astria runtime/API/workflow capabilities and local-only safety boundaries. |
| P2 | `phase5-kocoro-gap-audit` | After Phase 5 validation, produce a fresh Kocoro/Shannon gap matrix and Phase 6 recommendation. |

## Acceptance Criteria

- [ ] Phase 5 parent lists validation/hardening before any new capability expansion.
- [ ] Each child has PRD acceptance criteria, clear non-goals, and validation commands.
- [ ] Runtime E2E checks exercise the Phase 3/4 platform features together, not as isolated unit tests only.
- [ ] Observability and API checks prove local compatibility without cloud telemetry.
- [ ] Secret-leakage regression covers metrics, traces, summaries, support/handoff-style output, and Web UI trace/recovery surfaces.
- [ ] Docs explain the current Astria capability set, including limitations and local-only boundaries.
- [ ] Final Kocoro gap audit is based on validated behavior after hardening, not assumptions.

## Non-Goals

- No Phase 6 capability implementation inside this parent.
- No external hosted service integration.
- No new remote sync or team account model.
- No broad visual redesign unless a bug blocks validation.

## Goal

TBD.

## Requirements

- TBD

## Acceptance Criteria

- [ ] TBD

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
