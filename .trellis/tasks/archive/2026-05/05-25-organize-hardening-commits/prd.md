# Organize commits for hardening work

## Goal

Create clean commits for the completed StarClaw hardening, release-readiness, and build-smoke work while leaving unrelated Trellis/agent infrastructure migration changes unstaged.

## Requirements

- Group product code, tests, release docs, build metadata, and backend spec updates into a coherent commit.
- Exclude pre-existing Trellis/agent/Codex infrastructure migration files from the product hardening commit.
- Verify staged files before committing.
- Leave a clear git status after the commit.

## Acceptance Criteria

- [ ] Product hardening files are committed.
- [ ] Pre-existing `.claude/**`, `.agents/**`, `.codex/**`, broad `.trellis/scripts/**`, `.trellis/workflow.md`, and AGENTS migration changes remain unstaged.
- [ ] Commit message clearly describes the hardening work.
- [ ] Final git status is reported.

## Notes

- This task is about git organization, not code changes.
