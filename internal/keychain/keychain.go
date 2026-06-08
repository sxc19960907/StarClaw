// Package keychain wraps OS credential storage behind a StarClaw-specific
// store boundary. It is intentionally not wired into config loading yet.
package keychain

import (
	"errors"
	"fmt"
)

const (
	// ServiceDaemonAPIKey stores per-user daemon API key entries. Account is the
	// StarClaw/cloud user id, with AccountLegacy reserved for future migration.
	ServiceDaemonAPIKey = "ai.starclaw.daemon.api_key"

	// ServiceDaemonState stores daemon state pointers such as the active user id.
	ServiceDaemonState = "ai.starclaw.daemon.state"

	// AccountCurrentUser is the well-known account under ServiceDaemonState whose
	// value is the active user id.
	AccountCurrentUser = "current_user_id"

	// AccountLegacy is the placeholder account for future config-to-keychain
	// migration when a real user id is not known yet.
	AccountLegacy = "legacy"
)

var (
	// ErrUnsupportedPlatform is returned by NewOSStore on unsupported platforms.
	ErrUnsupportedPlatform = errors.New("keychain: unsupported platform")

	// ErrNotFound is returned by Backend implementations when an entry is absent.
	ErrNotFound = errors.New("keychain: not found")
)

// Backend is the low-level credential store contract. Production uses an OS
// backend where supported; tests use NewMemBackend.
type Backend interface {
	Read(service, account string) (string, error)
	Write(service, account, value string) error
	Delete(service, account string) error
}

// Store is the high-level StarClaw credential store. It composes a Backend with
// daemon-domain operations such as active-user lookup and legacy key rename.
type Store struct {
	backend Backend
}

// NewStore builds a Store from a Backend.
func NewStore(backend Backend) *Store {
	return &Store{backend: backend}
}

// Read returns the raw value at (service, account). Missing entries are treated
// as empty values because callers generally use absence as "not configured".
func (s *Store) Read(service, account string) (string, error) {
	if s == nil || s.backend == nil {
		return "", ErrUnsupportedPlatform
	}
	value, err := s.backend.Read(service, account)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// Write stores value at (service, account), replacing any existing value.
func (s *Store) Write(service, account, value string) error {
	if s == nil || s.backend == nil {
		return ErrUnsupportedPlatform
	}
	return s.backend.Write(service, account, value)
}

// Delete removes (service, account). It is idempotent for absent entries.
func (s *Store) Delete(service, account string) error {
	if s == nil || s.backend == nil {
		return ErrUnsupportedPlatform
	}
	err := s.backend.Delete(service, account)
	if err == nil || errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

// CurrentUserID returns the active user id, or "" when no active user is set.
func (s *Store) CurrentUserID() (string, error) {
	return s.Read(ServiceDaemonState, AccountCurrentUser)
}

// GetAPIKey returns the API key for the active user, or "" when unset.
func (s *Store) GetAPIKey() (string, error) {
	userID, err := s.CurrentUserID()
	if err != nil {
		return "", err
	}
	if userID == "" {
		return "", nil
	}
	return s.Read(ServiceDaemonAPIKey, userID)
}

// GetActiveUserAndKey returns the active user id and its API key.
func (s *Store) GetActiveUserAndKey() (string, string, error) {
	userID, err := s.CurrentUserID()
	if err != nil {
		return "", "", err
	}
	if userID == "" {
		return "", "", nil
	}
	key, err := s.Read(ServiceDaemonAPIKey, userID)
	if err != nil {
		return "", "", err
	}
	return userID, key, nil
}

// SetAPIKey writes an API key for userID, then marks that user as active.
func (s *Store) SetAPIKey(userID, key string) error {
	if s == nil || s.backend == nil {
		return ErrUnsupportedPlatform
	}
	if userID == "" {
		return fmt.Errorf("keychain: SetAPIKey requires non-empty userID")
	}
	if key == "" {
		return fmt.Errorf("keychain: SetAPIKey requires non-empty key")
	}
	if err := s.Write(ServiceDaemonAPIKey, userID, key); err != nil {
		return fmt.Errorf("keychain: write api_key: %w", err)
	}
	if err := s.Write(ServiceDaemonState, AccountCurrentUser, userID); err != nil {
		return fmt.Errorf("keychain: write current_user_id: %w", err)
	}
	return nil
}

// RenameLegacy moves the AccountLegacy key to realUserID and marks realUserID
// active. It is a no-op when no legacy key exists.
func (s *Store) RenameLegacy(realUserID string) error {
	if realUserID == "" {
		return fmt.Errorf("keychain: RenameLegacy requires non-empty realUserID")
	}
	legacy, err := s.Read(ServiceDaemonAPIKey, AccountLegacy)
	if err != nil {
		return err
	}
	if legacy == "" {
		return nil
	}
	if err := s.SetAPIKey(realUserID, legacy); err != nil {
		return err
	}
	return s.Delete(ServiceDaemonAPIKey, AccountLegacy)
}

// DeleteAPIKey removes the API key for the current active user while preserving
// the active-user pointer.
func (s *Store) DeleteAPIKey() error {
	userID, err := s.CurrentUserID()
	if err != nil {
		return err
	}
	if userID == "" {
		return nil
	}
	return s.Delete(ServiceDaemonAPIKey, userID)
}

// ClearActiveUser removes only the active-user pointer.
func (s *Store) ClearActiveUser() error {
	return s.Delete(ServiceDaemonState, AccountCurrentUser)
}
