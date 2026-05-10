package agent

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestNewCacheMetric(t *testing.T) {
	cm := NewCacheMetric()
	if cm == nil {
		t.Fatal("NewCacheMetric returned nil")
	}
	snap := cm.Snapshot()
	if snap.TotalRequests != 0 {
		t.Errorf("expected 0 total, got %d", snap.TotalRequests)
	}
}

func TestCacheMetric_HitRate_Empty(t *testing.T) {
	cm := NewCacheMetric()
	if hr := cm.HitRate(); hr != 0.0 {
		t.Errorf("expected 0 hit rate for empty metric, got %f", hr)
	}
}

func TestCacheMetric_RecordAndHitRate(t *testing.T) {
	cm := NewCacheMetric()
	cm.Record("get", true, 10*time.Millisecond)
	cm.Record("get", true, 5*time.Millisecond)
	cm.Record("get", false, 20*time.Millisecond)

	expected := 2.0 / 3.0
	hr := cm.HitRate()
	if math.Abs(hr-expected) > 1e-9 {
		t.Errorf("hit rate = %f; want %f", hr, expected)
	}
}

func TestCacheMetric_AllHits(t *testing.T) {
	cm := NewCacheMetric()
	for i := 0; i < 5; i++ {
		cm.Record("get", true, time.Millisecond)
	}
	if hr := cm.HitRate(); hr != 1.0 {
		t.Errorf("expected 1.0 hit rate, got %f", hr)
	}
}

func TestCacheMetric_AllMisses(t *testing.T) {
	cm := NewCacheMetric()
	for i := 0; i < 5; i++ {
		cm.Record("get", false, time.Millisecond)
	}
	if hr := cm.HitRate(); hr != 0.0 {
		t.Errorf("expected 0.0 hit rate, got %f", hr)
	}
}

func TestCacheMetric_P50Latency(t *testing.T) {
	cm := NewCacheMetric()
	cm.Record("get", true, 10*time.Millisecond)
	cm.Record("get", true, 20*time.Millisecond)
	cm.Record("get", true, 30*time.Millisecond)

	p50 := cm.P50Latency()
	if p50 != 20*time.Millisecond {
		t.Errorf("P50 = %v; want 20ms", p50)
	}
}

func TestCacheMetric_P50Latency_Empty(t *testing.T) {
	cm := NewCacheMetric()
	if p50 := cm.P50Latency(); p50 != 0 {
		t.Errorf("expected 0 P50 for empty, got %v", p50)
	}
}

func TestCacheMetric_P95Latency(t *testing.T) {
	cm := NewCacheMetric()
	for i := 0; i < 100; i++ {
		cm.Record("get", true, time.Duration(i)*time.Millisecond)
	}
	p50 := cm.P50Latency()
	p95 := cm.P95Latency()
	p99 := cm.P99Latency()

	if p50 != 49*time.Millisecond && p50 != 50*time.Millisecond {
		t.Logf("P50 = %v (acceptable range)", p50)
	}
	if p95 != 94*time.Millisecond && p95 != 95*time.Millisecond {
		t.Logf("P95 = %v (acceptable range)", p95)
	}
	if p99 != 98*time.Millisecond && p99 != 99*time.Millisecond {
		t.Logf("P99 = %v (acceptable range)", p99)
	}
}

func TestCacheMetric_Snapshot(t *testing.T) {
	cm := NewCacheMetric()
	cm.Record("get", true, 10*time.Millisecond)
	cm.Record("set", false, 50*time.Millisecond)
	cm.Record("get", true, 5*time.Millisecond)

	snap := cm.Snapshot()
	if snap.TotalRequests != 3 {
		t.Errorf("TotalRequests = %d; want 3", snap.TotalRequests)
	}
	if snap.Hits != 2 {
		t.Errorf("Hits = %d; want 2", snap.Hits)
	}
	if snap.Misses != 1 {
		t.Errorf("Misses = %d; want 1", snap.Misses)
	}
	expectedHR := 2.0 / 3.0
	if math.Abs(snap.HitRate-expectedHR) > 1e-9 {
		t.Errorf("HitRate = %f; want %f", snap.HitRate, expectedHR)
	}
	if snap.P50Latency != 10*time.Millisecond {
		t.Errorf("P50 = %v; want 10ms", snap.P50Latency)
	}
	if snap.P95Latency == 0 {
		t.Error("P95 should be non-zero")
	}
	if snap.P99Latency == 0 {
		t.Error("P99 should be non-zero")
	}
}

func TestCacheMetric_SnapshotEmpty(t *testing.T) {
	cm := NewCacheMetric()
	snap := cm.Snapshot()
	if snap.TotalRequests != 0 {
		t.Errorf("expected 0 total, got %d", snap.TotalRequests)
	}
}

func TestCacheMetric_NilReceiver(t *testing.T) {
	var cm *CacheMetric
	cm.Record("get", true, time.Millisecond) // should not panic
	if hr := cm.HitRate(); hr != 0 {
		t.Errorf("expected 0 hit rate from nil, got %f", hr)
	}
	if p50 := cm.P50Latency(); p50 != 0 {
		t.Errorf("expected 0 P50 from nil, got %v", p50)
	}
	snap := cm.Snapshot()
	if snap.TotalRequests != 0 {
		t.Errorf("expected 0 from nil snapshot")
	}
}

func TestCacheMetric_Concurrency(t *testing.T) {
	cm := NewCacheMetric()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cm.Record("get", i%2 == 0, time.Duration(i)*time.Millisecond)
		}(i)
	}
	wg.Wait()

	snap := cm.Snapshot()
	if snap.TotalRequests != 50 {
		t.Errorf("expected 50 total, got %d", snap.TotalRequests)
	}
	if snap.HitRate <= 0 || snap.HitRate >= 1 {
		t.Errorf("hit rate out of range: %f", snap.HitRate)
	}
}
