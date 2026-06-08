package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const MarkerVersion = 1

const (
	OutcomeOK             = "ok"
	OutcomePartial        = "partial"
	OutcomeTransportError = "transport_error"
	OutcomeNoop           = "noop"
)

const (
	CategoryTransient = "transient"
	CategoryPermanent = "permanent"
)

// Marker tracks local sync watermark and failed session retry state.
type Marker struct {
	Version         int                    `json:"version"`
	LastSyncAt      time.Time              `json:"last_sync_at"`
	LastSyncCount   int                    `json:"last_sync_count"`
	LastSyncOutcome string                 `json:"last_sync_outcome"`
	Failed          map[string]FailedEntry `json:"failed"`
}

// FailedEntry records a session failure for future retry decisions.
type FailedEntry struct {
	Reason                string     `json:"reason"`
	Category              string     `json:"category"`
	Attempts              int        `json:"attempts"`
	SizeBytes             uint64     `json:"size_bytes"`
	FirstAttemptAt        time.Time  `json:"first_attempt_at"`
	LastAttemptAt         time.Time  `json:"last_attempt_at"`
	LastObservedUpdatedAt time.Time  `json:"last_observed_updated_at"`
	NextAttemptAt         *time.Time `json:"next_attempt_at"`
}

// EmptyMarker returns a current-version marker with no sync watermark.
func EmptyMarker() Marker {
	return Marker{
		Version: MarkerVersion,
		Failed:  map[string]FailedEntry{},
	}
}

// ReadMarker loads a marker. Missing, corrupt, or future-version files reset to
// EmptyMarker; corrupt and future-version files are sidecarred for inspection.
func ReadMarker(path string) (Marker, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return EmptyMarker(), nil
		}
		return Marker{}, fmt.Errorf("read marker: %w", err)
	}

	var probe struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		_ = sidecarMarker(path, data, ".corrupt.bak")
		return EmptyMarker(), nil
	}
	if probe.Version <= 0 {
		_ = sidecarMarker(path, data, ".corrupt.bak")
		return EmptyMarker(), nil
	}
	if probe.Version != MarkerVersion {
		_ = sidecarMarker(path, data, fmt.Sprintf(".unknown-v%d.bak", probe.Version))
		return EmptyMarker(), nil
	}

	var marker Marker
	if err := json.Unmarshal(data, &marker); err != nil {
		_ = sidecarMarker(path, data, ".corrupt.bak")
		return EmptyMarker(), nil
	}
	if marker.Failed == nil {
		marker.Failed = map[string]FailedEntry{}
	}
	return marker, nil
}

func sidecarMarker(originalPath string, contents []byte, suffix string) error {
	return os.WriteFile(originalPath+suffix, contents, 0600)
}

// WriteMarkerAtomic writes marker via same-directory temp file and rename.
func WriteMarkerAtomic(path string, marker Marker) error {
	if marker.Version == 0 {
		marker.Version = MarkerVersion
	}
	if marker.Failed == nil {
		marker.Failed = map[string]FailedEntry{}
	}
	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal marker: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("mkdir marker dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "sync_marker-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create marker temp: %w", err)
	}
	tmpPath := tmp.Name()
	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("chmod marker temp: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write marker temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close marker temp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename marker temp: %w", err)
	}
	return nil
}
