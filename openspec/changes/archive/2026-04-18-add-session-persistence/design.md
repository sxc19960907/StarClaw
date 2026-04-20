# Design: Session Persistence

## Overview

Session persistence saves conversation history as JSON files, enabling users to resume previous sessions and maintain context across restarts.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   Session Persistence                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │
│  │     CLI     │────▶│   Session   │────▶│    Store    │   │
│  │   Commands  │     │   Manager   │     │             │   │
│  │             │     │             │     │ - Save      │   │
│  │ --resume    │     │ - Current   │     │ - Load      │   │
│  │ --list      │     │ - New       │     │ - List      │   │
│  │             │     │ - Resume    │     │ - Delete    │   │
│  └─────────────┘     └──────┬──────┘     └──────┬──────┘   │
│                             │                    │          │
│                             ▼                    ▼          │
│                      ┌─────────────┐     ┌─────────────┐    │
│                      │   Session   │     │  JSON Files │    │
│                      │   (struct)  │     │             │    │
│                      │             │     │ ~/.starclaw │    │
│                      │ - ID        │     │ /sessions/  │    │
│                      │ - Messages  │     │             │    │
│                      │ - Title     │     │ <id>.json   │    │
│                      │ - CWD       │     │             │    │
│                      │ - CreatedAt │     │ 0600 perms  │    │
│                      └─────────────┘     └─────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Data Model

### Session Structure

```go
// internal/session/session.go
package session

import (
    "time"
    "github.com/starclaw/starclaw/internal/client"
)

// Session represents a conversation session.
type Session struct {
    ID        string           `json:"id"`
    CreatedAt time.Time        `json:"created_at"`
    UpdatedAt time.Time        `json:"updated_at"`
    Title     string           `json:"title"`
    CWD       string           `json:"cwd"`           // Working directory
    Messages  []client.Message `json:"messages"`      // Conversation history
}

// SessionSummary is a lightweight version for listing.
type SessionSummary struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"created_at"`
    MsgCount  int       `json:"msg_count"`
}
```

### JSON File Format

```json
{
  "id": "2026-04-16-10-30-00-abcd1234",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:35:00Z",
  "title": "Refactor database code",
  "cwd": "/home/user/myproject",
  "messages": [
    {
      "role": "user",
      "content": "Help me refactor the database code"
    },
    {
      "role": "assistant",
      "content": "I'll help you refactor..."
    }
  ]
}
```

## Store Implementation

```go
// internal/session/store.go
package session

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"
)

// Store manages session persistence.
type Store struct {
    dir string
}

// NewStore creates a new session store.
func NewStore(dir string) *Store {
    os.MkdirAll(dir, 0700)
    return &Store{dir: dir}
}

// Save persists a session to disk.
func (s *Store) Save(sess *Session) error {
    sess.UpdatedAt = time.Now()
    if sess.CreatedAt.IsZero() {
        sess.CreatedAt = sess.UpdatedAt
    }

    data, err := json.MarshalIndent(sess, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal session: %w", err)
    }

    path := filepath.Join(s.dir, sess.ID+".json")
    if err := os.WriteFile(path, data, 0600); err != nil {
        return fmt.Errorf("write session: %w", err)
    }

    return nil
}

// Load retrieves a session from disk.
func (s *Store) Load(id string) (*Session, error) {
    path := filepath.Join(s.dir, id+".json")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read session: %w", err)
    }

    var sess Session
    if err := json.Unmarshal(data, &sess); err != nil {
        return nil, fmt.Errorf("parse session: %w", err)
    }
    
    return &sess, nil
}

// List returns summaries of all sessions.
func (s *Store) List() ([]SessionSummary, error) {
    entries, err := os.ReadDir(s.dir)
    if err != nil {
        return nil, err
    }

    var summaries []SessionSummary
    for _, e := range entries {
        if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
            continue
        }
        
        id := strings.TrimSuffix(e.Name(), ".json")
        sess, err := s.Load(id)
        if err != nil {
            continue // Skip corrupted files
        }
        
        summaries = append(summaries, SessionSummary{
            ID:        sess.ID,
            Title:     sess.Title,
            CreatedAt: sess.CreatedAt,
            MsgCount:  len(sess.Messages),
        })
    }

    // Sort by creation time (newest first)
    sort.Slice(summaries, func(i, j int) bool {
        return summaries[i].CreatedAt.After(summaries[j].CreatedAt)
    })
    
    return summaries, nil
}

// Delete removes a session.
func (s *Store) Delete(id string) error {
    path := filepath.Join(s.dir, id+".json")
    return os.Remove(path)
}
```

## Manager Implementation

```go
// internal/session/manager.go
package session

import (
    "crypto/rand"
    "encoding/hex"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// Manager provides session lifecycle operations.
type Manager struct {
    mu      sync.Mutex
    store   *Store
    current *Session
}

// NewManager creates a new session manager.
func NewManager(sessionsDir string) *Manager {
    return &Manager{
        store: NewStore(sessionsDir),
    }
}

// NewSession creates a new session and sets it as current.
func (m *Manager) NewSession() *Session {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.current = &Session{
        ID:        generateSessionID(),
        CreatedAt: time.Now(),
        Title:     "New session",
        CWD:       getCWD(),
        Messages:  []client.Message{},
    }
    return m.current
}

// Current returns the current session (may be nil).
func (m *Manager) Current() *Session {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.current
}

// Resume loads a session and sets it as current.
func (m *Manager) Resume(id string) (*Session, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    sess, err := m.store.Load(id)
    if err != nil {
        return nil, err
    }
    
    m.current = sess
    return sess, nil
}

// Save persists the current session.
func (m *Manager) Save() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if m.current == nil {
        return nil
    }
    
    return m.store.Save(m.current)
}

// List returns all session summaries.
func (m *Manager) List() ([]SessionSummary, error) {
    return m.store.List()
}

// Delete removes a session.
func (m *Manager) Delete(id string) error {
    return m.store.Delete(id)
}

// ResumeLatest loads the most recently updated session.
func (m *Manager) ResumeLatest() (*Session, error) {
    summaries, err := m.store.List()
    if err != nil {
        return nil, err
    }
    if len(summaries) == 0 {
        return nil, nil
    }
    
    // Load the most recent (first in sorted list)
    return m.Resume(summaries[0].ID)
}

// Helper functions
func generateSessionID() string {
    // Format: 2026-04-16-10-30-00-<random>
    now := time.Now().UTC()
    random := make([]byte, 4)
    rand.Read(random)
    
    return fmt.Sprintf("%s-%s",
        now.Format("2006-01-02-%H-%M-%S"),
        hex.EncodeToString(random))
}

func getCWD() string {
    cwd, _ := os.Getwd()
    return cwd
}
```

## Agent Loop Integration

### Modified Agent Loop

```go
// internal/agent/loop.go

type AgentLoop struct {
    // ... existing fields ...
    session     *session.Session
    sessionMgr  *session.Manager
}

// SetSession sets the current session.
func (a *AgentLoop) SetSession(sess *session.Session) {
    a.session = sess
}

// SetSessionManager sets the session manager for auto-save.
func (a *AgentLoop) SetSessionManager(mgr *session.Manager) {
    a.sessionMgr = mgr
}

// Run executes the agent loop with session support.
func (a *AgentLoop) Run(ctx context.Context, query string) (*client.Response, error) {
    // Initialize or resume session
    messages := []client.Message{}
    if a.session != nil {
        messages = append(messages, a.session.Messages...)
    }
    messages = append(messages, client.Message{Role: "user", Content: query})
    
    for i := 0; i < a.maxIter; i++ {
        // ... existing loop logic ...
        
        // After successful response, update session
        if a.session != nil {
            a.session.Messages = messages
            a.session.UpdatedAt = time.Now()
            
            // Auto-save after each turn (best effort)
            if a.sessionMgr != nil {
                a.sessionMgr.Save()
            }
        }
    }
    
    return nil, fmt.Errorf("reached maximum iterations (%d)", a.maxIter)
}
```

## CLI Integration

### New Commands

```go
// cmd/root.go additions

var resumeSession string
var listSessions bool

func init() {
    rootCmd.PersistentFlags().StringVar(&resumeSession, "resume", "", "Resume a previous session by ID")
    rootCmd.PersistentFlags().BoolVar(&listSessions, "list-sessions", false, "List all saved sessions")
}

// In command execution:
func runWithSessions() {
    sessionsDir := filepath.Join(config.StarclawDir(), "sessions")
    sessionMgr := session.NewManager(sessionsDir)
    
    // Handle --list-sessions
    if listSessions {
        summaries, _ := sessionMgr.List()
        for _, s := range summaries {
            fmt.Printf("%s  %s  (%d messages)  %s\n", 
                s.ID, s.Title, s.MsgCount, s.CreatedAt.Format("2006-01-02"))
        }
        return
    }
    
    // Determine session
    var sess *session.Session
    if resumeSession != "" {
        sess, _ = sessionMgr.Resume(resumeSession)
    } else {
        sess = sessionMgr.NewSession()
    }
    
    // Pass to agent loop
    agentLoop.SetSession(sess)
    agentLoop.SetSessionManager(sessionMgr)
    
    // Run...
    
    // Save on exit
    defer sessionMgr.Save()
}
```

## Session Title Generation

Simple heuristic for generating session titles from first message:

```go
// internal/session/title.go
package session

import (
    "strings"
    "unicode"
)

// GenerateTitle creates a title from the first user message.
func GenerateTitle(firstMessage string) string {
    // Truncate to 50 chars
    title := firstMessage
    if len(title) > 50 {
        title = title[:50]
    }
    
    // Clean up whitespace
    title = strings.TrimSpace(title)
    title = strings.Join(strings.Fields(title), " ")
    
    // If empty, use default
    if title == "" {
        title = "New session"
    }
    
    return title
}
```

## Testing Strategy

### Unit Tests

```go
// internal/session/store_test.go

func TestStore_SaveLoad(t *testing.T) {
    tmpDir := t.TempDir()
    store := NewStore(tmpDir)
    
    sess := &Session{
        ID:      "test-session",
        Title:   "Test",
        CWD:     "/tmp",
        Messages: []client.Message{
            {Role: "user", Content: "Hello"},
        },
    }
    
    err := store.Save(sess)
    require.NoError(t, err)
    
    loaded, err := store.Load("test-session")
    require.NoError(t, err)
    assert.Equal(t, sess.Title, loaded.Title)
    assert.Len(t, loaded.Messages, 1)
}

func TestStore_List(t *testing.T) {
    tmpDir := t.TempDir()
    store := NewStore(tmpDir)
    
    // Create multiple sessions
    for i := 0; i < 3; i++ {
        store.Save(&Session{
            ID:    fmt.Sprintf("session-%d", i),
            Title: fmt.Sprintf("Session %d", i),
        })
    }
    
    summaries, err := store.List()
    require.NoError(t, err)
    assert.Len(t, summaries, 3)
}
```

```go
// internal/session/manager_test.go

func TestManager_NewSession(t *testing.T) {
    tmpDir := t.TempDir()
    mgr := NewManager(tmpDir)
    
    sess := mgr.NewSession()
    require.NotNil(t, sess)
    assert.NotEmpty(t, sess.ID)
    assert.NotEmpty(t, sess.CWD)
}

func TestManager_Resume(t *testing.T) {
    tmpDir := t.TempDir()
    mgr := NewManager(tmpDir)
    
    // Create and save a session
    sess := mgr.NewSession()
    sess.Messages = []client.Message{{Role: "user", Content: "Test"}}
    mgr.Save()
    
    // Create new manager and resume
    mgr2 := NewManager(tmpDir)
    resumed, err := mgr2.Resume(sess.ID)
    
    require.NoError(t, err)
    assert.Equal(t, sess.ID, resumed.ID)
    assert.Len(t, resumed.Messages, 1)
}
```

### Integration Tests

```go
// tests/session_test.go

func TestSession_FullFlow(t *testing.T) {
    // Create session, add messages, save, resume
    // Verify messages are preserved
}
```

## Storage Location

```
~/.starclaw/
├── config.yaml
├── sessions/
│   ├── 2026-04-16-10-30-00-abcd1234.json
│   ├── 2026-04-16-11-00-00-efgh5678.json
│   └── 2026-04-16-12-15-00-ijkl9012.json
└── logs/
    └── audit.log
```

## Security Considerations

- Session files: 0600 permissions
- Sessions directory: 0700 permissions
- Sensitive content in messages may need redaction for search (Phase 2)
- No session encryption in Phase 1 (relies on filesystem permissions)

## Future Enhancements

- Session search (FTS5 or grep-based)
- Session expiration/cleanup
- Session encryption
- Session export/import
- Cloud sync
- Session sharing
