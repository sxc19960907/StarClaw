# Gap Analysis: StarClaw vs ShanClaw

- **Query**: Thorough gap analysis comparing StarClaw against ShanClaw
- **Scope**: mixed
- **Date**: 2026-05-08

## Findings

### Overview

Both projects share identical `internal/` directory structure (20 packages each). No packages are exclusive to either project. The gap is entirely in scope and depth: ShanClaw has ~50,800 Go lines vs StarClaw's ~23,506 (2.16x larger). Tools, agent, and MCP packages account for the bulk of the delta.

---

### 1. Tools Gap (CRITICAL)

StarClaw is missing entire tool categories that ShanClaw implements:

| Tool | Files | Lines | Severity | Purpose |
|---|---|---|---|---|
| `browser.go` + test + pinchtab | 4 | ~56,000 | CRITICAL | Full browser automation (Playwright-based) |
| `accessibility.go` + test | 2 | ~24,000 | CRITICAL | macOS Accessibility API for GUI interaction |
| `screenshot.go` + test | 2 | ~7,200 | CRITICAL | Screen capture for visual context |
| `computer.go` + test | 2 | ~15,400 | CRITICAL | Mouse/keyboard control |
| `axclient.go` | 1 | ~9,300 | CRITICAL | macOS AX client for GUI automation |
| `axserver/` (Swift pkg) | 10 Swift files | ~2,500 | CRITICAL | Native macOS accessibility server |
| `clipboard.go` + test | 2 | ~4,800 | MEDIUM | Clipboard read/write |
| `cloud_delegate.go` + test | 2 | ~18,400 | MEDIUM | Offload tasks to cloud agents |
| `ghostty.go` + darwin + stub + test | 4 | ~18,600 | MEDIUM | Ghostty terminal integration |
| `imaging.go` + test | 2 | ~11,200 | MEDIUM | Image processing (OCR, analysis) |
| `process.go` + test | 2 | ~5,900 | MEDIUM | Process management (start/stop/list) |
| `readonly.go` + test | 2 | ~11,800 | MEDIUM | Enforce read-only file mode |
| `server.go` + test | 2 | ~6,700 | MEDIUM | Expose tools via MCP stdio server |
| `skill.go` + test | 2 | ~5,800 | MEDIUM | Dynamic skill loading/execution |
| `notify.go` + test | 2 | ~3,800 | LOW | Desktop notifications |
| `pinchtab.go` | 1 | ~11,300 | LOW | Pinch tab gesture detection |
| `applescript.go` + test | 2 | ~4,900 | MEDIUM | AppleScript execution |

StarClaw has `use_skill.go`, `system_info_linux.go`, `version.go`, `wait.go` with tests that ShanClaw lacks or has without tests.

### 2. Agent Loop Features (CRITICAL)

| Feature | ShanClaw | StarClaw | Severity |
|---|---|---|---|
| Streaming (OnStreamDelta) | Yes | No | CRITICAL |
| Thinking/Reasoning config | Yes | No | CRITICAL |
| Specific model override | Yes | No | MEDIUM |
| Web query normalization | `normalize.go` | No | MEDIUM |
| Retry logic | `retry_test.go` | No | MEDIUM |
| Cloud agent delegation events | Yes | No | MEDIUM |
| Skills integration in loop | Yes | No | MEDIUM |
| Context retry/partition concurrency | Yes | No | LOW |
| Total agent lines | 9,143 | 3,594 | - |

### 3. MCP Infrastructure (CRITICAL)

| Feature | ShanClaw | StarClaw | Severity |
|---|---|---|---|
| MCP server (stdio) | Yes (12 files, 5000 lines) | No (only client, 2 files, 1061 lines) | CRITICAL |
| Chrome CDP integration | Yes | No | MEDIUM |
| Playwright probe | Yes | No | MEDIUM |
| Health supervisor | Yes | No | MEDIUM |
| Readiness checks | Yes | No | MEDIUM |

### 4. Config Fields Missing in StarClaw

| Field | ShanClaw Example | Severity |
|---|---|---|
| `agent.thinking` / `agent.thinking_mode` / `agent.thinking_budget` | `true` / `"adaptive"` / `10000` | CRITICAL |
| `agent.reasoning_effort` | `""` | MEDIUM |
| `agent.model` | `""` (specific model override) | MEDIUM |
| `tools.server_tool_timeout` | `5` (seconds) | MEDIUM |
| `tools.grep_max_results` | `100` | MEDIUM |
| `cloud.enabled` / `cloud.timeout` | `true` / `3600` | MEDIUM |
| `daemon.auto_approve` | `false` | MEDIUM |
| Config source tracking | Full `ConfigSource` system | LOW |
| Settings file (`settings.json`) | Spinner/UX customization | LOW |
| Multi-level config merge (global/project/local) | Pointer-based overlay merge | MEDIUM |

### 5. CLI Commands Missing in StarClaw

| Command | Subcommands | Severity |
|---|---|---|
| `daemon` | `start`, `stop`, `status` | CRITICAL |
| `mcp serve` | (none) | MEDIUM |
| `schedule` | `list`, `create`, `update`, `remove`, `enable`, `disable` | MEDIUM |
| `ghostty workspace` | (none) | LOW |

StarClaw has `version`, `setup`, `chat`, `interactive`, `sessions`, `mcp list`, `update` which are either differently named or absent in ShanClaw (ShanClaw uses positional args for chat, `--version` flag, `--setup` flag, and TUI default mode).

### 6. Other Notable Gaps

| Package | StarClaw | ShanClaw | Delta |
|---|---|---|---|
| `daemon/` | 17 files, 4,785 lines | 25 files, 7,844 lines | +64% (launchd, router, permissions) |
| `tui/` | 2 files, 681 lines | 6 files, 2,889 lines | +324% (markdown rendering, frog animation) |
| `session/` | 7 files, 807 lines | 10 files, 2,191 lines | +171% (index, smoke tests, scenarios) |
| `context/` | 8 files, 1,431 lines | 10 files, 2,777 lines | +94% (with consolidation logic) |
| `client/` | 3 files, 549 lines | 4 files, 1,281 lines | +133% (SSE streaming, gateway client) |

### Summary of Critical Actionable Gaps

1. **Browser + GUI automation** (accessibility, computer, screenshot, axserver) -- whole category missing
2. **Agent streaming + thinking** -- core loop lacks streaming delta, thinking/reasoning budget config
3. **Daemon mode** -- background daemon with WS client, approval broker, launchd management
4. **MCP server** -- can only connect as client, cannot serve tools to MCP consumers
5. **Skill system** -- no dynamic skill loading/execution from config

StarClaw has a functional but minimal CLI agent; ShanClaw has matured into a full platform with cloud integration, GUI automation, and background service capabilities.
