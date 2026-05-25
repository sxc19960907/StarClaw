# Repository State Baseline

Date: 2026-05-25
Branch: `main`

## Git State

- `main` is ahead of `origin/main` by 3 commits.
- Current Trellis task: `.trellis/tasks/05-25-stabilize-repo-state`.
- There are no uncommitted StarClaw product-code changes under `cmd/`, `internal/`, `tests/`, `docs/`, `npm/`, `main.go`, `go.mod`, or `go.sum`.

## Uncommitted Change Groups

### Trellis / Agent Infrastructure

Tracked modifications are concentrated in:

- `.claude/agents/`
- `.claude/commands/`
- `.claude/hooks/`
- `.claude/skills/`
- `.trellis/scripts/`
- `.trellis/workflow.md`
- `.trellis/config.yaml`
- `.trellis/.version`
- `.trellis/.template-hashes.json`
- `AGENTS.md`

Observed theme: migration from Trellis `0.5.0-rc.6` to `0.6.0-beta.18`, with updated workflow text, Codex inline dispatch configuration, task/session context scripts, and auto-commit / worker guard settings.

### New Platform Support Files

Untracked additions include:

- `.agents/skills/trellis-*`
- `.agents/skills/obsidian-cli/`
- `.codex/agents/`
- `.codex/hooks/`
- `.codex/config.toml`
- `.codex/hooks.json`
- `.trellis/scripts/common/safe_commit.py`
- `.trellis/scripts/common/trellis_config.py`

Observed theme: project-local Trellis skills plus Codex platform support files.

### Research / Review Artifacts

Untracked additions include:

- `.trellis/research/gap-analysis-starclaw-vs-shanclaw.md`
- `BUG_REVIEW.md`

These are directly useful for planning the bug-fix work and should be preserved.

### Active Task Artifacts

Untracked additions include this task directory:

- `.trellis/tasks/05-25-stabilize-repo-state/`

## Risk Notes

- No product code is currently dirty, so critical bug fixes can start without colliding with existing product edits.
- The Trellis/agent infrastructure update is broad. It should be reviewed and committed separately from StarClaw product-code fixes.
- Because local `main` is already ahead of `origin/main`, pushing or branching should be handled intentionally after deciding whether these local workflow updates belong in the repository.

## Recommendation

1. Preserve all current Trellis/agent changes.
2. Start critical bug-fix work in product code without modifying the existing infrastructure changes unless required by the workflow.
3. Keep bug-fix commits separate from Trellis/agent migration commits.
4. Use `BUG_REVIEW.md` as the source list for the next task, starting with the 5 severe findings.
