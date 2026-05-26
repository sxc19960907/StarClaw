# Configuration Guide

## Configuration Locations

StarClaw uses a hierarchical configuration system:

1. **Global Config** (`~/.starclaw/config.yaml`)
   - User-level defaults
   - API keys and endpoints

2. **Local Config** (`.starclaw/config.local.yaml`)
   - Project-specific overrides
   - Team-shared settings

3. **Environment Variables**
   - Runtime overrides
   - CI/CD integration

## Configuration Format

### Full Example

```yaml
# API Configuration
endpoint: "https://api.anthropic.com"
api_key: "sk-ant-api03-..."
model_tier: "standard"  # or "high", "low"

# Agent Behavior
agent:
  max_iterations: 25      # Maximum tool call cycles
  max_tokens: 8192        # Response token limit
  temperature: 0          # 0 = deterministic, 1 = creative

# Tool Settings
tools:
  bash_timeout: 120         # seconds
  bash_max_output: 30000    # characters
  result_truncation: 30000  # characters
  args_truncation: 200      # characters
  grep_max_results: 100      # maximum grep matches
  server_tool_timeout: 0     # MCP server tool timeout in seconds (0 = disabled)
  mcp_expose: []             # MCP server expose allow-list (empty = all local tools)
  allowed: []               # Restrict to these tools (empty = all)
  denied: []                # Block these tools
```

### Minimal Example

```yaml
api_key: "your-api-key"
```

## Environment Variables

Override any config value via environment:

```bash
export ANTHROPIC_AUTH_TOKEN="sk-..."
export ANTHROPIC_BASE_URL="https://api.anthropic.com"
export ANTHROPIC_MODEL="kimi-k2.5"
```

## Tool Filtering

### Allow Specific Tools

```yaml
tools:
  allowed:
    - file_read
    - glob
    - grep
```

### Block Dangerous Tools

```yaml
tools:
  denied:
    - bash
    - file_write
```

## MCP Server Mode

Run StarClaw as an MCP stdio server with:

```bash
starclaw mcp serve
```

The server exposes StarClaw local tools to MCP-compatible clients. By default, all registered local tools are exposed.

### Safer MCP Server Exposure

Use `tools.mcp_expose` to publish only the tools another MCP client should see:

```yaml
tools:
  server_tool_timeout: 30
  mcp_expose:
    - file_read
    - grep
    - directory_list
    - version
```

`server_tool_timeout` is measured in seconds. Set it to `0` to disable the MCP server-level timeout.

`mcp_expose` is separate from `tools.allowed` and `tools.denied`: it controls which tools are registered with the MCP server. If `mcp_expose` is omitted or empty, `starclaw mcp serve` exposes all registered local tools.

## Model Tiers

| Tier | Description | Use Case |
|------|-------------|----------|
| `low` | Fastest, lowest cost | Simple queries |
| `standard` | Balanced | General use |
| `high` | Best quality | Complex tasks |

## Security Notes

- Config file permissions: `0600` (user read/write only)
- API keys are trimmed of whitespace
- Never commit config files to version control
