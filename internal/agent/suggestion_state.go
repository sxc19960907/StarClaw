package agent

import (
	"strings"
	"sync"
	"time"
)

// Suggestion is the latest UI-facing follow-up suggestion for a session.
type Suggestion struct {
	Text        string
	SuggestedAt time.Time
	AcceptedAt  *time.Time
}

// SuggestionState stores the latest suggestion per session.
type SuggestionState struct {
	mu    sync.RWMutex
	items map[string]*Suggestion
}

func NewSuggestionState() *SuggestionState {
	return &SuggestionState{items: make(map[string]*Suggestion)}
}

func (s *SuggestionState) Set(sessionID, text string, at time.Time) {
	if s == nil || strings.TrimSpace(sessionID) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[sessionID] = &Suggestion{Text: text, SuggestedAt: at}
}

func (s *SuggestionState) Get(sessionID string) (Suggestion, bool) {
	if s == nil || strings.TrimSpace(sessionID) == "" {
		return Suggestion{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[sessionID]
	if !ok {
		return Suggestion{}, false
	}
	return *item, true
}

func (s *SuggestionState) Clear(sessionID string) {
	if s == nil || strings.TrimSpace(sessionID) == "" {
		return
	}
	s.mu.Lock()
	delete(s.items, sessionID)
	s.mu.Unlock()
}

func (s *SuggestionState) MarkAccepted(sessionID string) bool {
	if s == nil || strings.TrimSpace(sessionID) == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[sessionID]
	if !ok {
		return false
	}
	now := time.Now()
	item.AcceptedAt = &now
	return true
}
