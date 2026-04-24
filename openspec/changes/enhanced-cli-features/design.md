# Design: Enhanced CLI Features

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           StarClaw CLI                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐         │
│  │   MCP      │  │  Named     │  │  Skills    │  │   Self     │         │
│  │  Client    │  │  Agents    │  │  System    │  │  Update    │         │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘         │
│        │               │               │               │                │
│        ▼               ▼               ▼               ▼                │
│  ┌────────────────────────────────────────────────────────────┐         │
│  │                    Configuration System                     │         │
│  │  (Extended with MCP servers, agent selection, skills path)  │         │
│  └────────────────────────────────────────────────────────────┘         │
│        │                                                                 │
│        ▼                                                                 │
│  ┌────────────────────────────────────────────────────────────┐         │
│  │                    Tool Registry                           │         │
│  │  (Local tools + MCP tools merged with priority)            │         │
│  └────────────────────────────────────────────────────────────┘         │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## 1. MCP Client Design

### Component Structure

```
internal/mcp/
├── client.go          # ClientManager implementation
├── client_test.go     # Unit and integration tests
└── config.go          # MCPServerConfig types
```

### ClientManager

```go
type ClientManager struct {
    mu          sync.Mutex
    clients     map[string]mcpclient.MCPClient  // server name → client
    configs     map[string]MCPServerConfig       // server name → config
    toolCache   map[string][]RemoteTool          // server name → tools
    reconnectMu map[string]*sync.Mutex           // per-server reconnect lock
}

func (m *ClientManager) ConnectAll(ctx context.Context, servers map[string]MCPServerConfig) ([]RemoteTool, error)
func (m *ClientManager) CallTool(ctx context.Context, serverName, toolName string, args json.RawMessage) (*ToolCallResult, error)
func (m *ClientManager) Close() error
```

### Configuration Extension

```yaml
# ~/.starclaw/config.yaml
mcp_servers:
  github:
    command: npx
    args: ["-y", "@modelcontextprotocol/server-github"]
    env:
      GITHUB_PERSONAL_ACCESS_TOKEN: ${GITHUB_TOKEN}
    keep_alive: true  # stay connected between turns
  
  fetch:
    command: uvx
    args: ["mcp-server-fetch"]
    type: stdio  # default
```

### Tool Integration

```go
// Register MCP tools alongside local tools
func (r *ToolRegistry) RegisterMCPTools(remoteTools []mcp.RemoteTool) {
    for _, rt := range remoteTools {
        r.Register(&MCPToolAdapter{
            serverName: rt.ServerName,
            tool:       rt.Tool,
        })
    }
}
```

### Testing Strategy

| Test Type | Approach |
|-----------|----------|
| Unit | Mock MCP client interface |
| Integration | Test MCP server using mcp-go test helpers |
| Config | Table-driven tests for YAML parsing |
| Error | Simulate connection failures, timeouts |

---

## 2. Named Agents Design

### File Structure

```
~/.starclaw/
├── config.yaml              # Global config
└── agents/
    ├── coder/
    │   ├── AGENT.md         # Agent instructions
    │   ├── MEMORY.md        # Agent memory/persistence
    │   └── config.yaml      # Agent-specific config
    └── writer/
        ├── AGENT.md
        ├── MEMORY.md
        └── config.yaml
```

### Agent Loader

```go
type Agent struct {
    Name     string
    Prompt   string           // From AGENT.md
    Memory   string           // From MEMORY.md
    Config   *AgentConfig     // From config.yaml
    Commands map[string]string // Custom slash commands
}

type AgentConfig struct {
    Model       *string        `yaml:"model"`
    MaxTokens   *int           `yaml:"max_tokens"`
    Temperature *float64       `yaml:"temperature"`
    Tools       *ToolsFilter   `yaml:"tools"`
}

// Loader functions
func LoadAgent(name string) (*Agent, error)
func ListAgents() ([]AgentInfo, error)
func ValidateAgentName(name string) error
```

### Config Merge Order

Priority (low to high):
1. `~/.starclaw/config.yaml` (global defaults)
2. `~/.starclaw/agents/<name>/config.yaml` (agent overlay)
3. `./.starclaw/config.local.yaml` (project overlay, gitignored)

Merge rules:
- Scalars: higher priority wins
- Lists: merge + deduplicate
- Structs: field-level merge

### CLI Integration

```go
// New flag
cmd.Flags().StringVar(&agentName, "agent", "", "Use named agent")

// Usage
starclaw --agent coder "review this PR"
starclaw --agent writer "draft a blog post"
```

### Testing Strategy

| Test Type | Coverage |
|-----------|----------|
| Name validation | Regex tests for valid/invalid names |
| Config merge | Table tests for merge scenarios |
| File loading | Mock filesystem with temp dirs |
| Integration | End-to-end agent loading |

---

## 3. Skills System Design

### SKILL.md Format

```markdown
---
name: github-ops
description: Advanced GitHub operations for CI/CD
license: MIT
compatibility: ">=1.0.0"
metadata:
  author: starclaw-team
  version: "1.2.0"
allowed-tools: file_read bash http
---

# GitHub Operations Skill

You are a GitHub CI/CD expert. You help users with:

- Setting up GitHub Actions workflows
- Debugging failed builds
- Optimizing workflow performance
...
```

### Skill Registry

```go
type Skill struct {
    Name          string
    Description   string
    Prompt        string           // Skill body content
    License       string
    Compatibility string
    Metadata      map[string]string
    AllowedTools  []string         // Tool allowlist
    Source        string           // "global", "bundled", "agent"
    Dir           string           // Source directory
}

type SkillMeta struct {
    Name        string
    Description string
    Source      string
}

// Loader
func LoadSkills(sources ...SkillSource) ([]*Skill, error)
func ValidateSkillName(name string) error
```

### Skill Sources

Priority order:
1. **Bundled** - Embedded in binary (default skills)
2. **Global** - `~/.starclaw/skills/`
3. **Agent** - `~/.starclaw/agents/<name>/skills/`

### Activation

```go
// Tool: use_skill
func (t *UseSkillTool) Execute(args UseSkillArgs) (string, error) {
    skill, err := skills.LoadSkill(args.Name)
    if err != nil {
        return "", err
    }
    // Inject skill prompt into context
    return skill.Prompt, nil
}
```

### Testing Strategy

| Test Type | Coverage |
|-----------|----------|
| Frontmatter | Parse valid/invalid frontmatter |
| Name validation | Skill naming rules |
| Loading | Load from multiple sources |
| Integration | Full skill activation flow |

---

## 4. Self-Update Design

### Update Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Check     │────▶│   Compare   │────▶│   Update    │
│   Version   │     │  Versions   │     │   Binary    │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
  GitHub API         semver compare     Atomic replace
```

### Implementation

```go
package update

const (
    RepoOwner = "starclaw"
    RepoName  = "starclaw"
)

// Check for available update
func CheckForUpdate(currentVersion string) (*Release, bool, error)

// Perform update
func DoUpdate(currentVersion string) (string, error)

// Background auto-update (non-blocking)
func AutoUpdate(currentVersion, cacheDir string) string
```

### CLI Commands

```bash
starclaw update              # Manual update check
starclaw update --check      # Check only, don't install
```

### Configuration

```yaml
# ~/.starclaw/config.yaml
update:
  auto_check: true           # Check on startup
  channel: stable            # stable, beta
  cache_ttl: 24h             # Hours between checks
```

### Security

- Verify checksums from `checksums.txt`
- Atomic binary replacement
- Rollback on failure
- Skip for dev builds (non-semver versions)

### Testing Strategy

| Test Type | Coverage |
|-----------|----------|
| Version compare | semver edge cases |
| GitHub API | Mock GitHub responses |
| Update flow | Test with fake release |
| Error handling | Network failure, permission denied |

---

## Integration Points

### Tool Registration Priority

```go
// Order matters: later registrations override earlier
func InitializeTools(cfg *Config) *ToolRegistry {
    reg := tools.RegisterLocalTools()
    
    // 1. Local tools (highest priority)
    
    // 2. MCP tools
    if cfg.MCPServers != nil {
        mcpTools := mcpClient.ConnectAll(cfg.MCPServers)
        reg.RegisterMCPTools(mcpTools)
    }
    
    return reg
}
```

### Agent Loop Integration

```go
type AgentLoop struct {
    // Existing fields...
    agent        *agents.Agent    // Current agent (optional)
    mcpManager   *mcp.ClientManager
}

func (a *AgentLoop) buildSystemPrompt() string {
    parts := []string{basePrompt}
    
    if a.agent != nil {
        parts = append(parts, a.agent.Prompt)
        parts = append(parts, a.agent.Memory)
    }
    
    return strings.Join(parts, "\n\n")
}
```

---

## Dependencies

### New Dependencies

```go
// go.mod additions
require (
    github.com/mark3labs/mcp-go v0.2.0
    github.com/creativeprojects/go-selfupdate v1.0.0
    github.com/Masterminds/semver/v3 v3.2.0
    github.com/adrg/frontmatter v0.2.0
)
```

### Version Requirements

- Go 1.26+ (existing requirement)
- All dependencies must be MIT/Apache/BSD licensed

---

## Migration Path

### Backward Compatibility

All new features are **opt-in**:
- Without `--agent`, behavior is unchanged
- Without MCP config, no MCP tools loaded
- Without skills, no skill functionality
- Auto-update disabled by default

### Config Migration

Existing `config.yaml` continues to work. New fields are optional with sensible defaults.

---

## Success Metrics

| Metric | Target |
|--------|--------|
| MCP servers supported | 5+ (GitHub, fetch, filesystem, etc.) |
| Agent load time | <100ms |
| Skill load time | <50ms per skill |
| Update check time | <5s |
| Test coverage | >80% per feature |
| Binary size increase | <5MB |
