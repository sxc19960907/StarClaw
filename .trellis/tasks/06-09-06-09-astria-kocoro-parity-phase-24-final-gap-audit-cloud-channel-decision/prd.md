# Astria Kocoro parity phase 24: final gap audit and cloud channel decision

## Goal

Produce a final evidence-backed Kocoro parity audit after Phase23 and decide
the next track: keep Astria as a local-first product with explicit cloud
boundaries, or open a future cloud/channel parity foundation phase.

This phase must not implement Shannon Cloud transport. Its value is to prevent
scope drift by making the remaining gaps, risks, and product decision explicit.

## Confirmed Facts

- Current StarClaw working tree was clean when this task started.
- Kocoro local baseline is available at `/Users/timmy/PycharmProjects/Kocoro`
  and currently resolves to commit
  `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.
- Phase16-22 closed most local-first desktop/release/updater gaps, ending with
  a CLI/npm-only updater decision that keeps real app replacement disabled.
- Phase23 added user-visible streaming/long-run status in Astria Chat and kept
  `/message`, `/events`, and OpenAI-compatible streaming local-first.
- Kocoro's production path is daemon + Kocoro Desktop + Shannon Cloud:
  `cmd/daemon.go` builds a WebSocket endpoint ending in `/v1/ws/messages`,
  receives cloud/channel messages, injects follow-ups, tracks IM lifecycle,
  and emits `MESSAGE_LIFECYCLE` events.
- Astria/StarClaw has local planning surfaces for cloud/channel work:
  `/cloud/lifecycle`, `/queue`, `/channel/adapters`, `/channel/routes/{id}`,
  `/channel/state`, and local inbox/webhook APIs.
- Astria's current cloud lifecycle controller explicitly states:
  "Cloud WebSocket lifecycle boundary is local-only; no external transport is
  active."

## Requirements

- Compare Astria against Kocoro across the current high-value parity dimensions:
  runtime/streaming, daemon events, workflow control, desktop/native,
  release/updater, session/sync/share, cloud/channel/IM lifecycle, and
  team/distribution surfaces.
- Base the audit on local repository evidence, not memory.
- Produce a concise gap matrix with:
  - Astria status,
  - Kocoro baseline evidence,
  - remaining gap,
  - recommended next action.
- Update the parity estimate after Phase23.
- Decide whether the next phase should:
  - keep Astria local-first and declare cloud/channel out of scope for now, or
  - open a scoped cloud/channel foundation phase.
- Keep the decision credential-free and local-safe:
  - no real Shannon Cloud auth,
  - no external WebSocket connection,
  - no IM provider OAuth,
  - no remote uploads,
  - no account/team model implementation.
- If a future cloud/channel phase is recommended, define its first safe slice
  as contract and simulation work before any production transport.

## Acceptance Criteria

- [ ] `final-gap-audit.md` records the evidence-backed parity matrix.
- [ ] The audit includes a clear updated Kocoro parity estimate.
- [ ] The audit names which gaps are product decisions rather than missing
      local platform plumbing.
- [ ] The audit recommends the next phase and first child slices.
- [ ] No production cloud/channel transport or external credentials are added.
- [ ] Existing validation that is relevant to this documentation-only phase
      passes: task context validation, markdown/static checks if available,
      and `git diff --check`.

## Out of Scope

- Implementing Shannon Cloud WebSocket client behavior in StarClaw.
- Adding real Slack, Feishu/Lark, Telegram, LINE, WeCom, or webhook outbound
  delivery.
- Adding cloud auth, team accounts, billing, remote sync, or public upload
  execution.
- Changing runtime APIs, Web UI behavior, or release updater mechanics.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
