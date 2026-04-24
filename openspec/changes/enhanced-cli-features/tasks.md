# Tasks: Enhanced CLI Features

## Phase 1: Foundation (Week 1)

### Task 1.1: Extend Configuration System
**Owner:** TBD
**Duration:** 2 days
**Depends:** None

- [x] Add `MCPServerConfig` type to `internal/config/config.go`
- [x] Add `UpdateConfig` type for auto-update settings
- [x] Extend `Config` struct with new sections
- [x] Update `Load()` to handle new config fields
- [x] Write unit tests for config parsing

**Files:**
- `internal/config/config.go`
- `internal/config/config_test.go`

**Acceptance Criteria:**
- Can parse extended config with MCP servers
- Backward compatible with existing configs
- Tests cover all new config fields

---

### Task 1.2: Create Module Skeletons
**Owner:** TBD
**Duration:** 1 day
**Depends:** None

- [x] Create `internal/mcp/` package with placeholder files
- [x] Create `internal/agents/` package with types
- [x] Create `internal/skills/` package with types
- [x] Create `internal/update/` package with types
- [x] Add go.mod dependencies

**Files:**
- `internal/mcp/client.go` (skeleton)
- `internal/agents/loader.go` (skeleton)
- `internal/skills/registry.go` (skeleton)
- `internal/update/selfupdate.go` (skeleton)
- `go.mod`

**Acceptance Criteria:**
- All packages compile
- Dependencies vendored/imported correctly
- No test failures

---

### Task 1.3: Update CLI Commands
**Owner:** TBD
**Duration:** 2 days
**Depends:** 1.1

- [x] Add `--agent` flag to root command
- [x] Add `mcp` subcommand group
- [x] Add `update` subcommand
- [x] Update help text

**Files:**
- `cmd/root.go`
- `cmd/mcp.go` (new)
- `cmd/update.go` (new)

**Acceptance Criteria:**
- `starclaw --help` shows new flags
- `starclaw mcp --help` works
- `starclaw update --help` works

---

## Phase 2: MCP Client (Week 2-3)

### Task 2.1: Implement MCP ClientManager
**Owner:** TBD
**Duration:** 3 days
**Depends:** 1.2

- [x] Implement `ClientManager` struct
- [x] Implement `ConnectAll()` with parallel connections
- [x] Implement `CallTool()` with error handling
- [x] Implement `Close()` for cleanup
- [x] Add per-server mutex for thread safety

**Files:**
- `internal/mcp/client.go`
- `internal/mcp/config.go`

**Reference:** `/Users/timmy/PycharmProjects/ShanClaw/internal/mcp/client.go`

**Acceptance Criteria:**
- Can connect to multiple MCP servers concurrently
- Tool calls work with proper error handling
- Resources cleaned up on Close()

---

### Task 2.2: MCP Tool Adapter
**Owner:** TBD
**Duration:** 2 days
**Depends:** 2.1

- [ ] Create `MCPTool` adapter implementing `Tool` interface
- [ ] Convert MCP tool schema to StarClaw tool schema
- [ ] Handle argument serialization
- [ ] Handle result deserialization

**Files:**
- `internal/mcp/tool.go` (new)
- `internal/mcp/tool_test.go`

**Acceptance Criteria:**
- MCP tools integrate with ToolRegistry
- Tool calls execute correctly
- Results formatted properly

---

### Task 2.3: MCP Integration Tests
**Owner:** TBD
**Duration:** 2 days
**Depends:** 2.2

- [ ] Write unit tests with mocked MCP client
- [ ] Write integration tests with real MCP server
- [ ] Test connection failures and recovery
- [ ] Test concurrent tool calls

**Files:**
- `internal/mcp/client_test.go`
- `internal/mcp/integration_test.go`

**Acceptance Criteria:**
- >80% test coverage
- Integration tests pass with real MCP server
- Concurrent access safe

---

## Phase 3: Named Agents (Week 4)

### Task 3.1: Implement Agent Loader
**Owner:** TBD
**Duration:** 2 days
**Depends:** 1.2

- [x] Implement `LoadAgent()` function
- [x] Implement `ListAgents()` function
- [x] Implement `ValidateAgentName()` with regex
- [x] Load AGENT.md and MEMORY.md files
- [x] Parse agent-specific config.yaml

**Files:**
- `internal/agents/loader.go`
- `internal/agents/types.go`

**Reference:** `/Users/timmy/PycharmProjects/ShanClaw/internal/agents/loader.go`

**Acceptance Criteria:**
- Can load agent from `~/.starclaw/agents/<name>/`
- Name validation rejects invalid names
- Returns meaningful errors for missing files

---

### Task 3.2: Config Merge Logic
**Owner:** TBD
**Duration:** 2 days
**Depends:** 3.1

- [ ] Implement config merge function
- [ ] Handle scalar override
- [ ] Handle list merge+dedup
- [ ] Handle struct field-level merge
- [ ] Write comprehensive tests

**Files:**
- `internal/config/merge.go` (new)
- `internal/config/merge_test.go`

**Acceptance Criteria:**
- Global < Agent < Project merge works correctly
- Lists are deduplicated
- Nil values handled properly

---

### Task 3.3: Agent Loop Integration
**Owner:** TBD
**Duration:** 2 days
**Depends:** 3.2

- [ ] Modify AgentLoop to accept Agent parameter
- [ ] Update system prompt builder to include agent instructions
- [ ] Wire up `--agent` flag to agent loading
- [ ] Handle agent not found error

**Files:**
- `internal/agent/loop.go`
- `cmd/root.go`
- `cmd/chat.go`

**Acceptance Criteria:**
- `starclaw --agent coder "query"` works
- Agent prompt injected into system prompt
- Memory loaded and available

---

### Task 3.4: Agent Tests
**Owner:** TBD
**Duration:** 1 day
**Depends:** 3.3

- [ ] Unit tests for loader functions
- [ ] Unit tests for config merge
- [ ] Integration tests for agent loading
- [ ] Test custom commands loading

**Files:**
- `internal/agents/loader_test.go`
- `internal/config/merge_test.go`

**Acceptance Criteria:**
- >80% test coverage
- All edge cases covered

---

## Phase 4: Skills System (Week 5)

### Task 4.1: Implement Skill Registry
**Owner:** TBD
**Duration:** 1 day
**Depends:** 1.2

- [x] Define `Skill` struct
- [x] Define `SkillMeta` struct
- [x] Implement `ToMeta()` method

**Files:**
- `internal/skills/loader.go`

**Reference:** `/Users/timmy/PycharmProjects/ShanClaw/internal/skills/registry.go`

**Acceptance Criteria:**
- Types defined with proper JSON tags
- Methods implemented

---

### Task 4.2: Implement Skill Loader
**Owner:** TBD
**Duration:** 2 days
**Depends:** 4.1

- [x] Implement `LoadSkills()` function
- [x] Parse SKILL.md frontmatter
- [x] Load from multiple sources
- [x] Implement `ValidateSkillName()`
- [x] Deduplicate by priority

**Files:**
- `internal/skills/loader.go`
- `internal/skills/validate.go`

**Reference:** `/Users/timmy/PycharmProjects/ShanClaw/internal/skills/loader.go`

**Acceptance Criteria:**
- Can parse SKILL.md with frontmatter
- Sources loaded in correct priority
- Invalid skills rejected with error

---

### Task 4.3: Implement use_skill Tool
**Owner:** TBD
**Duration:** 1 day
**Depends:** 4.2

- [x] Create `UseSkillTool` implementing Tool interface
- [x] Load skill by name
- [x] Return skill prompt for injection
- [x] Register in tool registry

**Files:**
- `internal/tools/use_skill.go` (new)

**Acceptance Criteria:**
- Tool registered and callable
- Skill prompt returned correctly
- Error handling for missing skills

---

### Task 4.4: Skills Tests
**Owner:** TBD
**Duration:** 1 day
**Depends:** 4.3

- [x] Unit tests for frontmatter parsing
- [x] Unit tests for skill loading
- [x] Unit tests for name validation
- [x] Integration tests

**Files:**
- `internal/skills/loader_test.go`
- `internal/skills/validate_test.go`

**Acceptance Criteria:**
- >80% test coverage
- Test various frontmatter formats

---

## Phase 5: Self-Update (Week 6)

### Task 5.1: Implement Update Check
**Owner:** TBD
**Duration:** 2 days
**Depends:** 1.2

- [x] Implement `CheckForUpdate()` function
- [x] Query GitHub API for latest release
- [x] Parse semver versions
- [x] Compare versions
- [x] Handle dev build detection

**Files:**
- `internal/update/selfupdate.go`

**Reference:** `/Users/timmy/PycharmProjects/ShanClaw/internal/update/selfupdate.go`

**Acceptance Criteria:**
- Correctly detects newer versions
- Skips dev/non-semver builds
- Handles API errors gracefully

---

### Task 5.2: Implement Update Download
**Owner:** TBD
**Duration:** 2 days
**Depends:** 5.1

- [x] Download release asset
- [ ] Verify checksum (placeholder)
- [ ] Atomic binary replacement (placeholder)
- [ ] Rollback on failure (placeholder)
- [x] Platform detection

**Files:**
- `internal/update/selfupdate.go`

**Acceptance Criteria:**
- Update installs successfully
- Checksum verification works
- Failed updates don't corrupt binary

---

### Task 5.3: CLI Integration
**Owner:** TBD
**Duration:** 1 day
**Depends:** 5.2

- [x] Implement `update` command
- [x] Add `--check` flag
- [x] Implement auto-update on startup (optional)
- [x] Add update config options

**Files:**
- `cmd/update.go`
- `internal/config/config.go`

**Acceptance Criteria:**
- `starclaw update` works
- `starclaw update --check` reports status
- Config options respected

---

### Task 5.4: Update Tests
**Owner:** TBD
**Duration:** 1 day
**Depends:** 5.3

- [x] Unit tests for version comparison
- [ ] Mock tests for GitHub API
- [ ] Integration tests with test release
- [x] Error handling tests

**Files:**
- `internal/update/selfupdate_test.go`

**Acceptance Criteria:**
- >80% test coverage
- Mock server for GitHub API
- Test various failure modes

---

## Phase 6: Integration & Polish (Week 7)

### Task 6.1: End-to-End Testing
**Owner:** TBD
**Duration:** 2 days
**Depends:** All above

- [ ] Test MCP + Agent combination
- [ ] Test Agent + Skills combination
- [ ] Test full feature stack
- [ ] Performance benchmarks

**Files:**
- `tests/e2e/` (new)

**Acceptance Criteria:**
- All features work together
- No performance regressions

---

### Task 6.2: Documentation
**Owner:** TBD
**Duration:** 2 days
**Depends:** All above

- [x] Update README.md with new features
- [x] Document MCP server configuration
- [x] Document agent creation
- [x] Document skills format
- [x] Document update configuration

**Files:**
- `README.md`
- `docs/mcp.md` (new)
- `docs/agents.md` (new)
- `docs/skills.md` (new)

**Acceptance Criteria:**
- Documentation complete and accurate
- Examples provided
- Configuration reference updated

---

### Task 6.3: Final Review
**Owner:** TBD
**Duration:** 1 day
**Depends:** 6.2

- [x] Review all code for quality
- [x] Ensure test coverage >80%
- [x] Check error handling
- [x] Verify backward compatibility
- [x] Run full test suite

**Acceptance Criteria:**
- All tests pass
- No linting errors
- Ready for release

---

## Summary

| Phase | Tasks | Duration | Deliverable |
|-------|-------|----------|-------------|
| Phase 1 | 3 | 1 week | Config + skeleton |
| Phase 2 | 3 | 2 weeks | MCP client |
| Phase 3 | 4 | 1 week | Named agents |
| Phase 4 | 4 | 1 week | Skills system |
| Phase 5 | 4 | 1 week | Self-update |
| Phase 6 | 3 | 1 week | Integration |
| **Total** | **21** | **7 weeks** | **Release ready** |

## Dependencies Graph

```
1.1 Config Extension ─────────────────┐
1.2 Module Skeletons ─────────────────┤
1.3 CLI Commands ─────────────────────┤
                                      ▼
2.1 MCP ClientManager ◄───────────────┤
2.2 MCP Tool Adapter ◄────────────────┤
2.3 MCP Tests ◄───────────────────────┤
                                      ▼
3.1 Agent Loader ◄────────────────────┤
3.2 Config Merge ◄────────────────────┤
3.3 Agent Integration ◄───────────────┤
3.4 Agent Tests ◄─────────────────────┤
                                      ▼
4.1 Skill Registry ◄──────────────────┤
4.2 Skill Loader ◄────────────────────┤
4.3 use_skill Tool ◄──────────────────┤
4.4 Skills Tests ◄────────────────────┤
                                      ▼
5.1 Update Check ◄────────────────────┤
5.2 Update Download ◄─────────────────┤
5.3 Update CLI ◄──────────────────────┤
5.4 Update Tests ◄────────────────────┤
                                      ▼
6.1 E2E Tests ◄───────────────────────┤
6.2 Documentation ◄───────────────────┤
6.3 Final Review ◄────────────────────┘
```
