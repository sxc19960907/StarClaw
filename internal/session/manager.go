package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
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

// AddTag adds a tag to a session
func (m *Manager) AddTag(sessionID string, tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, err := m.store.Load(sessionID)
	if err != nil {
		return err
	}

	// Avoid duplicate tags
	for _, t := range sess.Tags {
		if t == tag {
			return nil
		}
	}

	sess.Tags = append(sess.Tags, tag)
	return m.store.Save(sess)
}

// RemoveTag removes a tag from a session
func (m *Manager) RemoveTag(sessionID string, tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, err := m.store.Load(sessionID)
	if err != nil {
		return err
	}

	found := false
	for i, t := range sess.Tags {
		if t == tag {
			sess.Tags = append(sess.Tags[:i], sess.Tags[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	return m.store.Save(sess)
}

// SetFavorite sets the favorite flag on a session
func (m *Manager) SetFavorite(sessionID string, fav bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, err := m.store.Load(sessionID)
	if err != nil {
		return err
	}

	sess.Favorite = fav
	return m.store.Save(sess)
}

// SearchByTag returns sessions that have the given tag
func (m *Manager) SearchByTag(tag string) []SessionSummary {
	summaries, err := m.store.List()
	if err != nil {
		return nil
	}

	var result []SessionSummary
	for _, s := range summaries {
		for _, t := range s.Tags {
			if t == tag {
				result = append(result, s)
				break
			}
		}
	}
	return result
}

// ListFavorites returns all favorited sessions
func (m *Manager) ListFavorites() []SessionSummary {
	summaries, err := m.store.List()
	if err != nil {
		return nil
	}

	var result []SessionSummary
	for _, s := range summaries {
		if s.Favorite {
			result = append(result, s)
		}
	}
	return result
}

// Export formats a session as markdown or HTML
func (m *Manager) Export(sessionID string, format string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, err := m.store.Load(sessionID)
	if err != nil {
		return "", err
	}

	switch strings.ToLower(format) {
	case "md", "markdown":
		return ExportMarkdown(sess), nil
	case "html":
		return ExportHTML(sess), nil
	default:
		return "", fmt.Errorf("unsupported export format: %q (use md or html)", format)
	}
}

// generateSessionID creates a human-readable session ID
// Format: YYYY-MM-DD-HH-MM-SS-<random>
func generateSessionID() string {
	now := time.Now().UTC()
	random := make([]byte, 4)
	if _, err := io.ReadFull(rand.Reader, random); err != nil {
		panic(fmt.Errorf("generate session id randomness: %w", err))
	}

	return fmt.Sprintf("%s-%s",
		now.Format("2006-01-02-15-04-05"),
		hex.EncodeToString(random))
}

func getCWD() string {
	cwd, _ := os.Getwd()
	return cwd
}
