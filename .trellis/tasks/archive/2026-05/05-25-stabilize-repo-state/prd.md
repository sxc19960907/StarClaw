# Stabilize repo state before bug fixes

## Goal

Establish a clean, well-understood repository baseline before starting the critical bug-fix work.

The project currently has no active Trellis task, the local `main` branch is ahead of `origin/main`, and the working tree contains many uncommitted changes concentrated in Trellis/agent workflow files. Before changing StarClaw product code, classify the existing changes so future bug-fix commits start from an intentional state.

## Requirements

- Inspect git status and categorize all current uncommitted changes.
- Identify whether uncommitted changes are product-code changes, Trellis/agent infrastructure changes, generated artifacts, or local-only files.
- Preserve all existing user/generated work; do not revert unrelated changes without explicit user instruction.
- Produce a concise repository-state summary that can guide the next development task.
- Leave the project ready for the next task: critical bug fixes from `BUG_REVIEW.md`.

## Acceptance Criteria

- [ ] Current branch, ahead/behind state, and active Trellis task status are known.
- [ ] Uncommitted changes are grouped by purpose and risk.
- [ ] Any product-code changes requiring attention are identified.
- [ ] No unrelated user changes are reverted or overwritten.
- [ ] A clear recommendation exists for the next step: commit, ignore, or follow up on each change group.

## Notes

- This is a lightweight task; PRD-only planning is sufficient.
- This task should not fix bugs yet. It prepares the repository for the bug-fix task.
