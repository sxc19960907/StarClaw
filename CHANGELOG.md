# Changelog

All notable changes to StarClaw are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/).

## v0.2.1 — 2026-05-29 — release hardening

### Fixed

- **Critical bug-review fixes** — hardened `MockClient` concurrent test use, locked in streaming behavior so successful streams do not issue duplicate non-streaming calls, preserved content when sanitizing consecutive same-role messages, covered context-window turn counting, and strengthened checkpoint ID sanitization.
- **Path and publish security** — replaced string-prefix path containment with symlink-aware relative containment, rejected sibling-prefix and symlink escapes, validated `grep` and `publish_to_web` paths before access, and validated screenshot output paths before platform execution.
- **Concurrency and lifecycle** — made process stdout/stderr capture safe while status readers poll running processes, guarded heartbeat Start/Close lifecycle state, added concurrent registry/read-tracker regression tests, widened watchdog reset timing coverage for race-mode stability, and guarded watcher debounce timer callbacks against stale generations.
- **Cancellation and timers** — stopped wait-tool, SSE reconnect, approval timeout, LLM retry, and scheduler alignment timers promptly on cancellation so cancelled runs do not leave delayed operations behind.

### Changed

- **Release readiness** — recorded passing full tests, targeted race tests, local/cross-platform builds, daemon lifecycle smoke, schedule create/remove smoke, shell completion smoke, and MCP serve help smoke.
- **Update UX** — clarified that StarClaw currently supports update checks and notifications; automatic binary replacement remains a documented follow-up.

## v0.2.0 — 2026-05-10 — Phase 4+5: full ShanClaw parity

### Added

- **Agent** — normalize (URL/query规范化), toolbudget (字符预算), skill_discovery, phase tracker, usage tracker (token+成本), statecache, watchdog timer, warmset, resultshape (响应形状分析), cachemetric (缓存P50/P95/P99), testing_helpers
- **Daemon** — attachment (文件附件), session_cwd, readtracker_cache, rules (agent规则), safeguard (危险命令拦截: rm -rf/, fork bomb等), marketplace, project_init, checkpoint (会话恢复)
- **Tools** — filepreview (文件预览), memory (记忆搜索/删除), readonly (只读模式), skill (动态加载), imaging (describe/resize/convert/OCR), mcp_error_hints
- **macOS tools** — accessibility (GUI元素查询), computer (鼠标/键盘), browser (打开URL/获取标题)
- **Ollama provider** — 本地模型支持
- **MCP** — health supervisor, readiness checker
- **New packages** — cwdctx, uploads, runstatus
- **Config** — settings.json (Spinner/MaxResponseLines/ShowTips), ollama配置

## v0.1.0 — 2026-05-09 — OpenAI support, session upgrade, shell integration

### Added

- **OpenAI provider** — configurable LLM provider alongside Anthropic. Set `provider: openai` in config with `openai_api_key`; maps tools, thinking, and streaming to OpenAI's chat completions API. Provider selection per-agent via `model_provider` in agent config.
- **Session tags and favorites** — tag sessions with `starclaw session tag <id> <key=value>` and search by tag. Favorite sessions (`starclaw session favorite <id>`) pin them to the top of listings. Configurable per-agent default tags.
- **Session Markdown/HTML export** — `starclaw session export <id> --format markdown|html` renders full conversation history. HTML output includes syntax-highlighted code blocks and responsive layout. Supports `--output` flag for file destination.
- **Shell completion** — `starclaw completion bash|zsh|fish|powershell` generates shell completion scripts. Auto-install option writes to the system completion directory.
- **Pipe mode CWD context** — piped input (`cat main.go | starclaw chat "explain"`) uses the current working directory for file operations, matching interactive session behavior. CWD propagates through one-shot and resume paths.

### Changed

- **Session list CLI** — `starclaw sessions` now supports `--tag` filter, `--favorites` flag, and `--export` flag for direct export output.
- **Config file discovery** — project-local `.starclaw/config.local.yaml` merged at runtime, scoped to session-safe fields only.

### Fixed

- **Data race in SessionCache.LockRoute** — concurrent cancel logic no longer races on route entry access.

## v0.0.2 — 2026-05-08 — Thinking mode, TUI, bundled skills, CLI commands

### Added

- **Thinking mode** — extended thinking via Anthropic's `thinking` parameter, configurable via `agent.claude_thinking` and `agent.thinking_budget` (default 16K tokens). Streamed progressively in TUI and one-shot modes.
- **Streaming support** — token-by-token streaming in interactive TUI and one-shot modes. Tool calls and text streamed as they arrive from the LLM.
- **Config extensions** — `agent.max_thinking_tokens`, `agent.thinking_budget`, `agent.temperature`, `agent.top_k`, `agent.top_p` per-agent overrides. MCP server `disabled` flag for toggling without removal.
- **TUI** — `starclaw interactive` with Bubble Tea. Markdown rendering, startup animation, tool call formatting with elapsed time, session-aware layout. `/research` and `/swarm` slash commands supported.
- **CLI** — `starclaw daemon start|stop|restart|status` for daemon lifecycle management. `starclaw schedule add|list|remove` for cron task management.
- **Bundled skills** (14) — upstream skills embedded at build time: `kocoro` (identity/language), `kocoro-guide` (platform config), `kocoro-generative-ui` (inline HTML artifacts), `explain-code`, `heatmap-analyze`, `go-refactoring`, `python-expert`, `shell-master`, and more.
- **OnPreamble event** — agents announce "about to run X" between tool calls. Emitted via `agent_text` SSE event; rendered in dim style in TUI. New `OnPreamble(text)` callback separated from `OnText`.
- **Prompt enhancement** — system prompt rebalanced for brief narration; contrast examples for common failure modes (over-engineering, coding-default bias, premature completion, wrong cloud/local boundary).
- **Grep VCS/glob support** — `grep` tool now skips VCS metadata (`.git`, etc.) automatically. `glob` filter list, `head_limit`, `before_context`/`after_context`, `sort_by mtime`.
- **Context bloat detection** — `OnRunStatus("tool_result_bloat", ...)` surfaces when tool output exceeds per-turn threshold. SSE and Desktop subscribers notified.
- **Bash output cap** — configurable `bash_max_output` (default 30K chars). Caller-controlled via `max_output_chars` per invocation. Truncated output head+tail with indicator.
- **`publish_to_web` tool** — uploads files to cloud endpoint returning a permanent public URL. Multi-layer guards: path blocklist, extension allowlist, 50 MiB pre-check, 3-attempt retry. Gated on `cloud.enabled` and `api_key`.
- **Watcher module** — filesystem watcher for config reload and session directory monitoring.
- **Heartbeat module** — periodic heartbeat with daemon health reporting, configurable interval and timeout.

### Changed

- **Config structure** — `agent.max_iterations`, `agent.max_tokens` moved under `agent.*` namespace. `claude_thinking` group added for thinking mode configuration.
- **TUI startup** — `starclaw` without subcommand defaults to interactive mode. `--one-shot` / `-1` flag for one-shot from TUI.

### Fixed

- **CI Go version alignment** — updated to Go 1.24 across all CI jobs to match `go.mod` toolchain directive.

## v0.0.1 — 2026-05-06 — Initial release: agent loop, daemon, permissions, MCP

### Added

- **Agent loop** — full agent execution loop with tool calling, message history management, turn state tracking. Loop detection prevents infinite tool-call cycles. Auto-retry with exponential backoff for transient API errors. Large tool results (>50KB) spill to disk. Read tracker deduplicates repeat file reads within a session.
- **MCP client** — Model Context Protocol client with stdio transport. Server lifecycle management (spawn, health check, reconnect). Tool schema discovery and invocation via `mcp_tool`. Configurable via `mcp_servers` in config.
- **Tools (12 built-in)** — `file_read`, `file_write`, `file_edit`, `glob`, `directory_list`, `grep`, `think`, `system_info`, `http`, `bash`, `mcp_tool`, `use_skill`. Approval-gated where appropriate.
- **Permissions** — 4-layer security model: path validation under session CWD, per-tool approval dialogs, configurable allowed/denied tool lists, always-ask gate for high-risk shell commands.
- **Daemon module** — HTTP server with 23 API endpoints: session CRUD, agent management, skill management, config management, message injection, instructions management. SSE event stream for real-time updates. WebSocket support for approval requests.
- **Context compaction** — tiered compaction with head+tail truncation, reactive compaction on overflow, LLM-summarized micro-compact for long sessions. Approval cache module for cross-turn permission deduplication.
- **Session persistence** — JSON session files with message history, tool results, metadata. Autosave after each turn. Resume via `--resume <id>`. Session list/search with `starclaw sessions`.
- **Schedule/Cron** — flock-protected cron task manager. Task scheduling with cron expressions. Daemon-affiliated execution with configurable timeouts.
- **Memory system** — persistent agent memory via `MEMORY.md` files. `memory_append` tool for updating memory. Memory recalled at session start.
- **Hooks system** — event hooks for tool execution lifecycle: `OnToolCall`, `OnToolResult`, `OnRunStatus`, `OnMessage`, `OnSessionClose`. Fan-out to multiple handlers.
- **Prompt builder** — structured prompt assembly with system prompt, tool schemas, conversation history, instructions. Hierarchical instruction injection from agent config, skills, and project config.
- **Utility tools** — `session_search` (FTS5 full-text search across sessions), `version` (version info), `wait` (pause execution for specified duration).
- **CLI** — Cobra-based CLI with subcommands: `chat` (one-shot), `interactive` (TUI), `sessions`, `schedule`, `daemon`, `mcp`, `update`. `-y`/`--yes` for auto-approval mode.
- **Audit logging** — append-only JSON-lines audit log with automatic secret redaction (API keys, tokens, credentials).
- **Self-update** — automatic update checking on startup (configurable via `update.auto_check`). Manual update via `starclaw update`.
- **Project-level config** — `.starclaw/config.local.yaml` per-project overlay for session-safe settings. Config merge with process-global config.
