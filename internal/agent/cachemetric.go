package agent

import (
	"sort"
	"sync"
	"time"
)

// MetricSnapshot is a point-in-time summary of cache performance.
type MetricSnapshot struct {
	TotalRequests int
	Hits          int
	Misses        int
	HitRate       float64
	P50Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
}

type cacheRecord struct {
	method  string
	hit     bool
	latency time.Duration
}

// CacheMetric collects cache performance statistics. All public methods
// are safe for concurrent use.
type CacheMetric struct {
	mu      sync.RWMutex
	records []cacheRecord
}

// NewCacheMetric returns an empty CacheMetric.
func NewCacheMetric() *CacheMetric {
	return &CacheMetric{}
}

// Record adds a single cache access observation.
func (cm *CacheMetric) Record(method string, hit bool, latency time.Duration) {
	if cm == nil {
		return
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.records = append(cm.records, cacheRecord{method: method, hit: hit, latency: latency})
}

// HitRate returns the fraction of accesses that were cache hits.
// Returns 0.0 when no records exist.
func (cm *CacheMetric) HitRate() float64 {
	if cm == nil {
		return 0
	}
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	n := len(cm.records)
	if n == 0 {
		return 0
	}
	hits := 0
	for _, r := range cm.records {
		if r.hit {
			hits++
		}
	}
	return float64(hits) / float64(n)
}

// P50Latency returns the median (50th percentile) latency across all records.
// Returns 0 when no records exist.
func (cm *CacheMetric) P50Latency() time.Duration {
	return cm.percentile(0.50)
}

// P95Latency returns the 95th percentile latency across all records.
func (cm *CacheMetric) P95Latency() time.Duration {
	return cm.percentile(0.95)
}

// P99Latency returns the 99th percentile latency across all records.
func (cm *CacheMetric) P99Latency() time.Duration {
	return cm.percentile(0.99)
}

// Snapshot returns a point-in-time MetricSnapshot of all accumulated records.
func (cm *CacheMetric) Snapshot() MetricSnapshot {
	if cm == nil {
		return MetricSnapshot{}
	}
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	n := len(cm.records)
	if n == 0 {
		return MetricSnapshot{}
	}

	hits := 0
	lats := make([]time.Duration, n)
	for i, r := range cm.records {
		if r.hit {
			hits++
		}
		lats[i] = r.latency
	}

	sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })

	return MetricSnapshot{
		TotalRequests: n,
		Hits:          hits,
		Misses:        n - hits,
		HitRate:       float64(hits) / float64(n),
		P50Latency:    lats[percentileIndex(n, 0.50)],
		P95Latency:    lats[percentileIndex(n, 0.95)],
		P99Latency:    lats[percentileIndex(n, 0.99)],
	}
}

func (cm *CacheMetric) percentile(p float64) time.Duration {
	if cm == nil {
		return 0
	}
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	n := len(cm.records)
	if n == 0 {
		return 0
	}
	lats := make([]time.Duration, n)
	for i, r := range cm.records {
		lats[i] = r.latency
	}
	sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
	return lats[percentileIndex(n, p)]
}

func percentileIndex(n int, p float64) int {
	if n <= 1 {
		return 0
	}
	idx := int(float64(n-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return idx
}
