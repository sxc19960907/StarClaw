package daemon

import (
	"sync"
)

// SessionCWD tracks the current working directory per session in daemon
// memory.  This allows agent sessions to maintain their own working
// directory across turns without persisting to disk.
type SessionCWD struct {
	mu   sync.RWMutex
	cwds map[string]string
}

// NewSessionCWD creates a new SessionCWD with an empty directory map.
func NewSessionCWD() *SessionCWD {
	return &SessionCWD{
		cwds: make(map[string]string),
	}
}

// GetCWD returns the current working directory for the given session.
// Returns an empty string if no directory has been set.
func (s *SessionCWD) GetCWD(sessionID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cwds[sessionID]
}

// SetCWD sets the current working directory for the given session.
func (s *SessionCWD) SetCWD(sessionID, dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cwds[sessionID] = dir
}
