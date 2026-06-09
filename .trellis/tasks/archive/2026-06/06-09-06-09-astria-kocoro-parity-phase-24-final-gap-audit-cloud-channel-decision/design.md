# Phase 24 design

## Boundary

This task is a parity audit and product decision gate. It may write Trellis
planning/review documents and update static docs only if needed. It must not
change runtime behavior.

The main boundary is between:

- Astria local-first parity: daemon-served Web UI, local Desktop RPC, local
  runtime controls, replayable SSE, local release/update gates.
- Kocoro/Shannon production parity: Cloud WebSocket, channel OAuth/install
  lifecycle, IM message lifecycle, proactive delivery, remote uploads, team
  distribution, and account/server-side state.

## Evidence Sources

- Kocoro repository at `/Users/timmy/PycharmProjects/Kocoro`, commit
  `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.
- StarClaw repository docs and code under the current workspace.
- Archived Trellis final reviews for Phase16 through Phase23.
- Current route/API lists from both repositories.

## Output

Create `final-gap-audit.md` in the task directory with:

1. Baseline and scope.
2. Capability matrix.
3. Updated parity estimate.
4. Explicit product decision.
5. Recommended Phase25 child plan if cloud/channel work proceeds.
6. Validation performed.

## Trade-Offs

- A documentation-only decision gate is less flashy than shipping a feature,
  but it prevents a local-first product from accidentally pretending to support
  production cloud channels.
- Implementing real channel transport now would be premature without explicit
  user approval for auth, remote endpoints, provider credentials, and account
  surfaces.
- A future cloud/channel phase should begin with local contracts and simulation
  before real Shannon Cloud connectivity.

## Compatibility

No runtime compatibility changes are expected. Existing docs that say StarClaw
does not enable Shannon Cloud/IM lifecycle by default should remain true unless
the final decision explicitly changes the roadmap.
