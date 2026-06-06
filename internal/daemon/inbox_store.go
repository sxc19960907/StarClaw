package daemon

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultInboxStoreLimit = 200

type InboxItem struct {
	ID         string            `json:"id"`
	Provider   string            `json:"provider"`
	ExternalID string            `json:"external_id"`
	Sender     string            `json:"sender,omitempty"`
	Text       string            `json:"text"`
	Status     string            `json:"status"`
	Agent      string            `json:"agent,omitempty"`
	RunID      string            `json:"run_id,omitempty"`
	Error      string            `json:"error,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type InboxStore struct {
	mu      sync.RWMutex
	limit   int
	order   []string
	items   map[string]*InboxItem
	byEvent map[string]string
}

func NewInboxStore(limit int) *InboxStore {
	if limit <= 0 {
		limit = defaultInboxStoreLimit
	}
	return &InboxStore{
		limit:   limit,
		items:   make(map[string]*InboxItem),
		byEvent: make(map[string]string),
	}
}

func (s *InboxStore) Upsert(item InboxItem) (InboxItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := inboxEventKey(item.Provider, item.ExternalID)
	if key != "" {
		if existingID := s.byEvent[key]; existingID != "" {
			return cloneInboxItem(s.items[existingID]), true
		}
	}

	now := time.Now()
	if item.ID == "" {
		item.ID = "inbox_" + generateRequestID()
	}
	if item.Status == "" {
		item.Status = "pending"
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	s.items[item.ID] = &item
	s.order = append([]string{item.ID}, s.order...)
	if key != "" {
		s.byEvent[key] = item.ID
	}
	s.trimLocked()
	return cloneInboxItem(&item), false
}

func (s *InboxStore) List() []InboxItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]InboxItem, 0, len(s.order))
	for _, id := range s.order {
		if item := s.items[id]; item != nil {
			items = append(items, cloneInboxItem(item))
		}
	}
	return items
}

func (s *InboxStore) Get(id string) (InboxItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item := s.items[id]
	if item == nil {
		return InboxItem{}, false
	}
	return cloneInboxItem(item), true
}

func (s *InboxStore) Update(id string, update func(*InboxItem) error) (InboxItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := s.items[id]
	if item == nil {
		return InboxItem{}, fmt.Errorf("inbox item not found")
	}
	if err := update(item); err != nil {
		return InboxItem{}, err
	}
	item.UpdatedAt = time.Now()
	return cloneInboxItem(item), nil
}

func (s *InboxStore) trimLocked() {
	for len(s.order) > s.limit {
		last := s.order[len(s.order)-1]
		s.order = s.order[:len(s.order)-1]
		if item := s.items[last]; item != nil {
			delete(s.byEvent, inboxEventKey(item.Provider, item.ExternalID))
		}
		delete(s.items, last)
	}
}

func inboxEventKey(provider, externalID string) string {
	provider = strings.TrimSpace(provider)
	externalID = strings.TrimSpace(externalID)
	if provider == "" || externalID == "" {
		return ""
	}
	return provider + ":" + externalID
}

func cloneInboxItem(item *InboxItem) InboxItem {
	if item == nil {
		return InboxItem{}
	}
	out := *item
	if item.Metadata != nil {
		out.Metadata = make(map[string]string, len(item.Metadata))
		keys := make([]string, 0, len(item.Metadata))
		for key := range item.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			out.Metadata[key] = item.Metadata[key]
		}
	}
	return out
}
