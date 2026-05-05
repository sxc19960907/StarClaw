# Directory Structure

> Go CLI project layout conventions for StarClaw

---

## Top-Level Layout

```
StarClaw/
├── main.go                 # Entry point — calls cmd.Execute(version)
├── cmd/                    # CLI commands (cobra)
├── internal/               # Application code (not importable externally)
├── tests/                  # Black-box integration tests
├── docs/                   # User-facing documentation
├── scripts/                # Build/CI helper scripts
├── npm/                    # npm distribution package
└── pkg/                    # (reserved for public library packages)
```

## Internal Package Layout

```
internal/
├── agent/                  # Agent loop + tool registry
├── agents/                 # Named agent loading (AGENT.md, config merge)
├── audit/                  # Audit logging (tool calls, decisions)
├── client/                 # LLM HTTP client + mock
├── config/                 # Configuration loading, setup wizard, merge
├── mcp/                    # MCP client (Model Context Protocol)
├── session/                # Session persistence (CRUD, title generation)
├── skills/                 # SKILL.md loading and parsing
├── tools/                  # Built-in tools (one file per tool)
├── tui/                    # Terminal UI (bubbletea)
└── update/                 # Self-update check + download
```

## Package Naming Rules

1. **Package name = directory name** — no abbreviations unless widely known (e.g., `mcp`, `tui`)
2. **One package per directory** — never have multiple packages in the same directory
3. **One concern per package** — don't mix unrelated functionality
4. **`internal/` for all app code** — nothing in `internal/` can be imported by other modules

## File Naming

- **Tool files**: Named after the tool (`file_read.go`, `bash.go`, `http.go`)
- **Test files**: `*_test.go` alongside the source (not in a separate test package)
- **Platform-specific**: Use Go build tags + suffixes: `system_info_darwin.go`, `system_info_linux.go`, `system_info_other.go`
- **Types & core**: `tools.go` for interfaces, `registry.go` for the registry

## Real Examples

| File | Purpose |
|------|---------|
| `internal/agent/tools.go` | `Tool`, `ToolInfo`, `ToolResult` interfaces & types |
| `internal/agent/registry.go` | `ToolRegistry` struct (Register, Get, List) |
| `internal/agent/loop.go` | `AgentLoop` — main conversation loop |
| `internal/tools/register.go` | `RegisterLocalTools()` — wires up all tools |
| `internal/config/config.go` | `Config` struct, `Load()`, `Save()` |
| `internal/config/setup.go` | Interactive first-run setup |
| `internal/config/merge.go` | Agent config overlay logic |

## Anti-Patterns

- ❌ Don't put CLI command logic in `main.go`
- ❌ Don't create deeply nested subdirectories (max 4 levels from `internal/`)
- ❌ Don't use `pkg/` for internal code — it signals public API intent
- ❌ Don't create `utils/` or `helpers/` packages — put functions near their usage
