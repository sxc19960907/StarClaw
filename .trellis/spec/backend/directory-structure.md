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
├── desktop/                # Optional native desktop shells
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
  - `GET /diagnostics` -> returns daemon runtime readiness checks for the Web UI
  - `GET /config` -> returns redacted provider setup for the Web UI
  - `PATCH /config` -> updates provider-level YAML config fields

### 3. Contracts

- The UI must call daemon APIs with same-origin paths such as `/status`, `/diagnostics`, `/message`, `/agents`, `/skills`, `/sessions`, and `/schedules`.
- `GET /diagnostics` must be read-only and return:
  - `status`: one of `ready`, `warning`, `needs_setup`, or `error`.
  - `summary`: human-readable runtime summary.
  - `checks`: rows with `id`, `label`, `status`, `detail`, and optional `action`.
- Diagnostics must not make paid remote LLM calls. Static provider checks are enough for Anthropic/OpenAI; Ollama may use a short best-effort local reachability probe such as `GET /api/tags`.
- `GET /config` for the GUI must return provider-level scalar fields only and must never return plaintext API keys. Use booleans such as `api_key_set` and `openai_api_key_set`.
- `PATCH /config` must read/write `config.yaml` as YAML, preserve existing API keys when key fields are omitted or blank, and refresh `ServerDeps.Config` after a successful write so diagnostics updates without daemon restart.
- Provider config patching is intentionally scoped to first-run repair fields: `provider`, `endpoint`, `model_tier`, `api_key`, `openai_endpoint`, `openai_model`, `openai_api_key`, `ollama_endpoint`, and `ollama_model`.
- Do not add external network assets or a Node/Vite build step for this embedded GUI path unless the project explicitly adopts a frontend build pipeline.
- UI assets must be included by Go embed patterns so `go test ./internal/daemon` catches missing files.

### 4. Validation & Error Matrix

- Missing route -> route tests fail with default mux 404.
- Missing embedded asset -> `/app/assets/...` route test fails or `go test` fails at compile/embed time.
- Diagnostics provider credentials/model incomplete -> `status=needs_setup`, not HTTP 500.
- Diagnostics local storage/manager unavailable -> `status=error` with actionable check detail.
- Ollama endpoint unreachable -> `status=warning` so local offline Ollama does not block the whole GUI.
- Config GET over YAML -> structured JSON view with secret redaction.
- Config PATCH unsupported provider -> HTTP 400.
- Config PATCH blank API key -> keep the existing stored key.
- Config PATCH success -> persisted YAML is parseable and in-memory diagnostics config is refreshed.
- API failure in browser -> render an error state in the UI, not an uncaught exception.

### 5. Good/Base/Bad Cases

- Good: `/app/` renders a usable console, diagnostics state is visible from the topbar, and static assets are served by the daemon binary.
- Base: daemon API endpoints return empty lists or `needs_setup`; UI shows empty/action states.
- Bad: root path serves marketing content, references external assets, requires a separate frontend dev server, or diagnostics probes paid providers.

### 6. Tests Required

- Route tests for `/`, `/app`, `/app/`, and at least one CSS and JS asset under `/app/assets/`.
- Route/unit tests for `/diagnostics` covering `ready` and `needs_setup`.
- Backend tests for `/config` covering YAML parsing, secret redaction, blank-key preservation, provider validation, and in-memory config refresh.
- Smoke coverage that opens `/app/` and asserts diagnostics render in the browser.
- Smoke coverage for provider setup form render/save.
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
// Read-only local readiness API consumed by the embedded GUI.
mux.HandleFunc("GET /diagnostics", srv.handleDiagnostics)
```

## Scenario: macOS Astria Shell

### 1. Scope / Trigger

- Trigger: adding or changing the optional native macOS shell for Astria.
- Scope: thin app wrapper under `desktop/macos/Astria/` that hosts the existing
  daemon-served Web UI. The Go daemon and Web UI remain the runtime owners.

### 2. Signatures

- Swift source lives under `desktop/macos/Astria/Sources/`.
- Bundle metadata lives in `desktop/macos/Astria/Info.plist`.
- Local development build script: `scripts/build_macos_astria_shell.sh`.
- Smoke script: `scripts/smoke_macos_astria_shell.sh`.
- Default hosted URL: `http://127.0.0.1:7533/app/`.
- Route recovery storage key: `astria.lastWebRoute` stores a relative `/app`
  route only.
- Development override env keys:
  - `ASTRIA_WEB_URL`
  - `ASTRIA_DIAGNOSTICS_URL`
  - `ASTRIA_HEALTH_URL`
  - `ASTRIA_STARCLAW_BIN`

### 3. Contracts

- The shell must host the daemon Web UI; it must not duplicate the Web UI into
  a second frontend implementation.
- The shell must remain optional. `starclaw app`, `starclaw app --no-open`, and
  `starclaw daemon start` remain valid fallback paths.
- Unsigned local builds must not require private signing, notarization, or
  update credentials.
- The shell may start or attach to the local daemon through the existing HTTP
  readiness contract. Deeper pidfile/socket reconciliation belongs in a
  separately scoped task.
- The shell may persist the last useful Web UI route only when it is same-origin
  with the configured Web UI URL and under `/app`. Store relative route values,
  never full origins.
- The shell may reload the WebView after daemon health recovers, but the Web UI
  remains responsible for run/event deduplication through `/events` and `/runs`.
- App Transport Security may allow local networking for `127.0.0.1`; do not add
  broad remote networking exceptions for the shell.

### 4. Validation & Error Matrix

- Non-macOS host -> smoke script prints skipped and exits 0.
- Missing `swiftc` on macOS -> build script exits non-zero with a clear message.
- Missing bundle executable -> smoke script fails.
- Missing `Info.plist` -> smoke script fails.
- Missing local networking ATS allowance -> smoke script fails.
- Daemon unavailable at runtime -> shell attempts `starclaw daemon start`, then
  shows a user-visible failure state if health does not become ready.
- Unsafe stored route such as `https://example.com/app` or `/diagnostics` ->
  restore falls back to `/app/`.

### 5. Good/Base/Bad Cases

- Good: `scripts/smoke_macos_astria_shell.sh` builds an unsigned `Astria.app`,
  verifies bundle structure, route recovery, and daemon supervision through a
  temporary `ASTRIA_STARCLAW_BIN`.
- Base: user opens the unsigned app shell, which reuses an already-running
  daemon or starts a local daemon, then restores the last same-origin `/app`
  route.
- Bad: app shell requires cloud credentials, private signing keys, a separate
  frontend build pipeline, or silently replaces the CLI/browser launch path.

### 6. Tests Required

- Run `scripts/smoke_macos_astria_shell.sh` on macOS when the shell changes.
  The smoke must cover bundle structure, route recovery, daemon supervision,
  attach, and launch failure.
- Run `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1` to protect
  existing CLI launch readiness behavior.
- Run `go test ./...` before committing mixed shell/documentation changes.

### 7. Wrong vs Correct

#### Wrong

```swift
let url = URL(string: "https://remote.example/app")!
```

#### Correct

```swift
let url = URL(string: "http://127.0.0.1:7533/app/")!
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
