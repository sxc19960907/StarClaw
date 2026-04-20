# StarClaw

[![CI](https://github.com/starclaw/starclaw/actions/workflows/ci.yml/badge.svg)](https://github.com/starclaw/starclaw/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/starclaw/starclaw)](https://goreportcard.com/report/github.com/starclaw/starclaw)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**StarClaw** is an AI-powered CLI agent with local tool execution capabilities. It provides a terminal interface for interacting with Large Language Models (LLMs) while enabling the AI to safely execute commands on your local system through a controlled tool system.

![StarClaw Demo](docs/demo.gif)

## Features

- 🤖 **AI-Powered Conversations** - Chat with Claude, Kimi, or other LLMs
- 🛠️ **Local Tool Execution** - Read files, search code, run shell commands
- 🔒 **Security First** - Path validation, approval dialogs, configurable tool allowlists
- 💻 **Interactive TUI** - Beautiful terminal UI with Bubble Tea
- 🚀 **One-Shot Mode** - Quick queries from command line or pipes
- ⚡ **Fast Startup** - Optimized for sub-100ms cold start
- 🎯 **Cross-Platform** - Linux, macOS, Windows support

## Installation

### Using Go

```bash
go install github.com/starclaw/starclaw@latest
```

### Using Homebrew

```bash
brew tap starclaw/tap
brew install starclaw
```

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

**Pipe input:**
```bash
cat main.go | starclaw chat "Explain this code"
```

## Available Tools

StarClaw provides 10 built-in tools for the AI agent:

| Tool | Description | Requires Approval |
|------|-------------|-------------------|
| `file_read` | Read file contents with line numbers | Yes |
| `file_write` | Write content to a file | Yes |
| `file_edit` | Find and replace in files | Yes |
| `glob` | Find files matching a pattern | No |
| `directory_list` | List directory contents | No |
| `grep` | Search file contents | No |
| `think` | AI reasoning and planning | No |
| `system_info` | Get system information | No |
| `http` | Make HTTP requests | Yes |
| `bash` | Execute shell commands | Yes |

## Configuration

Configuration is stored in `~/.starclaw/config.yaml`:

```yaml
endpoint: "https://api.anthropic.com"
api_key: "your-api-key"
model_tier: "standard"

agent:
  max_iterations: 25
  max_tokens: 8192
  temperature: 0

tools:
  bash_timeout: 120
  bash_max_output: 30000
  result_truncation: 30000
  allowed: []  # Restrict to specific tools
  denied: []   # Block specific tools

audit:
  enabled: true  # Enable audit logging (default: true)
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
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   CLI/TUI   │────▶│  Agent Loop │────▶│  LLM Client │
└─────────────┘     └─────────────┘     └─────────────┘
                            │
                            ▼
                    ┌─────────────┐
                    │ Tool System │
                    │  - file_*   │
                    │  - glob     │
                    │  - grep     │
                    │  - bash     │
                    └─────────────┘
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
