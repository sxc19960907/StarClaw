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
  token_budget:
    max_input_tokens: 0    # 0 disables this limit
    max_output_tokens: 0
    max_total_tokens: 0
    hard_stop: false       # Stop before the next model call when exhausted/projected exhausted

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

## Runtime Budget, Routing, and Fallback

`agent.token_budget` configures per-run runtime token budget enforcement:

```yaml
agent:
  token_budget:
    max_input_tokens: 120000
    max_output_tokens: 12000
    max_total_tokens: 132000
    hard_stop: true
```

Provider usage is authoritative when returned. If provider usage is missing, StarClaw reports an `unknown` budget status instead of inventing exact totals. With `hard_stop: true`, the agent loop stops before the next model call once concrete or projected usage is exhausted.

Complexity routing and fallback metadata are local deterministic run metadata. They help Astria and the daemon explain route choices, budget-constrained paths, provider-error fallback, budget-exhausted fallback, and repeated-failure fallback. They do not require a hosted routing service.

## Security Notes

- Config file permissions: `0600` (user read/write only)
- API keys are trimmed of whitespace
- Never commit config files to version control
- Runtime observability is local-first: metrics are aggregate-only, trace export writes only to an explicit local path, and no external collector is configured by default.
- Structured events, trace read/export, diagnostics, run summaries, replay plans, and Web UI trace/recovery renderers redact prompt text, assistant text, tool arguments, provider payloads, API keys, bearer tokens, passwords, and secret-like values.
- Detailed run Prompt/Result views and explicit local copy/rerun actions are operator-facing UI affordances. Treat them as local review surfaces, not shareable support bundles.
