# Proposal: Add Session Persistence

## Summary

Add session persistence to save conversation history to JSON files. Users can resume previous sessions, browse session history, and have context preserved across restarts.

## Motivation

Currently, StarClaw has no memory - each run starts fresh. Session persistence provides:

1. **Context preservation** - Resume long-running conversations
2. **History browsing** - Review past conversations
3. **Multi-session workflow** - Switch between different tasks
4. **Debugging** - Review conversation history for issues
5. **Audit trail** - Record of AI interactions

## Scope

### In Scope
- JSON file storage for sessions (`~/.starclaw/sessions/`)
- Session Store (CRUD operations)
- Session Manager (current session lifecycle)
- Agent Loop integration (save/load messages)
- CLI commands for session management (`--resume`, `--list-sessions`)
- Auto-save on exit
- Session listing with metadata

### Out of Scope
- Session search (Phase 2 - use grep on JSON files for now)
- Session expiration/cleanup
- Session encryption (Phase 2)
- Cloud sync
- Session sharing

## Success Criteria

- [ ] Sessions saved to JSON files with 0600 permissions
- [ ] Can resume any previous session
- [ ] Auto-save on graceful exit
- [ ] Session listing shows title, date, message count
- [ ] Agent Loop saves messages after each turn
- [ ] Session ID is human-readable (timestamp-based)
- [ ] Works with both TUI and one-shot modes
- [ ] Unit tests for Store and Manager
- [ ] Integration tests for resume flow
- [ ] Documentation for session commands

## Reference

Based on ShanClaw's implementation:
- `internal/session/store.go`
- `internal/session/manager.go`
- `internal/session/session.go` (Session struct)
