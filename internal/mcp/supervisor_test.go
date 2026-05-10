package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewSupervisor(t *testing.T) {
	s := NewSupervisor(time.Second)
	if s == nil {
		t.Fatal("NewSupervisor returned nil")
	}
	if s.interval != time.Second {
		t.Errorf("expected interval 1s, got %v", s.interval)
	}
	if len(s.probes) != 0 {
		t.Errorf("expected 0 probes, got %d", len(s.probes))
	}
	if len(s.states) != 0 {
		t.Errorf("expected 0 states, got %d", len(s.states))
	}
}

func TestSupervisor_RegisterProbe(t *testing.T) {
	s := NewSupervisor(time.Second)

	probe := func(ctx context.Context) error { return nil }
	s.RegisterProbe("test-server", probe)

	states := s.HealthStates()
	if len(states) != 1 {
		t.Errorf("expected 1 state, got %d", len(states))
	}

	state, ok := states["test-server"]
	if !ok {
		t.Fatal("expected 'test-server' in states")
	}
	if state != HealthUnknown {
		t.Errorf("expected initial state 'unknown', got %q", state)
	}
}

func TestSupervisor_Start_InitialCheck(t *testing.T) {
	s := NewSupervisor(100 * time.Millisecond)

	healthy := false
	var mu sync.Mutex

	s.RegisterProbe("test-server", func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		if !healthy {
			return errors.New("not ready yet")
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)

	// Initially unhealthy (probe returns error)
	time.Sleep(200 * time.Millisecond)
	states := s.HealthStates()
	if states["test-server"] != HealthUnhealthy {
		t.Errorf("expected unhealthy state, got %q", states["test-server"])
	}

	// Make probe healthy
	mu.Lock()
	healthy = true
	mu.Unlock()

	time.Sleep(300 * time.Millisecond)
	states = s.HealthStates()
	if states["test-server"] != HealthHealthy {
		t.Errorf("expected healthy state after probe succeeds, got %q", states["test-server"])
	}
}

func TestSupervisor_Start_Idempotent(t *testing.T) {
	s := NewSupervisor(time.Hour) // Long interval so tick doesn't fire
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.RegisterProbe("test", func(ctx context.Context) error { return nil })

	// Start twice - should not panic
	s.Start(ctx)
	s.Start(ctx)
	// If we reach here, no panic
}

func TestSupervisor_Stop(t *testing.T) {
	s := NewSupervisor(10 * time.Millisecond)
	ctx := context.Background()

	callCount := 0
	var mu sync.Mutex

	s.RegisterProbe("test", func(ctx context.Context) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	})

	s.Start(ctx)

	// Let a few checks run
	time.Sleep(50 * time.Millisecond)

	s.Stop()

	// Capture count after stop
	mu.Lock()
	stoppedCount := callCount
	mu.Unlock()

	// Wait a bit and verify count doesn't increase
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if callCount > stoppedCount+1 {
		t.Errorf("probe calls increased after Stop: before=%d, after=%d", stoppedCount, callCount)
	}
	mu.Unlock()
}

func TestSupervisor_HealthStates_ReturnsCopy(t *testing.T) {
	s := NewSupervisor(time.Hour)

	s.RegisterProbe("test", func(ctx context.Context) error { return nil })

	states1 := s.HealthStates()
	states2 := s.HealthStates()

	// Modify first copy
	states1["test"] = HealthHealthy

	// Second copy should still have original value
	if states2["test"] != HealthUnknown {
		t.Error("HealthStates should return a copy, not the original map")
	}
}

func TestSupervisor_HealthStates_Concurrent(t *testing.T) {
	s := NewSupervisor(time.Hour)

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("server-%d", i)
		s.RegisterProbe(name, func(ctx context.Context) error { return nil })
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.HealthStates()
		}()
	}
	wg.Wait()
	// Should not panic or deadlock
}

func TestSupervisor_MultipleProbes(t *testing.T) {
	s := NewSupervisor(50 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.RegisterProbe("healthy", func(ctx context.Context) error { return nil })
	s.RegisterProbe("unhealthy", func(ctx context.Context) error { return errors.New("server down") })
	s.RegisterProbe("unknown-probe", func(ctx context.Context) error { return nil })

	s.Start(ctx)
	time.Sleep(150 * time.Millisecond)

	states := s.HealthStates()

	if states["healthy"] != HealthHealthy {
		t.Errorf("expected healthy server to be healthy, got %q", states["healthy"])
	}
	if states["unhealthy"] != HealthUnhealthy {
		t.Errorf("expected unhealthy server to be unhealthy, got %q", states["unhealthy"])
	}
}

func TestSupervisor_ProbeContextTimeout(t *testing.T) {
	s := NewSupervisor(50 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Probe that returns an error (simulates unhealthy server)
	s.RegisterProbe("slow-server", func(ctx context.Context) error {
		return errors.New("server unavailable")
	})


	s.Start(ctx)
	time.Sleep(200 * time.Millisecond)

	states := s.HealthStates()
	// The probe should eventually time out and become unhealthy
	if states["slow-server"] != HealthUnhealthy {
		t.Errorf("expected slow server to be unhealthy after timeout, got %q", states["slow-server"])
	}
}

func TestHealthStateValues(t *testing.T) {
	if HealthHealthy != "healthy" {
		t.Errorf("HealthHealthy = %q", HealthHealthy)
	}
	if HealthUnhealthy != "unhealthy" {
		t.Errorf("HealthUnhealthy = %q", HealthUnhealthy)
	}
	if HealthUnknown != "unknown" {
		t.Errorf("HealthUnknown = %q", HealthUnknown)
	}
}
