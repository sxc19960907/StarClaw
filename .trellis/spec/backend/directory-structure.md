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

## Scenario: Daemon-hosted Web UI

### 1. Scope / Trigger

- Trigger: adding or changing the local browser UI served by the daemon.
- Scope: static GUI assets embedded in the Go binary and same-origin JSON calls to existing daemon endpoints.

### 2. Signatures

- Route registration belongs in `internal/daemon/router.go`.
- Embedded asset handlers belong in `internal/daemon/webui.go`.
- Static files live under `internal/daemon/webui/`.
- Required routes:
  - `GET /` -> redirects to `/app/`
  - `GET /app` -> redirects to `/app/`
  - `GET /app/` -> serves `internal/daemon/webui/index.html`
  - `GET /app/assets/{file}` -> serves embedded files from `internal/daemon/webui/assets/`

### 3. Contracts

- The UI must call daemon APIs with same-origin paths such as `/status`, `/message`, `/agents`, `/skills`, `/sessions`, and `/schedules`.
- Do not add external network assets or a Node/Vite build step for this embedded GUI path unless the project explicitly adopts a frontend build pipeline.
- UI assets must be included by Go embed patterns so `go test ./internal/daemon` catches missing files.

### 4. Validation & Error Matrix

- Missing route -> route tests fail with default mux 404.
- Missing embedded asset -> `/app/assets/...` route test fails or `go test` fails at compile/embed time.
- API failure in browser -> render an error state in the UI, not an uncaught exception.

### 5. Good/Base/Bad Cases

- Good: `/app/` renders a usable console and static assets are served by the daemon binary.
- Base: daemon API endpoints return empty lists; UI shows empty states.
- Bad: root path serves marketing content, references external assets, or requires a separate frontend dev server.

### 6. Tests Required

- Route tests for `/`, `/app`, `/app/`, and at least one CSS and JS asset under `/app/assets/`.
- Targeted validation with `go test ./internal/daemon ./cmd`.
- Full validation with `go test ./...` and `go vet ./...`.

### 7. Wrong vs Correct

#### Wrong

```go
mux.HandleFunc("GET /app/", proxyToFrontendDevServer)
```

#### Correct

```go
mux.HandleFunc("GET /app/", srv.handleWebApp)
mux.HandleFunc("GET /app/assets/", srv.handleWebAsset)
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
