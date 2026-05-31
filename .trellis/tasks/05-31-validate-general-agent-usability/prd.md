# Validate General Agent Usability

## Goal

Confirm that StarClaw is usable as a general-purpose CLI agent after CI recovery, and fix any blocking issues found during local end-to-end validation.

## User Value

A user should be able to install/build StarClaw, inspect the CLI, run common non-network commands, use isolated configuration, and trust that the advertised basic agent surfaces work before deeper feature development continues.

## Confirmed Facts

- `main` CI is green after run `26670622748`.
- The repository has mock/integration coverage for config loading, tool registration, agent loop tool execution, sessions, MCP, daemon, and many individual tools.
- Existing tests are mostly Go-level tests; they do not fully validate the built CLI binary as a user would run it.
- The CLI supports at least `version`, `setup`, `chat`, `interactive`, `sessions`, `mcp`, `update`, and shell completion commands.
- StarClaw config is read from user/project locations and environment variables; usability validation must not depend on the developer's real `~/.starclaw`.
- The worktree contains an unrelated untracked directory `.agents/skills/obsidian-cli/`; this task must not modify it.

## Requirements

- Build the StarClaw binary from the current workspace.
- Run CLI smoke checks against an isolated temporary `HOME` or equivalent config root.
- Validate help/version surfaces without needing real API credentials.
- Validate that missing configuration produces a controlled, understandable failure rather than a panic.
- Validate common local agent/tool plumbing using existing mock-level tests and, where practical, binary-level command execution.
- Identify blockers for using StarClaw as a general-purpose agent and fix issues that are narrow enough to address in this task.
- Leave external provider calls out of scope unless they can be run without secrets or network dependence.
- Preserve current CI behavior and keep `go test ./...`, `go vet ./...`, and GitHub Actions passing.

## Acceptance Criteria

- [x] A repeatable validation command or script exists for local CLI usability smoke checks.
- [x] The validation covers binary build, `version`, `--help`, missing-config behavior, session command behavior, and at least one isolated config path scenario.
- [x] Any blocking bugs found by the validation are fixed or explicitly documented if out of scope.
- [x] `go test ./...` passes locally.
- [x] `go vet ./...` passes locally.
- [x] CI remains green after the changes are pushed.
- [x] Results are summarized so the next development priority is clear.

## Validation Results

- Added `scripts/smoke_cli.sh`.
- Local smoke validation passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- No blocking runtime bug was found in the scoped non-network CLI paths.
- Real provider conversations, TUI manual behavior, and desktop/browser actions remain outside this smoke scope.

## Out Of Scope

- Real Anthropic/OpenAI/Ollama conversation tests requiring credentials or a running model server.
- Full TUI manual testing.
- GUI work.
- Large redesigns of config, tools, or session architecture.
- Touching unrelated untracked skill directories.

## Open Questions

- None blocking. Default scope is a pragmatic local smoke validation plus fixes for narrow blockers discovered during that validation.
