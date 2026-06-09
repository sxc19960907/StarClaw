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
- Optional bundled daemon path inside the app:
  `Contents/Resources/starclaw`.
- Default hosted URL: `http://127.0.0.1:7533/app/`.
- Main native window scene id: `astria-main`.
- Route recovery storage key: `astria.lastWebRoute` stores a relative `/app`
  route only.
- Per-window route recovery uses `astria.window.<window-id>.lastWebRoute` for
  safe relative `/app` routes, with `astria.lastWebRoute` as the shared fallback
  for new windows.
- Development override env keys:
  - `ASTRIA_WEB_URL`
  - `ASTRIA_DIAGNOSTICS_URL`
  - `ASTRIA_HEALTH_URL`
  - `ASTRIA_STARCLAW_BIN`
  - `ASTRIA_RUNTIME_DIR`
  - `ASTRIA_BUNDLED_STARCLAW_BIN` for build-time daemon copying
  - `ASTRIA_APP_VERSION` for build-time `CFBundleShortVersionString`
  - `ASTRIA_APP_BUILD` for build-time `CFBundleVersion`

### 3. Contracts

- The shell must host the daemon Web UI; it must not duplicate the Web UI into
  a second frontend implementation.
- Native menu/window commands should be modeled in Swift and smoke-testable.
  The baseline command set is New Window, Reload Astria, Open Diagnostics, and
  Retry Daemon. Export Diagnostics is part of the diagnostics/crash-report
  boundary and writes a local redacted report.
- Export Crash Summary is part of the local crash-report boundary. It writes a
  local redacted crash summary under Astria diagnostics storage and must not
  upload reports automatically.
- Native clipboard/file affordance commands are local-only and user-triggered:
  Copy Current Route copies only a safe relative `/app` route, Copy Support
  Summary copies a redacted local support summary, and Reveal Diagnostics Folder
  opens the Astria-owned diagnostics export directory in Finder.
- Permission Help is a local helper boundary for future native desktop tools.
  It may show/copy status and guidance for Calendar, Contacts, Reminders, file
  access, and notifications, but it must not silently request broad macOS
  privacy/TCC permissions.
- Notification readiness is part of Permission Help. Passive readiness checks
  may read notification settings but must not request authorization or send a
  notification.
- New Window should open another `astria-main` window around the same local
  daemon instead of disabling macOS new-window behavior.
- Each native window may restore its own last safe `/app` route while sharing
  the same configured local daemon/runtime boundary. New windows without a
  window-specific route may fall back to the shared last safe route.
- Reload Astria should refresh the hosted WebView without changing the
  persisted route. Retry Daemon should call the existing daemon supervision
  start path.
- Export Diagnostics must generate a local-only report under Astria-owned app
  support/runtime storage. It must redact API keys, bearer tokens, raw Desktop
  RPC payloads, user prompts, socket paths, and pidfile paths.
- Export Crash Summary must generate a local-only crash summary from known
  local app/daemon failure state first. It must redact API keys, bearer tokens,
  raw Desktop RPC payloads, user prompts, socket paths, pidfile paths, crash
  file paths, and private local filesystem paths.
- Copy Support Summary must use the same redaction boundary as diagnostics
  export. It must not include API keys, bearer tokens, raw Desktop RPC payloads,
  user prompts, socket paths, pidfile paths, or full local filesystem paths.
- Copy Current Route must reuse the route safety boundary: same-origin with the
  configured Web UI URL and under `/app`, then stored/copied as a relative route
  value. Unsafe routes fall back to `/app/`.
- Permission helper guidance must be unavailable-safe in unsigned development
  builds. It may read non-prompting status APIs and explain how to use System
  Settings or future explicit request tools, but actual permission prompts must
  remain user-triggered by the specific native tool that needs access.
- Notification readiness must distinguish ready, blocked,
  requires-explicit-request, and unavailable-safe states. Any future
  notification permission prompt or delivery test must be an explicit
  user-triggered action, not part of passive status gathering.
- The shell must remain optional. `starclaw app`, `starclaw app --no-open`, and
  `starclaw daemon start` remain valid fallback paths.
- Unsigned local builds must not require private signing, notarization, or
  update credentials.
- Build-time bundled daemon copies must place the executable at
  `Contents/Resources/starclaw`. At runtime, explicit `ASTRIA_STARCLAW_BIN`
  overrides still take precedence, then the bundled daemon, then `PATH`.
- Runtime Desktop RPC artifacts default under
  `~/Library/Application Support/dev.starclaw.astria/` and may be isolated with
  `ASTRIA_RUNTIME_DIR` for smoke tests.
- The shell may delete only `<runtime-dir>/daemon.sock` and
  `<runtime-dir>/daemon.pid`. It must not delete parent directories, arbitrary
  paths, or files outside its configured runtime directory.
- Release-candidate app builds may set `ASTRIA_APP_VERSION` and
  `ASTRIA_APP_BUILD`; signed release artifacts must not mix app and daemon
  versions from different release tags.
- The shell may start or attach to the local daemon through the existing HTTP
  readiness contract. Deeper pidfile/socket reconciliation belongs in a
  separately scoped task.
- When the shell starts the daemon itself, it must pass paired `--rpc-socket`
  and `--rpc-pidfile` arguments and validate Desktop RPC `system.capabilities`
  before declaring the desktop handshake ready.
- After a successful Desktop RPC handshake, the shell must keep a long-lived
  app-side session monitor. The monitor may use `system.capabilities` probes
  until a richer event stream exists, but it must distinguish connected,
  reconnecting, degraded, and compatibility-mismatch states.
- Desktop RPC reconnect attempts must be bounded and cancellable. Restarting
  daemon supervision must cancel stale health/RPC monitor tasks before starting
  new ones.
- The native shell must surface Desktop RPC session diagnostics for
  reconnecting, degraded, and compatibility-mismatch states while keeping the
  WebView usable when HTTP health is available.
- If an already-running daemon is HTTP-healthy but lacks usable Desktop RPC,
  the shell may attach in degraded HTTP fallback mode while surfacing that state
  to the user.
- The shell may persist the last useful Web UI route only when it is same-origin
  with the configured Web UI URL and under `/app`. Store relative route values,
  never full origins.
- Window-specific route keys must use a conservative identifier whitelist and
  store only the same safe relative route values as the shared route key.
- The shell may reload the WebView after daemon health recovers, but the Web UI
  remains responsible for run/event deduplication through `/events` and `/runs`.
- App Transport Security may allow local networking for `127.0.0.1`; do not add
  broad remote networking exceptions for the shell.
- Signing and notarization require external Apple credentials, Hardened Runtime,
  notarization, and stapling. Do not commit credentials, signing identities,
  keychain profiles, notarization secrets, or update private keys.
- Astria does not auto-update itself in this phase. Missing updater metadata is
  unavailable-safe. Present updater metadata must be signed JSON with checksum,
  public-key identity, and app/daemon compatibility fields before any future app
  or bundled-daemon replacement behavior exists.
- Updater dry-run validation may parse verified metadata and emit a local
  decision containing version, checksum, signature algorithm, public key id, and
  app/daemon compatibility fields. The decision must keep replacement disabled.
- Release validation with `scripts/validate_release_artifacts.sh --npm-only
  --astria-local` must remain credential-free. It should run the Astria smoke,
  verify private signing/notarization material is absent, and reject Astria
  updater metadata that lacks checksum/signature/compatibility fields or enables
  app replacement before verified updater implementation exists.

### 4. Validation & Error Matrix

- Non-macOS host -> smoke script prints skipped and exits 0.
- Missing `swiftc` on macOS -> build script exits non-zero with a clear message.
- Missing bundle executable -> smoke script fails.
- Missing `Info.plist` -> smoke script fails.
- Missing bundled daemon in a bundled-app smoke -> smoke script fails.
- Missing local networking ATS allowance -> smoke script fails.
- Desktop RPC `system.capabilities` protocol mismatch -> shell blocks
  desktop-ready and surfaces a mismatch diagnostic.
- Desktop RPC missing `system.ping` or `system.capabilities` -> shell blocks
  desktop-ready and surfaces a missing capability diagnostic.
- Desktop RPC disconnect after initial handshake -> shell retries within its
  configured bound, then keeps the Web UI usable through degraded HTTP fallback
  if daemon health is still available.
- Desktop RPC protocol/version mismatch after initial handshake -> shell
  surfaces compatibility mismatch and does not spin retries against the same
  daemon.
- Desktop RPC diagnostic banners -> may describe local session state and retry
  guidance, but must not expose socket paths, pidfile paths, raw event payloads,
  API keys, provider headers, or user content.
- Clipboard support summaries -> local-only and redacted; they may include app
  version, safe route, daemon state labels, Desktop RPC severity, and diagnostics
  URL, but must not include secrets, raw user content, raw Desktop RPC payloads,
  socket paths, pidfile paths, or full local filesystem paths.
- Crash summaries -> local-only and redacted; they may include app version,
  daemon/Desktop RPC state labels, source labels, and aggregate crash counts,
  but must not include secrets, raw prompts, raw Desktop RPC payloads, crash file
  paths, socket paths, pidfile paths, full local filesystem paths, or upload
  destinations.
- Reveal Diagnostics Folder -> opens only the Astria diagnostics directory
  returned by the diagnostics exporter; it must not create remote shares or
  upload artifacts.
- Permission Help -> local-only guidance and status text; it must not call
  broad request-access APIs, add cloud auth, emit telemetry, expose file paths,
  or bypass daemon tool permissions/approval.
- Notification readiness -> local-only status text; it must not call
  notification authorization request APIs, send test notifications, add remote
  notification services, or bypass daemon tool permissions/approval.
- Dead or malformed pidfile under the Astria runtime directory -> shell removes
  scoped pidfile/socket artifacts before relaunch.
- Live pidfile under the Astria runtime directory -> shell does not remove
  pidfile or socket artifacts.
- Healthy HTTP daemon without usable Desktop RPC -> shell keeps the Web UI
  usable and shows degraded fallback.
- Unsigned development build -> allowed locally; do not present it as a signed
  distributable release artifact.
- Missing update metadata -> no replacement; current app remains usable.
- App/daemon release version mismatch -> reject the release-candidate artifact
  or rebuild from matching release inputs.
- Committed `.p8`, `.p12`, provisioning profile, private key, notary/keychain
  profile, or updater private material -> release validation fails.
- Astria updater metadata without checksum/signature/compatibility fields, with
  private fields, or enabling app replacement -> release validation fails;
  missing updater metadata remains non-fatal.
- Astria updater dry-run with valid metadata -> returns a verified dry-run
  decision with replacement disabled. Replacement-enabled metadata -> validation
  fails before any replacement path is considered.
- Daemon unavailable at runtime -> shell attempts `starclaw daemon start`, then
  shows a user-visible failure state if health does not become ready.
- Unsafe stored route such as `https://example.com/app` or `/diagnostics` ->
  restore falls back to `/app/`.
- Unsafe window-specific routes -> restore falls back to the shared safe route
  when available, otherwise `/app/`; unsafe full origins are never persisted.
- Native command smoke -> verifies command ids, labels, shortcut metadata, and
  shortcut conflict boundaries without requiring UI automation.
- Diagnostics export smoke -> verifies a local JSON report is written and does
  not contain API keys, bearer tokens, socket paths, or pidfile paths.
- Crash summary smoke -> verifies a local JSON summary is written and does not
  contain API keys, bearer tokens, raw prompts, socket paths, pidfile paths, or
  private local filesystem paths.
- Notification readiness smoke -> verifies readiness states and guidance text
  without requesting notification authorization or sending notifications.
- Updater boundary smoke -> verifies missing updater metadata is
  unavailable-safe, unsafe metadata fails, private updater fields fail, and
  signed JSON metadata with checksum/signature/compatibility fields passes while
  app replacement remains disabled.
- Updater dry-run smoke -> verifies missing metadata, valid metadata, and
  replacement-enabled metadata decisions without building release artifacts or
  requiring Apple credentials.

### 5. Good/Base/Bad Cases

- Good: `scripts/smoke_macos_astria_shell.sh` builds an unsigned `Astria.app`,
  verifies bundle structure, bundled daemon layout, route recovery, and daemon
  supervision through a temporary daemon binary. It also validates Desktop RPC
  capabilities, fallback cleanup, session lifecycle smoke paths, and native
  command/export/clipboard/file/permission helper metadata.
- Base: user opens the unsigned app shell, which reuses an already-running
  daemon or starts a local daemon, then restores the last same-origin `/app`
  route.
- Bad: app shell requires cloud credentials, private signing keys, a separate
  frontend build pipeline, or silently replaces the CLI/browser launch path.

### 6. Tests Required

- Run `scripts/smoke_macos_astria_shell.sh` on macOS when the shell changes.
  The smoke must cover bundle structure, route recovery, daemon supervision,
  attach, bundled daemon resolution, Desktop RPC capabilities reconciliation,
  stale Desktop RPC artifact recovery, unsafe cleanup refusal, Desktop RPC
  session connected/retry/mismatch diagnostics, native command metadata,
  diagnostics export redaction, crash summary redaction, clipboard/file
  affordance route safety and support summary redaction, permission helper
  guidance/unavailable-safe behavior, notification readiness guidance,
  multi-window route isolation/fallback, and launch failure.
- Run `scripts/validate_release_artifacts.sh --npm-only --astria-local` when
  touching Astria distribution boundaries. It must pass without Apple
  credentials.
- Run `scripts/validate_release_artifacts.sh --npm-only --astria-local` on macOS
  when touching release validation boundaries.
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
