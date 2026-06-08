# Native tool parity followup

## Goal

Recompare StarClaw against the local Kocoro checkout after Phase 8 streaming/API work, then produce a concrete Phase 9 plan for native Desktop tools and local integration depth.

## Confirmed Facts

- Local Kocoro baseline is `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3c`.
- Phase 8 already closed OpenAI-compatible streaming deltas, streaming edge behavior, SSE reconnect/replay/idle watchdog behavior, and a local-first cloud lifecycle boundary.
- Kocoro's remaining deeper platform capability is concentrated in Desktop RPC calendar methods, calendar tools, browser handoff/leasing, Ghostty terminal helpers, image generation/editing, keychain, sync, and migration modules.
- StarClaw has a basic Desktop RPC transport, browser/computer/accessibility tools, local image processing, publish tools, and schedule tools, but not Kocoro-equivalent calendar, Ghostty, image generation/editing, keychain, sync, or migration packages.
- StarClaw must remain local-first by default. Real cloud credentials, external sync, off-machine delivery, and credential storage are not enabled in this planning task.

## Requirements

- Use the local Kocoro checkout as the comparison source; do not depend on live GitHub during routine parity planning.
- Produce a written evidence-based comparison for the remaining native/tool/Desktop gap.
- Separate local-safe implementation work from credentialed, cloud-backed, or off-machine behavior.
- Recommend the next parent phase and child task ordering.
- Preserve StarClaw naming in code/task artifacts and reserve Astria for product/UI language.
- Do not modify runtime code as part of this followup; this is a planning deliverable.

## Acceptance Criteria

- [ ] A research document records the Kocoro baseline, StarClaw baseline, evidence files, gap estimates, and Phase 9 recommendation.
- [ ] Phase 9 has a parent title, goal, and ordered child tasks with clear local-first boundaries.
- [ ] Credentialed/cloud/sync/keychain work is explicitly separated from local native-tool work.
- [ ] The parent Phase 8 PRD can point to this task as its final Kocoro comparison and Phase 9 plan.
- [ ] Trellis task validation passes.

## Out of Scope

- Implementing Phase 9 runtime code in this task.
- Enabling Shannon/Kocoro cloud credentials.
- Enabling real external sync, channel delivery, upload, or telemetry.
- Refreshing the local Kocoro checkout unless explicitly requested.
