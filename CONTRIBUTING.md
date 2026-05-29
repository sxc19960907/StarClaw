# Contributing to StarClaw

Thanks for your interest in contributing to StarClaw. This document covers the practical details you need to get started.

## Getting Started

### Prerequisites

- **Go 1.25+** (the project currently targets Go 1.25)
- Git

### Setup

```bash
git clone https://github.com/starclaw/starclaw.git
cd starclaw
go build -o starclaw .
```

This produces a `starclaw` binary in the project root. Verify it works:

```bash
./starclaw --help
```

> The `Makefile` offers convenience targets: `make build`, `make test`, `make test-race`, `make fmt`, `make lint`, `make coverage`.

## Project Structure

StarClaw is a single Go module with ~28,000 lines of code across 20 internal packages. The high-level layout:

```
cmd/            CLI entry points
internal/
  agent/        Agent execution loop (loop detection, retry, spill-to-disk)
  tools/        Built-in tools (file_read, bash, grep, think, http, etc.)
  daemon/       Background HTTP server, scheduler, session cache
  tui/          Bubble Tea terminal UI
  client/       LLM API client (Anthropic)
  session/      JSON-based session persistence
  config/       YAML configuration loading
  permissions/  4-layer security model
  mcp/          Model Context Protocol client
  skills/       Skill loader and registry
  hooks/        Lifecycle hooks (pre/post tool execution)
  audit/        JSON-lines audit logging
  prompt/       System prompt builder
  instructions/ Hierarchical instruction loading
  context/      Context window management and compression
  schedule/     Cron schedule CRUD
  heartbeat/    Heartbeat health check
  watcher/      fsnotify recursive file watcher
  update/       Self-update mechanism
pkg/            Shared packages
docs/           Documentation and assets
```

For a detailed walkthrough of each package and the data flow, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Running Tests

Run all tests:

```bash
go test ./...
```

Run with the race detector (recommended before submitting a PR):

```bash
go test -race ./...
```

Run a specific package:

```bash
go test ./internal/config/...
```

Generate and view coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # opens in browser
```

> Most packages use `testify` for assertions. Tests live alongside the code they test, following Go conventions.

## Commit Message Convention

We use conventional commit messages. Every commit message should follow this format:

```
<type>: <short description>

<optional body>
```

The `type` must be one of:

| Type       | When to use                                      |
|------------|--------------------------------------------------|
| `feat`     | A new feature                                    |
| `fix`      | A bug fix                                        |
| `docs`     | Documentation-only changes                        |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test`     | Adding or improving tests                        |
| `chore`    | Maintenance tasks (deps, build, CI config, etc.) |

Examples:

```
feat: add session_search tool for querying past sessions
fix: handle empty response from LLM API gracefully
docs: update README with MCP configuration example
refactor: extract retry logic into shared helper
test: add unit tests for permissions checker
chore: bump github.com/spf13/cobra to v1.10.2
```

Keep the subject line under 70 characters. Use the imperative mood ("add" not "added" / "adds").

## Pull Request Process

1. **Fork** the repository and create your branch from `main`.
2. **Branch name**: use a descriptive name, e.g. `feat/awesome-feature` or `fix/bug-description`.
3. **Keep PRs focused** on a single concern. A PR should do one thing and do it well.
4. **Run checks** before submitting:
   ```bash
   go test -race ./...
   go vet ./...
   go fmt ./...
   ```
5. **Write tests** for new functionality. If you're adding a tool, include tests for both happy and error paths.
6. **Update documentation** if your change affects the user-facing API, configuration, or CLI output.
7. **Open a pull request** against `main`. Provide a clear description of what the change does and why.
8. A maintainer will review your PR. Address any feedback with additional commits -- we squash-merge on approval.

> Do not bump version numbers or update the changelog. Maintainers handle releases.

## Code Style

StarClaw follows standard Go conventions:

- Format your code with `gofmt` (or `go fmt ./...`).
- Run `go vet ./...` and address any warnings before submitting.
- We recommend `golangci-lint` for deeper analysis (`make lint`).
- Keep functions focused and reasonably short.
- Follow the [Go Proverbs](https://go-proverbs.github.io/) and [Effective Go](https://go.dev/doc/effective_go) guidelines.
- Prefer interfaces for testability -- core components (tools, event handlers, LLM client) communicate through interfaces.
- When adding a new internal package, give it a short, lowercase name that describes its responsibility without stutter (e.g., `agent`, not `agentmanager`).
- Errors should be wrapped with context: `fmt.Errorf("reading config: %w", err)`.
- Use `require` and `assert` from `testify` in tests, not raw `t.Fatal`.

## Reporting Issues

- **Bug reports**: open a GitHub issue with reproduction steps, expected behavior, and actual behavior. Include your StarClaw version (`starclaw --version`) and OS.
- **Feature requests**: open a GitHub issue describing the use case and a proposed solution. The more concrete, the better.

## Questions?

Open a GitHub Discussion or reach out via the project's issue tracker.
