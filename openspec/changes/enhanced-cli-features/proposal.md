# Proposal: Enhanced CLI Features

## Summary

Add four high-priority features from ShanClaw to StarClaw to enhance its capabilities as an AI-powered CLI agent:

1. **MCP Client Support** - Integrate with Model Context Protocol servers for extended tool ecosystem
2. **Named Agents** - Support multiple agent definitions with independent instructions and memory
3. **Skills System** - Composable capabilities loaded from SKILL.md files following Anthropic spec
4. **Self-Update Mechanism** - Automatic update checking and installation from GitHub releases

## Motivation

StarClaw currently provides basic AI CLI functionality with 10 built-in tools. To compete with modern AI agent platforms and provide extensibility, we need:

- **MCP Integration**: Access external tools (GitHub, databases, Slack, etc.) without hardcoding
- **Named Agents**: Support different personas/workflows (coder, writer, ops) with specialized instructions
- **Skills**: Reusable, shareable capabilities that can be versioned and distributed
- **Auto-Update**: Improve user experience with seamless updates

## Scope

### In Scope

- MCP client manager with stdio/HTTP transport support
- Agent definition loading from `~/.starclaw/agents/<name>/`
- SKILL.md parser with frontmatter support
- Self-update using go-selfupdate library
- Configuration extensions for MCP servers and agent selection
- CLI flag `--agent` for agent selection
- Command `mcp` for MCP server management
- Command `update` for manual update check

### Out of Scope

- MCP server mode (we'll only implement client)
- WebSocket daemon mode (ShanClaw-specific feature)
- macOS GUI control tools (platform-specific)
- Scheduling system (launchd-specific)

## Success Criteria

| Feature | Criteria |
|---------|----------|
| MCP Client | Can connect to at least 3 different MCP servers and invoke tools |
| Named Agents | Can load and switch between 2+ agents with different instructions |
| Skills | Can load and activate skills from SKILL.md files |
| Self-Update | Can check and install updates from GitHub releases |
| Testing | All features have >80% test coverage |

## References

- ShanClaw source: `/Users/timmy/PycharmProjects/ShanClaw/internal/mcp/`
- MCP Protocol: https://modelcontextprotocol.io/
- Anthropic Skills Spec: https://agentskills.io/specification
- go-selfupdate: https://github.com/creativeprojects/go-selfupdate

## Risks

| Risk | Mitigation |
|------|------------|
| MCP SDK compatibility | Use official mcp-go SDK, pin version |
| Agent name validation | Strict regex `^[a-z0-9][a-z0-9_-]{0,63}$` |
| Update failures | Atomic updates with rollback capability |
| Breaking changes | Maintain backward compatibility for existing config |

## Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Phase 1 | Week 1 | Config extensions, agents/ module skeleton |
| Phase 2 | Week 2-3 | MCP client implementation, tests |
| Phase 3 | Week 4 | Agent loader, config merging, tests |
| Phase 4 | Week 5 | Skills system, tests |
| Phase 5 | Week 6 | Self-update, integration tests |

## Decision

Status: **DRAFT**

Ready for design phase.
