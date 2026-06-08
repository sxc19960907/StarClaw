# Astria Phase 4 runtime durability and replay

## Goal

Close the next Kocoro/Shannon gap by moving Astria from a rich local workbench into a durable runtime: long-running work should survive daemon restarts, expose resumable state, and eventually support safe replay with explicit approval boundaries.

## Requirements

- Preserve StarClaw CLI/module/package/release names; Astria remains the product-facing Web UI name.
- Keep the embedded static daemon Web UI architecture unless a child explicitly scopes API changes.
- Prioritize backend runtime durability before additional UI panels.
- Make each child independently testable, commit-ready, and safe to archive.
- Treat replay as dangerous by default: no replay execution may repeat tool calls or external effects without an explicit approval boundary.
- Keep local-first behavior; do not introduce cloud sync, accounts, or remote durable state in this phase.

## Child Task Map

| Priority | Child Task | Purpose |
|---|---|---|
| P1 | `06-08-durable-workflow-run-store` | Persist run records/events/control metadata to local disk so Mission Control can recover recent run state after daemon restart. |
| P1 | `durable-workflow-step-state` | Introduce step-level workflow state for launched missions, with status transitions and safe restart semantics. |
| P1 | `safe-replay-execution-boundary` | Convert replay plans into approved replay launches that preserve destructive/external approval gates. |
| P2 | `runtime-pause-resume-support` | Replace staged pause/resume `409` behavior with real cooperative pause/resume where the runtime can safely honor it. |
| P2 | `observability-trace-export` | Export structured run traces to a local JSONL/OpenTelemetry-ready artifact without prompt or secret leakage. |
| P2 | `runtime-recovery-ui` | Surface recovered durable runs, restart state, and replay approvals in Astria Mission Control. |

## Acceptance Criteria

- [ ] Parent task lists runtime durability and safe replay work before UI follow-up.
- [ ] Each child has PRD acceptance criteria and clear non-goals.
- [ ] The first slice persists local run state without exporting prompts/secrets through metrics or traces.
- [ ] Replay remains approval-required until a child explicitly implements approved replay execution.
- [ ] Implemented children pass targeted daemon tests and full `go test ./...`.

## Notes

- This phase starts from the Phase 3 control API: `pause`/`resume` are currently staged unsupported, and `replay` returns an approval-required plan only.
