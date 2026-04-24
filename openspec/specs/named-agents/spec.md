# Specification: Named Agents

## Overview

| Field | Value |
|-------|-------|
| Feature | Named Agents |
| Type | Core Enhancement |
| Optional | Yes (default agent when none specified) |
| Default | No agent (global config only) |

## Purpose

Enable multiple agent definitions with independent instructions, memory, and configuration. Users can switch between agents for different tasks (coding, writing, ops, etc.).

## Agent Name Format

Agent names must match regex: `^[a-z0-9][a-z0-9_-]{0,63}$`

Requirements:
- Start with alphanumeric (lowercase)
- Lowercase only
- Allowed: `a-z`, `0-9`, `_`, `-`
- Maximum 64 characters
- Minimum 1 character

Valid: `coder`, `ops-bot`, `writer-2`
Invalid: `MyAgent`, `ops bot`, `coder.exe`, `-bot`

## File Structure

```
~/.starclaw/
├── config.yaml              # Global config
└── agents/
    ├── coder/               # Agent directory (name = "coder")
    │   ├── AGENT.md         # Agent instructions (required)
    │   ├── MEMORY.md        # Persistent memory (optional)
    │   ├── config.yaml      # Agent-specific config (optional)
    │   ├── commands/        # Custom slash commands (optional)
    │   │   ├── review.md
    │   │   └── refactor.md
    │   └── skills/          # Agent-specific skills (optional)
    │       └── code-review/
    │           └── SKILL.md
    └── writer/
        ├── AGENT.md
        ├── MEMORY.md
        └── config.yaml
```

## AGENT.md Format

```markdown
---
name: coder
description: A coding assistant specialized in Go and Python
---

# Coder Agent

You are a senior software engineer with expertise in Go and Python.

## Guidelines

- Write clean, idiomatic code
- Follow language-specific best practices
- Explain your reasoning
- Suggest tests when appropriate

## Preferences

- Prefer composition over inheritance
- Use standard library when possible
- Document exported functions
```

## MEMORY.md Format

```markdown
# Memory

- User prefers snake_case for Go variables
- Project uses testify for testing
- User likes detailed comments
```

Agent memory is:
- Loaded at agent initialization
- Persisted across sessions
- Can be modified via `memory_append` tool
- Used to build system prompt

## Agent Config (config.yaml)

```yaml
# ~/.starclaw/agents/coder/config.yaml

# Model overrides
model: "claude-sonnet-4"
max_tokens: 8192
temperature: 0.1
max_iterations: 30

# Tool filtering
tools:
  allow:
    - file_read
    - file_write
    - file_edit
    - bash
    - grep
  deny:
    - http  # Prevent external calls

# Auto-approve (use with caution)
auto_approve: false
```

## Config Merge Order

Configuration is merged from multiple sources (low to high priority):

1. **Global defaults** - Built-in defaults
2. **Global config** - `~/.starclaw/config.yaml`
3. **Agent config** - `~/.starclaw/agents/<name>/config.yaml`
4. **Project config** - `./.starclaw/config.local.yaml` (gitignored)

### Merge Rules

- **Scalars**: Higher priority wins (override)
- **Lists**: Merge + deduplicate (lower priority first)
- **Structs**: Field-level merge
- **Maps**: Merge keys (higher priority wins on conflict)

Example:

```yaml
# Global config
tools:
  allowed: [file_read, file_write, bash]
  denied: []

# Agent config
tools:
  allowed: [file_edit]  # Adds to global
  denied: [http]        # Adds to global

# Result (merged)
tools:
  allowed: [file_read, file_write, bash, file_edit]
  denied: [http]
```

## Agent Loader API

```go
package agents

// Agent represents a loaded agent definition
type Agent struct {
    Name     string
    Prompt   string           // From AGENT.md (after frontmatter)
    Memory   string           // From MEMORY.md
    Config   *AgentConfig     // From config.yaml (nil = use global)
    Commands map[string]string // name → content from commands/
}

// Agent configuration overrides
type AgentConfig struct {
    Model         *string       `yaml:"model"`
    MaxTokens     *int          `yaml:"max_tokens"`
    Temperature   *float64      `yaml:"temperature"`
    MaxIterations *int          `yaml:"max_iterations"`
    Tools         *ToolsFilter  `yaml:"tools"`
    AutoApprove   *bool         `yaml:"auto_approve"`
}

type ToolsFilter struct {
    Allow []string `yaml:"allow"`
    Deny  []string `yaml:"deny"`
}

// Agent info (without full content)
type AgentInfo struct {
    Name        string
    Description string // From AGENT.md frontmatter
}

// Load agent by name
func LoadAgent(name string) (*Agent, error)

// List available agents
func ListAgents() ([]AgentInfo, error)

// Validate agent name format
func ValidateAgentName(name string) error

// Parse agent mention from text (e.g., "@coder do this")
func ParseAgentMention(text string) (name string, rest string, found bool)
```

## CLI Usage

```bash
# Use specific agent
starclaw --agent coder "review this Go code"
starclaw --agent writer "draft a blog post"

# List available agents
starclaw agents list

# Interactive mode with agent
starclaw --agent coder

# One-shot with agent
starclaw --agent ops-bot "check production health"
```

## Agent Mention in Chat

Users can switch agents mid-conversation:

```
User: @coder review this function
User: @writer draft documentation for this
```

Parse with:

```go
mentionRe := regexp.MustCompile(`^@([a-zA-Z0-9][a-zA-Z0-9_-]*)(?:\s|$)`)
```

## System Prompt Construction

When an agent is active, the system prompt is built as:

```
[Base System Prompt]

[Agent Prompt from AGENT.md]

[Agent Memory from MEMORY.md]
```

Order matters - agent instructions override defaults.

## Custom Commands

Agents can define custom slash commands:

```markdown
<!-- ~/.starclaw/agents/coder/commands/review.md -->
---
name: review
description: Review code for issues
---

Review the provided code for:
- Bugs and logic errors
- Performance issues
- Security vulnerabilities
- Style violations

Provide specific suggestions with line numbers.
```

Usage:

```
/review  # In TUI, executes the command
```

## Error Handling

| Error | Behavior |
|-------|----------|
| Agent not found | Error with available agents list |
| Invalid agent name | Validation error before loading |
| Missing AGENT.md | Error, agent incomplete |
| Config parse error | Error with line number |
| Permission denied | Error with path |

## Testing Requirements

### Unit Tests

- Name validation (regex)
- Config merge logic
- Agent struct creation

### Integration Tests

- Load agent from temp directory
- Full agent lifecycle
- Config merging scenarios

### Test Fixtures

```go
func createTestAgent(t *testing.T, dir, name string) {
    agentDir := filepath.Join(dir, "agents", name)
    os.MkdirAll(agentDir, 0755)
    
    // AGENT.md
    agentContent := fmt.Sprintf(`---
name: %s
description: Test agent
---

Test agent instructions.
`, name)
    os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(agentContent), 0644)
}
```

## Security Considerations

- Agent names validated before path construction (prevent traversal)
- Agent files are user-controlled (same user as StarClaw)
- Tool filtering limits agent capabilities
- No sandboxing between agents (shared process)

## Backward Compatibility

- Without `--agent`, behavior unchanged
- Existing configs work without modification
- Agent config is optional overlay

## References

- ShanClaw Reference: `/Users/timmy/PycharmProjects/ShanClaw/internal/agents/loader.go`
