# Native tool parity followup design

## Boundary

This task is a planning and evidence task. It compares StarClaw against the local Kocoro checkout and writes the next parent phase plan. It does not change runtime behavior.

## Comparison Sources

- Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3c`.
- StarClaw repository: `/Users/timmy/PycharmProjects/StarClaw`.
- Existing Phase 8 plan: `.trellis/tasks/06-08-astria-kocoro-parity-phase-8/prd.md`.
- Prior comparison: `.trellis/research/kocoro-local-comparison-phase8-plan.md`.

## Output Contract

The main output is `.trellis/research/kocoro-native-tool-parity-phase9-plan.md`.

It must include:

- Kocoro baseline path and commit.
- Current StarClaw Phase 8 state.
- Evidence-backed comparison by capability area.
- Local-safe versus credentialed/cloud-sensitive split.
- Recommended Phase 9 parent and child tasks.
- Updated gap estimate.

## Capability Areas

The comparison groups Kocoro capability into:

- Desktop RPC and calendar protocol.
- Calendar native tools.
- Browser handoff and lease behavior.
- Terminal workspace / Ghostty helpers.
- Image generation and editing provider boundary.
- Keychain, sync, migration, upload/share foundations.

## Compatibility and Safety

- Treat Kocoro implementation as evidence, not as a direct copy target.
- Keep StarClaw package naming and local-first defaults.
- Do not add code paths that require real cloud credentials, OS keychain writes, or off-machine sync in this planning task.
- Any future credentialed/provider-backed child task must require explicit opt-in and approval-oriented tool behavior.

## Rollback

Rollback is limited to removing or revising the Trellis/research planning files introduced by this task. No runtime migration is needed.
