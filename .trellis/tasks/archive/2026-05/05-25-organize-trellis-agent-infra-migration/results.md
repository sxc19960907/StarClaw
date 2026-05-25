# Infrastructure Migration Results

Date: 2026-05-25

## Classified Changes

- Trellis core migration: `.trellis/workflow.md`, `.trellis/config.yaml`, `.trellis/.version`, `.trellis/.template-hashes.json`, and `.trellis/scripts/**`. Safe to commit as project infrastructure.
- Claude platform migration: `.claude/agents/**`, `.claude/commands/**`, `.claude/hooks/**`, `.claude/settings.json`, and `.claude/skills/trellis-*`. Safe to commit as generated platform wiring.
- Codex platform migration: `.codex/agents/**`, `.codex/hooks/**`, and `.codex/hooks.json`. Safe to commit as project-local Codex wiring.
- Project-local Trellis skills: `.agents/skills/trellis-*`. Safe to commit as reusable project agent workflow skills.
- Review/research artifacts: `BUG_REVIEW.md` and `.trellis/research/gap-analysis-starclaw-vs-shanclaw.md`. Safe to commit as project planning context.

## Left Uncommitted

- `.codex/config.toml`: local Codex config with a Feishu/Lark MCP app id and secret-like value. Added to `.gitignore` and intentionally not staged.
- `.agents/skills/obsidian-cli/`: unrelated personal/editor skill, not part of Trellis/Codex/Claude migration.

## Safety Checks

- Narrow secret scan found no high-confidence secrets in the selected commit set.
- The only high-confidence credential-looking values were in `.codex/config.toml`, which is excluded from the commit.
- Local path scan found historical ShanClaw reference paths in archived Trellis task notes; those were pre-existing project history, not new selected migration files.
