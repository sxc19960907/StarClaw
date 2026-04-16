# Design Document

## Architecture Overview

StarClaw follows a layered architecture:

```
┌─────────────────────────────────────────┐
│           CLI Interface                 │
│    (cobra + custom command handlers)    │
├─────────────────────────────────────────┤
│           TUI Interface                 │
│    (bubbletea event-driven model)       │
├─────────────────────────────────────────┤
│           Agent Loop                    │
│    (conversation + tool orchestration)  │
├─────────────────────────────────────────┤
│      LLM Client      │   Tool System    │
│  (HTTP/REST client)  │  (7 built-in)    │
└─────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Tool System

- **Interface-based**: All tools implement `Tool` interface
- **Registry pattern**: Tools registered at runtime
- **Security**: Path validation before execution
- **Approval**: Destructive tools require explicit approval

### 2. Agent Loop

- **Stateless**: No persistent conversation state
- **Event-driven**: EventHandler interface for UI updates
- **Iterative**: Supports multi-step tool chains
- **Safe**: Max iteration limits prevent infinite loops

### 3. Configuration

- **Hierarchical**: Global → Local → Environment
- **Simple**: YAML-based, minimal structure
- **Secure**: API keys stored with restricted permissions

### 4. TUI

- **Framework**: Bubbletea for Go-native experience
- **Model-Update-View**: Pure functional UI updates
- **Responsive**: Handles terminal resize
- **Accessible**: Keyboard-driven navigation

## Security Model

1. **Path Restriction**: File operations limited to CWD
2. **Approval Gates**: User consent for destructive ops
3. **Tool Filtering**: Configurable allow/deny lists
4. **No Network**: Local-only file access

## Performance Considerations

- **Startup**: <100ms target (achieved ~50ms)
- **Memory**: <50MB working set
- **Binary Size**: ~13MB (compressed)
- **Caching**: Config and tool registry cached

## Future Extensions

- [ ] MCP (Model Context Protocol) support
- [ ] Plugin system for custom tools
- [ ] Session persistence
- [ ] Multi-model support
