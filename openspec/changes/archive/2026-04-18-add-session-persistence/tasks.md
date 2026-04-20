# Tasks: Add Session Persistence

## Task 1: Create Session Model

**ID**: T1  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Create `Session` struct and `SessionSummary` in `internal/session/session.go`.

### Acceptance Criteria
- [ ] `session.go` created
- [ ] `Session` struct with all fields (ID, CreatedAt, UpdatedAt, Title, CWD, Messages)
- [ ] `SessionSummary` struct for listing
- [ ] JSON tags on all fields
- [ ] Package documentation

---

## Task 2: Implement Session Store

**ID**: T2  
**Status**: completed  
**Owner**:  
**Blocked By**: T1

### Description
Create `Store` with CRUD operations.

### Acceptance Criteria
- [ ] `store.go` created with `Store` struct
- [ ] `NewStore(dir)` creates directory with 0700 permissions
- [ ] `Save(sess)` writes JSON with 0600 permissions
- [ ] `Load(id)` reads and parses JSON
- [ ] `List()` returns summaries sorted by date
- [ ] `Delete(id)` removes file
- [ ] Proper error handling for all operations

---

## Task 3: Write Store Unit Tests

**ID**: T3  
**Status**: completed  
**Owner**:  
**Blocked By**: T2

### Description
Create `store_test.go` with comprehensive tests.

### Acceptance Criteria
- [ ] Test `Save()` creates file with correct permissions
- [ ] Test `Load()` returns correct data
- [ ] Test `List()` returns all sessions sorted
- [ ] Test `Delete()` removes file
- [ ] Test error cases (file not found, invalid JSON)
- [ ] All tests use temp directories
- [ ] All tests pass

---

## Task 4: Implement Session Manager

**ID**: T4  
**Status**: completed  
**Owner**:  
**Blocked By**: T2

### Description
Create `Manager` with lifecycle operations.

### Acceptance Criteria
- [ ] `manager.go` created with `Manager` struct
- [ ] `NewManager(dir)` initializes
- [ ] `NewSession()` creates new session with generated ID
- [ ] `Current()` returns current session
- [ ] `Resume(id)` loads and sets as current
- [ ] `Save()` persists current session
- [ ] `List()` delegates to Store
- [ ] `Delete(id)` delegates to Store
- [ ] `ResumeLatest()` loads most recent session
- [ ] Thread-safe (mutex protection)

---

## Task 5: Write Manager Unit Tests

**ID**: T5  
**Status**: completed  
**Owner**:  
**Blocked By**: T4

### Description
Create `manager_test.go` with tests.

### Acceptance Criteria
- [ ] Test `NewSession()` generates unique IDs
- [ ] Test `Current()` returns nil initially
- [ ] Test `Resume()` loads correct session
- [ ] Test `Save()` persists changes
- [ ] Test `ResumeLatest()` finds most recent
- [ ] Test thread safety (concurrent operations)
- [ ] All tests pass

---

## Task 6: Add Session ID Generation

**ID**: T6  
**Status**: completed  
**Owner**:  
**Blocked By**: T4

### Description
Implement `generateSessionID()` helper.

### Acceptance Criteria
- [ ] Format: `YYYY-MM-DD-HH-MM-SS-<random>`
- [ ] Human-readable timestamp
- [ ] Random suffix to avoid collisions
- [ ] UTC time used
- [ ] Unit test for format

---

## Task 7: Add Title Generation

**ID**: T7  
**Status**: completed  
**Owner**:  
**Blocked By**: T1

### Description
Create `GenerateTitle()` for session titles.

### Acceptance Criteria
- [ ] Truncate to 50 chars
- [ ] Clean up whitespace
- [ ] Default "New session" if empty
- [ ] Unit tests for various inputs

---

## Task 8: Integrate with Agent Loop

**ID**: T8  
**Status**: completed  
**Owner**:  
**Blocked By**: T1, T4

### Description
Modify `AgentLoop` to support sessions.

### Acceptance Criteria
- [ ] Add `session` and `sessionMgr` fields to `AgentLoop`
- [ ] Add `SetSession(sess)` method
- [ ] Add `SetSessionManager(mgr)` method
- [ ] Modify `Run()` to load messages from session
- [ ] Modify `Run()` to save messages to session after each turn
- [ ] Update `UpdatedAt` timestamp
- [ ] Call `sessionMgr.Save()` after each turn
- [ ] Handle session being nil (backwards compatible)

---

## Task 9: Add CLI Commands

**ID**: T9  
**Status**: completed  
**Owner**:  
**Blocked By**: T8

### Description
Add `--resume` and `--list-sessions` flags.

### Acceptance Criteria
- [ ] `--resume <id>` resumes specific session
- [ ] `--list-sessions` prints session list
- [ ] List format: `ID  Title  (N messages)  Date`
- [ ] Works for both interactive and one-shot modes
- [ ] Auto-save on graceful exit

---

## Task 10: Write Integration Tests

**ID**: T10  
**Status**: completed  
**Owner**:  
**Blocked By**: T9

### Description
Add integration tests for session flow.

### Acceptance Criteria
- [ ] Test create session, add messages, save
- [ ] Test resume session, messages preserved
- [ ] Test list sessions
- [ ] Test multiple sessions don't interfere
- [ ] Tests pass

---

## Task 11: Add Documentation

**ID**: T11  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Document session commands in README.

### Acceptance Criteria
- [ ] Update README with `--list-sessions` example
- [ ] Update README with `--resume` example
- [ ] Document session storage location
- [ ] Document session file format

---

## Dependencies

```
T1 (Session Model)
  └── T2 (Store)
       └── T3 (Store Tests)
       └── T4 (Manager)
            └── T5 (Manager Tests)
            └── T6 (ID Generation)

T1 ──▶ T7 (Title Gen)

T1 + T4 ──▶ T8 (Agent Integration)
                └── T9 (CLI Commands)
                     └── T10 (Integration Tests)

T11 (Docs) - independent
```

## Estimated Effort

- T1: 30 minutes
- T2: 1 hour
- T3: 1 hour
- T4: 1 hour
- T5: 1 hour
- T6: 30 minutes
- T7: 30 minutes
- T8: 1.5 hours (Agent Loop modification)
- T9: 1 hour
- T10: 1 hour
- T11: 30 minutes

**Total**: ~9.5 hours
