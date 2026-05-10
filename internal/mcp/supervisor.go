package mcp

import (
	"context"
	"sync"
	"time"
)

// HealthState represents the health status of an MCP server.
type HealthState string

const (
	// HealthHealthy indicates the server responded successfully to a probe.
	HealthHealthy HealthState = "healthy"
	// HealthUnhealthy indicates the server failed to respond to a probe.
	HealthUnhealthy HealthState = "unhealthy"
	// HealthUnknown indicates the server has not been probed yet.
	HealthUnknown HealthState = "unknown"
)

// ProbeFunc is a function that checks the health of an MCP server.
// A nil error indicates the server is healthy.
type ProbeFunc func(ctx context.Context) error

// Supervisor periodically checks the health of MCP servers.
type Supervisor struct {
	mu       sync.RWMutex
	probes   map[string]ProbeFunc
	states   map[string]HealthState
	interval time.Duration
	cancel   context.CancelFunc
	running  bool
}

// NewSupervisor creates a new MCP supervisor with the given check interval.
func NewSupervisor(interval time.Duration) *Supervisor {
	return &Supervisor{
		probes:   make(map[string]ProbeFunc),
		states:   make(map[string]HealthState),
		interval: interval,
	}
}

// Start launches the health check loop in a background goroutine.
// It runs an initial health check immediately, then repeats at the configured interval.
func (s *Supervisor) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	checkCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()

	go s.loop(checkCtx)
}

// Stop stops the health check loop.
func (s *Supervisor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.running = false
}

// RegisterProbe registers a health probe for a named MCP server.
// The initial health state is set to HealthUnknown.
func (s *Supervisor) RegisterProbe(name string, probe ProbeFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.probes[name] = probe
	s.states[name] = HealthUnknown
}

// HealthStates returns a copy of the current health states for all registered servers.
func (s *Supervisor) HealthStates() map[string]HealthState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]HealthState, len(s.states))
	for k, v := range s.states {
		result[k] = v
	}
	return result
}

func (s *Supervisor) loop(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run an initial check immediately.
	s.checkAll(ctx)

	for {
		select {
		case <-ticker.C:
			s.checkAll(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Supervisor) checkAll(ctx context.Context) {
	s.mu.RLock()
	probes := make(map[string]ProbeFunc, len(s.probes))
	for k, v := range s.probes {
		probes[k] = v
	}
	s.mu.RUnlock()

	var wg sync.WaitGroup

	for name, probe := range probes {
		wg.Add(1)
		go func(name string, probe ProbeFunc) {
			defer wg.Done()

			probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			err := probe(probeCtx)

			s.mu.Lock()
			if err != nil {
				s.states[name] = HealthUnhealthy
			} else {
				s.states[name] = HealthHealthy
			}
			s.mu.Unlock()
		}(name, probe)
	}

	wg.Wait()
}
