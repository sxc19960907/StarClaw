# Astria Kocoro parity phase 6: engine orchestration and channel runtime

## Goal

Move StarClaw/Astria from a hardened local runtime into Kocoro-like engine behavior: executable workflow commands, durable daemon message/channel foundations, memory sidecar readiness, Desktop RPC boundaries, deeper local tool orchestration, and share/sync lifecycle contracts.

## Source Evidence

- Local Kocoro checkout: `/Users/timmy/PycharmProjects/Kocoro` at `74cdb3c`.
- Comparison artifact: `.trellis/research/kocoro-local-comparison-phase6-plan.md`.
- Completed Phase5 audit: `.trellis/tasks/archive/2026-06/06-08-phase5-kocoro-gap-audit/gap-audit.md`.

## Requirements

- Preserve StarClaw as the CLI/module/package/release/docs name and Astria as the product-facing Web UI name.
- Use the local Kocoro checkout for parity comparison instead of repeatedly consulting GitHub.
- Keep the implementation local-first unless the user explicitly approves cloud sync, external telemetry, external channel providers, or account-based delivery.
- Treat Kocoro Desktop behavior as closed-source unless it is visible through open daemon contracts.
- Build Phase6 as independently verifiable child tasks; the parent owns planning, ordering, and final integration review.
- Prefer StarClaw's existing `/message`, SSE, run store, workflow step, routing, approval, observability, and Web UI contracts over parallel replacement paths.

## Child Task Order

1. `kocoro-style-workflow-orchestration`
   - Add Kocoro-style `/research` and `/swarm` workflow entry points over StarClaw daemon `POST /message` + SSE.
   - Upgrade Astria council from static planning handoff toward executable workflow orchestration.

2. `daemon-mailbox-channel-runtime`
   - Add typed queued messages, route keys, mailbox persistence, claim/ack lifecycle, busy/inject behavior, and bounded worker semantics.

3. `episodic-memory-sidecar-foundation`
   - Add local-first memory provider abstraction, bundle files, readiness state, and preflight recall injection.

4. `desktop-rpc-boundary`
   - Add Unix socket RPC broker/listener contracts for future native Desktop integration.

5. `desktop-browser-tool-depth`
   - Deepen browser, computer, accessibility, and screenshot tool behavior with leases, handoff, and visual verification.

6. `share-sync-delivery-lifecycle`
   - Add local share artifact rendering, retractable publishing abstraction, and opt-in sync dry-run/marker/uploader lifecycle.

## Acceptance Criteria

- [ ] Each child task has its own PRD, design, implementation plan, validation evidence, commit, archive entry, and journal update.
- [ ] The first child proves Kocoro-style workflow commands can enter through StarClaw daemon `POST /message` and stream over SSE without bypassing run metadata, approval, observability, routing, or control APIs.
- [ ] Phase6 final review updates the Kocoro parity gap assessment with remaining gaps and recommended Phase7 scope.
- [ ] No Phase6 child introduces external cloud transport, external telemetry, accounts, or off-machine sync without explicit approval.
