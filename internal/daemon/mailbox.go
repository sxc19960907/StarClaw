package daemon

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultMailboxCapacity = 100

const (
	QueuePriorityHigh   = 1
	QueuePriorityNormal = 5
	QueuePriorityLow    = 9
)

const (
	QueueStatusQueued       = "queued"
	QueueStatusClaimed      = "claimed"
	QueueStatusAcknowledged = "acknowledged"
	QueueStatusReleased     = "released"
)

var ErrMailboxFull = errors.New("mailbox full")

type QueuedMessage struct {
	ID         string            `json:"id"`
	RouteKey   string            `json:"route_key"`
	SessionID  string            `json:"session_id,omitempty"`
	Source     string            `json:"source,omitempty"`
	ExternalID string            `json:"external_id,omitempty"`
	Sender     string            `json:"sender,omitempty"`
	Agent      string            `json:"agent,omitempty"`
	Text       string            `json:"text"`
	Priority   int               `json:"priority"`
	Status     string            `json:"status"`
	ClaimID    string            `json:"claim_id,omitempty"`
	Attempt    int               `json:"attempt"`
	Duplicate  bool              `json:"duplicate,omitempty"`
	EnqueuedAt time.Time         `json:"enqueued_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type MailboxStore struct {
	mu       sync.RWMutex
	capacity int
	order    []string
	items    map[string]*QueuedMessage
	dedup    map[string]string
}

func NewMailboxStore(capacity int) *MailboxStore {
	if capacity <= 0 {
		capacity = defaultMailboxCapacity
	}
	return &MailboxStore{
		capacity: capacity,
		items:    make(map[string]*QueuedMessage),
		dedup:    make(map[string]string),
	}
}

func (s *MailboxStore) Enqueue(msg QueuedMessage) (QueuedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg.RouteKey = strings.TrimSpace(msg.RouteKey)
	msg.SessionID = strings.TrimSpace(msg.SessionID)
	if msg.RouteKey == "" && msg.SessionID != "" {
		msg.RouteKey = "session:" + msg.SessionID
	}
	msg.Source = strings.TrimSpace(msg.Source)
	msg.ExternalID = strings.TrimSpace(msg.ExternalID)
	msg.Sender = strings.TrimSpace(msg.Sender)
	msg.Agent = strings.TrimSpace(msg.Agent)
	msg.Text = strings.TrimSpace(msg.Text)
	msg.Metadata = cleanQueueMetadata(msg.Metadata)
	if msg.Priority == 0 {
		msg.Priority = QueuePriorityNormal
	}

	if key := queueDedupKey(msg); key != "" {
		if existingID := s.dedup[key]; existingID != "" {
			existing := cloneQueuedMessage(s.items[existingID])
			existing.Duplicate = true
			return existing, nil
		}
	}
	if s.queuedCountLocked(msg.RouteKey) >= s.capacity {
		return QueuedMessage{}, ErrMailboxFull
	}
	now := time.Now()
	if msg.ID == "" {
		msg.ID = "queue_" + generateQueueID()
	}
	if msg.Status == "" {
		msg.Status = QueueStatusQueued
	}
	if msg.EnqueuedAt.IsZero() {
		msg.EnqueuedAt = now
	}
	msg.UpdatedAt = now
	if msg.Attempt <= 0 {
		msg.Attempt = 1
	}

	s.items[msg.ID] = &msg
	s.order = append(s.order, msg.ID)
	if key := queueDedupKey(msg); key != "" {
		s.dedup[key] = msg.ID
	}
	return cloneQueuedMessage(&msg), nil
}

func (s *MailboxStore) List(routeKey string) []QueuedMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	routeKey = strings.TrimSpace(routeKey)
	out := make([]QueuedMessage, 0, len(s.order))
	for _, id := range s.order {
		item := s.items[id]
		if item == nil {
			continue
		}
		if routeKey != "" && item.RouteKey != routeKey {
			continue
		}
		out = append(out, cloneQueuedMessage(item))
	}
	sortQueuedMessages(out)
	return out
}

func (s *MailboxStore) Get(id string) (QueuedMessage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item := s.items[strings.TrimSpace(id)]
	if item == nil {
		return QueuedMessage{}, false
	}
	return cloneQueuedMessage(item), true
}

func (s *MailboxStore) Claim(routeKey string, limit int) ([]QueuedMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	routeKey = strings.TrimSpace(routeKey)
	if routeKey == "" {
		return nil, errors.New("route_key is required")
	}
	if limit <= 0 {
		limit = 1
	}
	now := time.Now()
	available := make([]QueuedMessage, 0)
	for _, id := range s.order {
		item := s.items[id]
		if item == nil || item.RouteKey != routeKey || item.Status != QueueStatusQueued {
			continue
		}
		available = append(available, cloneQueuedMessage(item))
	}
	sortQueuedMessages(available)
	if limit > len(available) {
		limit = len(available)
	}
	out := make([]QueuedMessage, 0, limit)
	for _, item := range available[:limit] {
		stored := s.items[item.ID]
		if stored == nil {
			continue
		}
		stored.Status = QueueStatusClaimed
		stored.ClaimID = "claim_" + generateQueueID()
		stored.UpdatedAt = now
		out = append(out, cloneQueuedMessage(stored))
	}
	return out, nil
}

func (s *MailboxStore) Ack(id, claimID string) bool {
	return s.transitionClaimed(id, claimID, QueueStatusAcknowledged, false)
}

func (s *MailboxStore) Release(id, claimID string) bool {
	return s.transitionClaimed(id, claimID, QueueStatusQueued, true)
}

func (s *MailboxStore) transitionClaimed(id, claimID, status string, released bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := s.items[strings.TrimSpace(id)]
	if item == nil || item.Status != QueueStatusClaimed || item.ClaimID != strings.TrimSpace(claimID) {
		return false
	}
	item.Status = status
	item.ClaimID = ""
	item.UpdatedAt = time.Now()
	if released {
		item.Attempt++
	}
	return true
}

func (s *MailboxStore) queuedCountLocked(routeKey string) int {
	count := 0
	for _, item := range s.items {
		if item.RouteKey == routeKey && item.Status != QueueStatusAcknowledged {
			count++
		}
	}
	return count
}

func sortQueuedMessages(messages []QueuedMessage) {
	sort.SliceStable(messages, func(i, j int) bool {
		left, right := messages[i], messages[j]
		if left.Priority != right.Priority {
			return left.Priority < right.Priority
		}
		if !left.EnqueuedAt.Equal(right.EnqueuedAt) {
			return left.EnqueuedAt.Before(right.EnqueuedAt)
		}
		return left.ID < right.ID
	})
}

func queueDedupKey(msg QueuedMessage) string {
	if msg.Source == "" || msg.ExternalID == "" || msg.RouteKey == "" {
		return ""
	}
	return msg.RouteKey + ":" + msg.Source + ":" + msg.ExternalID
}

func cloneQueuedMessage(item *QueuedMessage) QueuedMessage {
	if item == nil {
		return QueuedMessage{}
	}
	out := *item
	if item.Metadata != nil {
		out.Metadata = make(map[string]string, len(item.Metadata))
		for key, value := range item.Metadata {
			out.Metadata[key] = value
		}
	}
	return out
}

func cleanQueueMetadata(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" || looksSensitive(key) || looksSensitive(value) {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func generateQueueID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(b[:])
}
