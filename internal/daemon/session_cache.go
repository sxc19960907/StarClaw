package daemon

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/session"
)

// ErrRouteActive is returned when a route is already executing.
var ErrRouteActive = errors.New("route is active")

// ErrSessionChanged is returned when the session ID doesn't match.
var ErrSessionChanged = errors.New("session changed")

// SessionCache lazily creates and caches session.Manager instances.
type SessionCache struct {
	mu          sync.Mutex
	routes      map[string]*routeEntry
	managers    map[string]*session.Manager
	starclawDir string
}

// routeEntry holds per-route state including a lock, cancel function, and
// a done channel that is closed when the run completes.
type routeEntry struct {
	mu            sync.Mutex
	cancel        context.CancelFunc
	done          chan struct{}
	cancelPending bool
	lastAccess    time.Time
}

// NewSessionCache creates a new SessionCache backed by starclawDir.
func NewSessionCache(starclawDir string) *SessionCache {
	return &SessionCache{
		routes:      make(map[string]*routeEntry),
		managers:    make(map[string]*session.Manager),
		starclawDir: starclawDir,
	}
}

// GetOrCreate returns a session.Manager for the given agent name,
// creating one if necessary. The agent's sessions are stored under
// <starclawDir>/sessions/<agent>.
func (sc *SessionCache) GetOrCreate(agent string) *session.Manager {
	return sc.GetOrCreateManager(sc.SessionsDir(agent))
}

// GetOrCreateManager returns a session.Manager for the given sessions
// directory, creating one if necessary.
func (sc *SessionCache) GetOrCreateManager(sessionsDir string) *session.Manager {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if mgr, ok := sc.managers[sessionsDir]; ok {
		return mgr
	}

	mgr := newManager(sessionsDir)
	sc.managers[sessionsDir] = mgr
	return mgr
}

// SessionsDir returns the sessions directory for the given agent.
// An empty agent name returns <starclawDir>/sessions.
func (sc *SessionCache) SessionsDir(agent string) string {
	base := filepath.Join(sc.starclawDir, "sessions")
	if agent == "" {
		return base
	}
	return filepath.Join(base, agent)
}

// ResolveLatestSession returns the latest session ID and its messages for
// a route. It uses TryLock to avoid reading session state during mutation.
// Returns ErrRouteActive if a run is in progress.
func (sc *SessionCache) ResolveLatestSession(routeKey, sessionsDir string) (sessionID string, messages []client.Message, err error) {
	if err := ensureSubDir(sc.starclawDir, sessionsDir); err != nil {
		return "", nil, err
	}

	entry := sc.getOrCreateEntry(routeKey)

	if !entry.mu.TryLock() {
		return "", nil, ErrRouteActive
	}
	defer entry.mu.Unlock()

	mgr := sc.GetOrCreateManager(sessionsDir)
	sess, err := mgr.ResumeLatest()
	if err != nil {
		return "", nil, err
	}
	if sess == nil {
		return "", nil, nil
	}
	return sess.ID, sess.Messages, nil
}

// AppendToSession appends messages to the latest session for a route.
// Returns ErrRouteActive if a run is in progress.
// Returns ErrSessionChanged if the session ID doesn't match.
func (sc *SessionCache) AppendToSession(routeKey, sessionsDir, expectedSessionID string, msgs []client.Message) error {
	if err := ensureSubDir(sc.starclawDir, sessionsDir); err != nil {
		return err
	}

	entry := sc.getOrCreateEntry(routeKey)

	if !entry.mu.TryLock() {
		return ErrRouteActive
	}
	defer entry.mu.Unlock()

	mgr := sc.GetOrCreateManager(sessionsDir)
	sess, err := mgr.ResumeLatest()
	if err != nil {
		return err
	}
	if sess == nil {
		return errors.New("no sessions available")
	}
	if sess.ID != expectedSessionID {
		return ErrSessionChanged
	}

	sess.Messages = append(sess.Messages, msgs...)
	return mgr.Save()
}

// LockRoute acquires the per-route lock, blocking until available.
// Returns the routeEntry for the given key.
func (sc *SessionCache) LockRoute(key string) *routeEntry {
	entry := sc.getOrCreateEntry(key)
	entry.mu.Lock()

	// Set up fresh cancel/done for this run.
	_, cancel := context.WithCancel(context.Background())
	entry.cancel = cancel
	entry.done = make(chan struct{})
	entry.cancelPending = false

	return entry
}

// UnlockRoute releases the per-route lock for the given key.
func (sc *SessionCache) UnlockRoute(key string) {
	sc.mu.Lock()
	entry := sc.routes[key]
	sc.mu.Unlock()
	if entry == nil {
		return
	}

	// Signal completion. Safely handle if already closed (e.g. from a
	// concurrent or deferred call).
	select {
	case <-entry.done:
	default:
		close(entry.done)
	}
	entry.lastAccess = time.Now()
	entry.mu.Unlock()
}

// getOrCreateEntry returns the routeEntry for key, creating one if it
// does not exist yet. This is safe for concurrent use.
func (sc *SessionCache) getOrCreateEntry(key string) *routeEntry {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	entry, ok := sc.routes[key]
	if !ok {
		entry = &routeEntry{
			done: make(chan struct{}),
		}
		close(entry.done) // Not running initially.
		sc.routes[key] = entry
	}
	return entry
}

// newManager creates a session.Manager for the given directory, ensuring
// at least one session exists.
func newManager(sessionsDir string) *session.Manager {
	mgr := session.NewManager(sessionsDir)
	sess, err := mgr.ResumeLatest()
	if err != nil || sess == nil {
		mgr.NewSession()
	}
	return mgr
}

// ensureSubDir checks that child is a subdirectory of parent (or equal to it)
// after resolving both to absolute paths.
func ensureSubDir(parent, child string) error {
	absParent, err := filepath.Abs(parent)
	if err != nil {
		return fmt.Errorf("resolve parent path: %w", err)
	}
	absChild, err := filepath.Abs(child)
	if err != nil {
		return fmt.Errorf("resolve child path: %w", err)
	}
	// Use filepath.Rel to securely check parent-child relationship.
	rel, err := filepath.Rel(absParent, absChild)
	if err != nil {
		return fmt.Errorf("relate paths: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path %q is outside %q", child, parent)
	}
	return nil
}
