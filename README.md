# StarClaw

[![CI](https://github.com/starclaw/starclaw/actions/workflows/ci.yml/badge.svg)](https://github.com/starclaw/starclaw/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/starclaw/starclaw)](https://goreportcard.com/report/github.com/starclaw/starclaw)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**StarClaw** is an AI-powered CLI agent with local tool execution capabilities. It provides a terminal interface for interacting with Large Language Models (LLMs) while enabling the AI to safely execute commands on your local system through a controlled tool system.

![StarClaw Demo](docs/demo.gif)

## Features

- 🤖 **Multi-Model AI** — Anthropic Claude + OpenAI GPT-4o + Ollama (local)
- 🛠️ **33 Built-in Tools** — Files, shell, search, desktop automation, macOS, imaging, scheduling, MCP
- 🖥️ **macOS Desktop Control** — Accessibility, mouse/keyboard, browser, screenshots, clipboard, notifications
- 🔌 **MCP Client + Server** — Connect to MCP servers, or expose tools via stdio
- 👤 **Named Agents** — Per-agent config, memory, custom prompts, heartbeat monitoring
- 🧩 **Skills System** — 14 bundled skills + dynamic skill discovery + loading
- ⏰ **Scheduled Tasks** — Cron-based local task scheduling with flock-protected persistence
- 🔄 **Background Daemon + Web UI** — HTTP API, embedded `/app/` GUI, scheduler, heartbeat, file watcher
- 💻 **Beautiful TUI** — Bubble Tea UI with markdown rendering and frog startup animation
- 🛡️ **9 Loop Detection Patterns** — ExactDup, IdentityCycle, UnproductiveStreak, FileReadRepeat, etc.
- 📦 **Context Management** — Spill to disk, time compaction, semantic consolidation, bloat detection
- 💬 **Session Management** — Tags, favorites, Markdown/HTML export, CWD tracking
- 🔒 **4-Layer Security** — Permissions + Safeguard (dangerous command blocking) + audit logging
- 🔁 **Auto-Retry + Watchdog** — Exponential backoff + anti-stuck watchdog timer
- 🐚 **Shell Integration** — Auto-completion (bash/zsh/fish), pipe mode with CWD context
- 🎯 **Cross-Platform** — Linux, macOS, Windows support

## Installation

### Using Go

```bash
go install github.com/starclaw/starclaw@latest
```

### Using Homebrew

Homebrew distribution is not available yet. Use the pre-built release binaries, Go, or npm for now.

### Using npm

```bash
npm install -g @starclaw/cli
```

### Pre-built Binaries

Download from [Releases](https://github.com/starclaw/starclaw/releases):

```bash
# Linux/macOS
curl -sSL https://get.starclaw.dev | sh

# Windows (PowerShell)
iwr -useb https://get.starclaw.dev/windows | iex
```

## Quick Start

### 1. Configure

```bash
starclaw setup
```

Or set environment variables:

```bash
export ANTHROPIC_AUTH_TOKEN="your-api-key"
export ANTHROPIC_BASE_URL="https://api.anthropic.com"
```

### 2. Chat

**One-shot mode:**
```bash
starclaw chat "What files are in this directory?"
```

**Interactive TUI:**
```bash
starclaw interactive
```

**Daemon Web UI:**
```bash
starclaw app
```

This starts the daemon when needed and opens the local Web UI at `http://127.0.0.1:7533/app/`. It provides a Codex-style workspace for chat, streaming runs, tool-call details, named agents, skills, sessions, and schedules.

For headless or remote shells:
```bash
starclaw app --no-open
```

To inspect launch readiness without starting anything:
```bash
starclaw app --check
```

`starclaw app` reuses an already-running daemon. If the browser cannot be opened automatically, copy the printed Web UI URL manually. The GUI Version page also shows local runtime context such as health, status, diagnostics, data, and config paths.

**Pipe input:**
```bash
cat main.go | starclaw chat "Explain this code"
```

## Available Tools

StarClaw provides 33 built-in tools for the AI agent:

| Category | Tools |
|---|---|
| File | `file_read`, `file_write`, `file_edit`, `filepreview` |
| Search | `glob`, `directory_list`, `grep` (ripgrep + VCS skip) |
| Shell | `bash` (output cap), `http` |
| System | `system_info`, `think`, `wait`, `version` |
| Desktop | `clipboard`, `notify`, `screenshot`, `applescript`, `process` |
| macOS | `accessibility`, `computer`, `browser` |
| Session | `session_search`, `memory_append` |
| Schedule | `schedule_create`, `schedule_list`, `schedule_update`, `schedule_remove` |
| Skills | `use_skill`, `skill` |
| Memory | `memory` |
| Publish | `publish_to_web` |
| Imaging | `imaging` (describe/resize/convert/OCR) |
| MCP | `mcp_tool` (varies by server) |

## Configuration

Configuration is stored in `~/.starclaw/config.yaml`:

```yaml
# Provider selection
provider: "anthropic"          # anthropic, openai, or ollama
endpoint: "https://api.anthropic.com"
api_key: "your-api-key"
model_tier: "medium"

# OpenAI
# openai_api_key: "sk-xxx"
# openai_model: "gpt-4o"

# Ollama (local)
# ollama_endpoint: "http://localhost:11434"
# ollama_model: "llama3.1"

agent:
  max_iterations: 25
  max_tokens: 8192
  temperature: 0
  thinking: true
  thinking_mode: "adaptive"
  thinking_budget: 10000

tools:
  bash_timeout: 120
  bash_max_output: 30000
  grep_max_results: 100
  result_truncation: 30000
  server_tool_timeout: 0  # MCP server tool timeout in seconds (0 = disabled)
  # mcp_expose: ["file_read", "grep", "directory_list", "version"]
  allowed: []  # Restrict to specific tools
  denied: []   # Block specific tools

audit:
  enabled: true  # Enable audit logging (default: true)

# MCP servers configuration (optional)
# mcp_servers:
#   github:
#     command: npx
#     args: ["-y", "@modelcontextprotocol/server-github"]
#     env:
#       GITHUB_PERSONAL_ACCESS_TOKEN: ${GITHUB_TOKEN}
#     keep_alive: true

# Update configuration (optional)
# update:
#   auto_check: true      # Check for updates on startup
#   auto_install: false   # Do not install automatically on startup
#   channel: stable       # Update channel (stable, beta)
#   cache_ttl: 24h        # How often to check
```

### Project-Level Configuration

Create `.starclaw/config.local.yaml` in your project directory for project-specific settings.

## Usage Examples

### Code Analysis
```bash
starclaw chat "Find all TODO comments in this project"
```

### File Operations
```bash
starclaw chat "Create a Python script that calculates fibonacci numbers"
```

### Refactoring
```bash
starclaw chat "Rename all occurrences of 'OldName' to 'NewName' in the src/ directory"
```

### With Auto-Approval
```bash
starclaw -y chat "Run 'go test ./...' and analyze the results"
```

## Audit Logging

StarClaw captures all tool calls to an append-only JSON-lines audit log for security and debugging purposes.

### Log Location

Audit logs are stored at:
```
~/.starclaw/logs/audit.log
```

### Log Format

Each line is a JSON object containing:

```json
{
  "timestamp": "2026-04-16T10:30:00Z",
  "session_id": "sess-abc123",
  "tool_name": "file_read",
  "input_summary": "{\"file_path\":\"/tmp/test.txt\"}",
  "output_summary": "Hello world content",
  "decision": "approved",
  "approved": true,
  "duration_ms": 5
}
```

### Secret Redaction

Sensitive data is automatically redacted from logs:
- AWS access keys (`AKIA...`)
- JWT tokens
- API keys (`sk-...`, `key-...`)
- Bearer tokens
- GitHub tokens (`ghp_...`, `gho_...`)
- Environment variables with secret-like names (`KEY=`, `SECRET=`, etc.)
- PEM certificate markers

### Querying Logs

```bash
# View recent entries
tail -f ~/.starclaw/logs/audit.log

# Pretty print with jq
jq . ~/.starclaw/logs/audit.log

# Filter by tool
grep '"tool_name":"bash"' ~/.starclaw/logs/audit.log | jq .

# Filter by session
grep '"session_id":"sess-abc123"' ~/.starclaw/logs/audit.log | jq .
```

### Disabling Audit Logging

Set `audit.enabled: false` in your config:

```yaml
audit:
  enabled: false
```

## Session Persistence

StarClaw saves conversation history to JSON files, allowing you to resume previous sessions and maintain context across restarts.

### Storage Location

Sessions are stored at:
```
~/.starclaw/sessions/
├── 2026-04-16-10-30-00-abcd1234.json
├── 2026-04-16-11-00-00-efgh5678.json
└── ...
```

### Session File Format

```json
{
  "id": "2026-04-16-10-30-00-abcd1234",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:35:00Z",
  "title": "Refactor database code",
  "cwd": "/home/user/myproject",
  "messages": [
    {"role": "user", "content": "Help me refactor..."},
    {"role": "assistant", "content": "I'll help you..."}
  ]
}
```

### List Sessions

```bash
starclaw sessions
```

Output:
```
ID                              Title                           Messages        Date
----------------------------------------------------------------------------------------------------
2026-04-16-10-30-00-abcd1234  Refactor database code                  12  2026-04-16
2026-04-16-11-00-00-efgh5678  Debug authentication issue               8  2026-04-16
```

### Resume a Session

```bash
# Resume specific session
starclaw --resume 2026-04-16-10-30-00-abcd1234 chat "Continue where we left off"

# Resume in interactive mode
starclaw --resume 2026-04-16-10-30-00-abcd1234 interactive
```

Sessions are automatically saved after each turn and on graceful exit.

## MCP Client Support

StarClaw supports the [Model Context Protocol (MCP)](https://modelcontextprotocol.io) for connecting to external tool servers.

### Configuring MCP Servers

Add MCP servers to your `~/.starclaw/config.yaml`:

```yaml
mcp_servers:
  github:
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: ${GITHUB_TOKEN}
    keep_alive: true
  
  filesystem:
    command: npx
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/docs"]
    disabled: false
```

### Managing MCP Servers

```bash
# List configured servers
starclaw mcp list
```

## MCP Server Support

StarClaw can also expose its local tools as an MCP stdio server:

```bash
starclaw mcp serve
```

By default, `mcp serve` exposes all registered local tools. For safer use with another MCP client, configure an allow-list and optional per-tool timeout:

```yaml
tools:
  server_tool_timeout: 30
  mcp_expose:
    - file_read
    - grep
    - directory_list
    - version
```

An empty or omitted `mcp_expose` list means "expose all local tools". MCP server calls are auto-approved because the MCP consumer is expected to provide its own authorization layer.

## Skills System

Skills are composable capabilities that can be activated on demand. They provide domain-specific knowledge and instructions to the AI.

### Creating a Skill

Create a skill directory in `~/.starclaw/skills/<skill-name>/` with a `SKILL.md` file:

```markdown
---
name: go-refactoring
description: Expert Go code refactoring assistant
license: MIT
allowed-tools: file_read file_write file_edit grep
---

# Go Refactoring Skill

You are an expert Go developer specializing in code refactoring.

## Guidelines

- Follow Go best practices and idioms
- Use meaningful variable names
- Add appropriate error handling
- Maintain backward compatibility when possible
```

### Using Skills

The AI can activate skills using the `use_skill` tool:

```
🔧 Tool: use_skill
   Args: {"name": "go-refactoring"}
```

Once activated, the skill's instructions are injected into the conversation context.

## Named Agents

Create specialized agent configurations with custom prompts and capabilities.

### Creating an Agent

Create an agent directory in `~/.starclaw/agents/<agent-name>/`:

```
~/.starclaw/agents/coder/
├── AGENT.md      # Agent instructions
├── MEMORY.md     # Persistent memory/context
└── config.yaml   # Agent-specific config
```

**AGENT.md:**
```markdown
# Coder Agent

You are an expert software engineer specializing in clean, maintainable code.

## Capabilities

- Code review and refactoring
- Test-driven development
- Performance optimization
- Documentation writing
```

**config.yaml:**
```yaml
max_iterations: 50
model_tier: large
tools:
  allowed: [file_read, file_write, file_edit, glob, grep, bash]
```

### Using Agents

```bash
starclaw --agent coder chat "Review this Go code for issues"
```

## Self-Update

StarClaw includes built-in update checks and one-command binary installation from GitHub Release assets.

### Manual Update Check

```bash
# Check for updates
starclaw update --check

# Install the latest version for your platform
starclaw update
```

### Automatic Updates

Enable automatic update checks on startup:

```yaml
update:
  auto_check: true    # Check on startup (once per day)
  auto_install: false # Do not install automatically on startup
  channel: stable     # Use stable releases
```

When an update is available, you'll see:
```
📦 Update available: v1.2.0 — run 'starclaw update --check' for details
```

## Security

- **Path Validation** - All file operations are restricted to current working directory by default
- **Approval System** - Destructive operations require explicit approval
- **Tool Filtering** - Configure allowed/denied tools via configuration
- **Audit Logging** - All tool calls are logged with automatic secret redaction
- **No Data Collection** - Your code and conversations stay local

## Development

### Prerequisites

- Go 1.22 or later
- Make (optional)

### Build

```bash
git clone https://github.com/starclaw/starclaw.git
cd starclaw
go build .
```

### Test

```bash
go test ./...
```

### Run

```bash
./starclaw --help
```

## Architecture

```
┌──────────────┐     ┌──────────────────────────┐     ┌─────────────┐
│   CLI / TUI  │────▶│       Agent Loop         │────▶│  LLM Client │
└──────────────┘     │  · Loop Detection        │     └─────────────┘
                     │  · Retry + Backoff       │
                     │  · Spill to Disk         │
                     │  · Read Tracker          │
                     └────────────┬─────────────┘
                                  │
           ┌──────────────────────┼──────────────────────┐
           ▼                      ▼                      ▼
   ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
   │  Local Tools │     │  MCP Client  │     │    Skills    │
   │  · file_*    │     │  · stdio     │     │  · SKILL.md  │
   │  · glob/grep │     │  · HTTP      │     │  · Frontmatter│
   │  · bash      │     │  · Reconnect │     │  · Registry  │
   │  · think     │     └──────────────┘     └──────────────┘
   │  · http      │
   └──────────────┘
```


## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by [ShanClaw](https://github.com/shan claw/shanclaw)
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Uses [Cobra](https://github.com/spf13/cobra) for CLI

---

<p align="center">Made with ❤️ by the StarClaw team</p>
