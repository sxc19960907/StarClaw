package daemon

import (
	"sync"

	"github.com/starclaw/starclaw/internal/agent"
)

const defaultSystemEventStoreCap = 20

// SystemEventStore holds bounded, route-scoped queues of daemon-authored
// system events for later turn injection.
type SystemEventStore struct {
	mu     sync.Mutex
	queues map[string][]agent.SystemEvent
	cap    int
}

func NewSystemEventStore(capPerRoute int) *SystemEventStore {
	if capPerRoute <= 0 {
		capPerRoute = defaultSystemEventStoreCap
	}
	return &SystemEventStore{
		queues: make(map[string][]agent.SystemEvent),
		cap:    capPerRoute,
	}
}

func (s *SystemEventStore) Enqueue(routeKey string, ev agent.SystemEvent) {
	if s == nil || routeKey == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	queue := s.queues[routeKey]
	if ev.ContextKey != "" && len(queue) > 0 && queue[len(queue)-1].ContextKey == ev.ContextKey {
		queue[len(queue)-1] = ev
		s.queues[routeKey] = queue
		return
	}
	queue = append(queue, ev)
	if len(queue) > s.cap {
		queue = queue[len(queue)-s.cap:]
	}
	s.queues[routeKey] = queue
}

func (s *SystemEventStore) Drain(routeKey string) []agent.SystemEvent {
	if s == nil || routeKey == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	queue := s.queues[routeKey]
	if len(queue) == 0 {
		return nil
	}
	delete(s.queues, routeKey)
	out := make([]agent.SystemEvent, len(queue))
	copy(out, queue)
	return out
}

func (s *SystemEventStore) Forget(routeKey string) {
	if s == nil || routeKey == "" {
		return
	}
	s.mu.Lock()
	delete(s.queues, routeKey)
	s.mu.Unlock()
}
