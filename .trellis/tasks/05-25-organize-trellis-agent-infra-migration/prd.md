# Organize Trellis agent infrastructure migration

## Goal

Review and organize the remaining Trellis/Claude/Codex/agent infrastructure migration changes so the repository can reach a clean, intentional state after product hardening.

## Requirements

- Inspect remaining uncommitted tracked and untracked infrastructure files.
- Check for obvious secrets, machine-local paths, or generated artifacts that should not be committed.
- Group safe Trellis/agent/Codex infrastructure changes into an appropriate commit.
- Keep unrelated build artifacts ignored and unstaged.
- Document any files intentionally left uncommitted.

## Acceptance Criteria

- [ ] Remaining infrastructure changes are classified.
- [ ] Safe changes are committed or explicitly documented as left uncommitted.
- [ ] No obvious secrets are introduced.
- [ ] Final git status is reported.

## Notes

- This task should not modify StarClaw product code.
