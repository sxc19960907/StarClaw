package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/client"
)

// Manager provides session lifecycle operations
type Manager struct {
	mu      sync.Mutex
	store   *Store
	current *Session
}

// NewManager creates a new session manager
func NewManager(sessionsDir string) *Manager {
	return &Manager{
		store: NewStore(sessionsDir),
	}
}

// NewSession creates a new session and sets it as current
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

// Current returns the current session (may be nil)
func (m *Manager) Current() *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current
}

// Resume loads a session and sets it as current
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

// Save persists the current session
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current == nil {
		return nil
	}

	return m.store.Save(m.current)
}

// List returns all session summaries
func (m *Manager) List() ([]SessionSummary, error) {
	return m.store.List()
}

// Delete removes a session
func (m *Manager) Delete(id string) error {
	return m.store.Delete(id)
}

// ResumeLatest loads the most recently updated session
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

// generateSessionID creates a human-readable session ID
// Format: YYYY-MM-DD-HH-MM-SS-<random>
func generateSessionID() string {
	now := time.Now().UTC()
	random := make([]byte, 4)
	rand.Read(random)

	return fmt.Sprintf("%s-%s",
		now.Format("2006-01-02-15-04-05"),
		hex.EncodeToString(random))
}

func getCWD() string {
	cwd, _ := os.Getwd()
	return cwd
}
