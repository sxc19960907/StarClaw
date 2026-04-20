# Specification: Session Persistence

## Overview

| Field | Value |
|-------|-------|
| Feature | Session Persistence |
| Storage | JSON files |
| Location | `~/.starclaw/sessions/` |
| Permissions | 0600 (files), 0700 (directory) |

## Purpose

Save conversation history to enable:
- Resuming previous conversations
- Browsing session history
- Maintaining context across restarts

## Session Structure

### Session Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique session identifier |
| `created_at` | ISO8601 | Creation timestamp |
| `updated_at` | ISO8601 | Last update timestamp |
| `title` | string | Human-readable session title |
| `cwd` | string | Working directory at creation |
| `messages` | array | Conversation message history |

### Message Format

```json
{
  "role": "user|assistant",
  "content": "Message content"
}
```

## Session ID Format

```
YYYY-MM-DD-HH-MM-SS-<random>

Example: 2026-04-16-10-30-00-abcd1234
```

- Date and time in UTC
- Random hex suffix (8 chars) to prevent collisions
- Human-readable and sortable

## Storage Format

### File Path
```
~/.starclaw/sessions/<id>.json
```

### JSON Structure
```json
{
  "id": "2026-04-16-10-30-00-abcd1234",
  "created_at": "2026-04-16T10:30:00Z",
  "updated_at": "2026-04-16T10:35:00Z",
  "title": "Refactor database code",
  "cwd": "/home/user/myproject",
  "messages": [
    {"role": "user", "content": "Help me refactor..."},
    {"role": "assistant", "content": "I'll help you..."}
  ]
}
```

## CLI Commands

### List Sessions

```bash
starclaw --list-sessions
```

Output:
```
2026-04-16-10-30-00-abcd1234  Refactor database code  (12 messages)  2026-04-16
2026-04-16-11-00-00-efgh5678  Check API status        (5 messages)   2026-04-16
2026-04-15-09-00-00-ijkl9012  Write tests             (8 messages)   2026-04-15
```

### Resume Session

```bash
# Resume by ID
starclaw --resume 2026-04-16-10-30-00-abcd1234

# Resume and ask a question
starclaw --resume 2026-04-16-10-30-00-abcd1234 "Continue where we left off"
```

### Interactive Mode

In interactive TUI mode, sessions are automatically:
- Created on first message
- Saved after each turn
- Resumed on restart (optional)

## API

### Store

```go
type Store struct { }

func NewStore(dir string) *Store
func (s *Store) Save(sess *Session) error
func (s *Store) Load(id string) (*Session, error)
func (s *Store) List() ([]SessionSummary, error)
func (s *Store) Delete(id string) error
```

### Manager

```go
type Manager struct { }

func NewManager(dir string) *Manager
func (m *Manager) NewSession() *Session
func (m *Manager) Current() *Session
func (m *Manager) Resume(id string) (*Session, error)
func (m *Manager) ResumeLatest() (*Session, error)
func (m *Manager) Save() error
func (m *Manager) List() ([]SessionSummary, error)
func (m *Manager) Delete(id string) error
```

## Configuration

```yaml
# ~/.starclaw/config.yaml
sessions:
  auto_resume: false  # Auto-resume last session on startup
```

Default: `auto_resume: false` (opt-in)

## Title Generation

Session titles are generated from the first user message:

1. Truncate to 50 characters
2. Normalize whitespace
3. Use "New session" if empty

### Examples

| First Message | Generated Title |
|---------------|-----------------|
| "Help me refactor the database code" | "Help me refactor the database code" |
| "Check API status and fix any issues" | "Check API status and fix any issues" |
| "   " | "New session" |

## Security

### File Permissions

- Session files: `0600` (owner read/write only)
- Sessions directory: `0700` (owner access only)

### Privacy Considerations

- Conversations stored in plain text
- Relies on filesystem permissions for security
- No encryption in Phase 1
- Sensitive content may be in messages

## Error Handling

| Error | Behavior |
|-------|----------|
| Session file corrupted | Skip in list, error on load |
| Session not found | Return error |
| Permission denied | Return error |
| Disk full | Return error |

## Usage Examples

### Example 1: Basic Session Flow

```bash
# Start new session
starclaw "Help me write a Go function"
# -> Session created: 2026-04-16-10-30-00-abcd1234

# Later, list sessions
starclaw --list-sessions

# Resume the session
starclaw --resume 2026-04-16-10-30-00-abcd1234 "Add error handling to that function"
# -> Context preserved, continues conversation
```

### Example 2: Multi-Session Workflow

```bash
# Work on feature A
starclaw "Help design the auth system"

# Work on feature B (new session)
starclaw "Review the database schema"

# List and switch
starclaw --list-sessions
starclaw --resume <auth-session-id> "Now implement the middleware"
```

### Example 3: One-Shot with Resume

```bash
# Start task
starclaw "Find all TODO comments"

# Continue same session from another terminal
starclaw --resume <session-id> "Now fix the TODO in main.go"
```

## Design Rationale

- **JSON files**: Human-readable, easy to debug, portable
- **Plain text**: No dependencies, simple to query with standard tools
- **Timestamp IDs**: Human-readable, sortable, unique
- **Auto-save**: Best-effort persistence, no manual action needed
- **Opt-in resume**: User chooses when to continue vs. start fresh
