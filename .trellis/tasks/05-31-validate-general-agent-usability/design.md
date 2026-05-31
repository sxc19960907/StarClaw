# Design

## Scope

This task adds and uses a lightweight CLI usability validation path. The goal is not to replace unit/integration tests, but to cover the user-facing binary paths that existing Go tests do not exercise directly.

## Validation Shape

- Build a temporary binary with `go build`.
- Run commands with an isolated `HOME` so the test never reads or writes the developer's real StarClaw state.
- Prefer deterministic commands that do not require API keys, model servers, TTY interaction, or desktop services.
- Treat missing configuration as a valid smoke target: it should fail clearly and without panic.
- Use existing Go tests for deeper agent loop and tool execution validation.

## Candidate Command Coverage

- `starclaw version`
- `starclaw --help`
- `starclaw chat "..."` with isolated empty config, expecting controlled configuration error
- `starclaw sessions` with isolated empty config or minimal config, depending on current CLI behavior
- `starclaw mcp` or `starclaw mcp --help`
- shell completion help or generation command if it is stable and non-interactive

## Artifact Location

Prefer a script under `scripts/` or `tests/` if the repository already uses a similar pattern. If no suitable location exists, add a focused shell script such as `scripts/smoke_cli.sh`.

## Compatibility

- The smoke script must run on macOS and Linux.
- It must avoid desktop/browser side effects.
- It must clean up temporary files automatically.

## Rollback

If the validation script introduces instability, it can be removed without changing runtime behavior. Runtime bug fixes should be committed separately enough to identify and revert.
