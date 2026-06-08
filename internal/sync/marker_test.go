package sync

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestMarkerRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "sync_marker.json")
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	next := now.Add(time.Hour)
	want := Marker{
		Version:         MarkerVersion,
		LastSyncAt:      now,
		LastSyncCount:   2,
		LastSyncOutcome: OutcomeOK,
		Failed: map[string]FailedEntry{
			"sess-1": {
				Reason:                "size_limit_exceeded",
				Category:              CategoryPermanent,
				Attempts:              1,
				SizeBytes:             4096,
				FirstAttemptAt:        now,
				LastAttemptAt:         now,
				LastObservedUpdatedAt: now.Add(-time.Minute),
			},
			"sess-2": {
				Reason:                "retryable",
				Category:              CategoryTransient,
				Attempts:              2,
				SizeBytes:             2048,
				FirstAttemptAt:        now.Add(-time.Hour),
				LastAttemptAt:         now,
				LastObservedUpdatedAt: now.Add(-time.Minute),
				NextAttemptAt:         &next,
			},
		},
	}

	if err := WriteMarkerAtomic(path, want); err != nil {
		t.Fatalf("WriteMarkerAtomic: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat marker: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("marker mode = %o, want 0600", got)
	}

	got, err := ReadMarker(path)
	if err != nil {
		t.Fatalf("ReadMarker: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marker mismatch:\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestMarkerMissingFileReturnsEmpty(t *testing.T) {
	marker, err := ReadMarker(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("ReadMarker missing: %v", err)
	}
	if marker.Version != MarkerVersion {
		t.Fatalf("Version = %d, want %d", marker.Version, MarkerVersion)
	}
	if !marker.LastSyncAt.IsZero() {
		t.Fatalf("LastSyncAt = %v, want zero", marker.LastSyncAt)
	}
	if marker.Failed == nil {
		t.Fatal("Failed map should be initialized")
	}
}

func TestMarkerCorruptSidecarsAndResets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sync_marker.json")
	bad := []byte("not json {{{")
	if err := os.WriteFile(path, bad, 0644); err != nil {
		t.Fatalf("seed corrupt marker: %v", err)
	}

	marker, err := ReadMarker(path)
	if err != nil {
		t.Fatalf("ReadMarker corrupt: %v", err)
	}
	if marker.Version != MarkerVersion || !marker.LastSyncAt.IsZero() {
		t.Fatalf("marker = %+v, want empty current marker", marker)
	}
	sidecar := path + ".corrupt.bak"
	got, err := os.ReadFile(sidecar)
	if err != nil {
		t.Fatalf("read sidecar: %v", err)
	}
	if string(got) != string(bad) {
		t.Fatalf("sidecar = %q, want %q", got, bad)
	}
}

func TestMarkerUnknownVersionSidecarsAndCanRecover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sync_marker.json")
	bad := []byte(`{"version":999,"last_sync_at":"2099-01-01T00:00:00Z","failed":{}}`)
	if err := os.WriteFile(path, bad, 0644); err != nil {
		t.Fatalf("seed unknown marker: %v", err)
	}

	marker, err := ReadMarker(path)
	if err != nil {
		t.Fatalf("ReadMarker unknown version: %v", err)
	}
	sidecar := path + ".unknown-v999.bak"
	if _, err := os.Stat(sidecar); err != nil {
		t.Fatalf("sidecar missing: %v", err)
	}
	marker.LastSyncOutcome = OutcomeNoop
	if err := WriteMarkerAtomic(path, marker); err != nil {
		t.Fatalf("WriteMarkerAtomic recovery: %v", err)
	}
	recovered, err := ReadMarker(path)
	if err != nil {
		t.Fatalf("ReadMarker recovered: %v", err)
	}
	if recovered.Version != MarkerVersion || recovered.LastSyncOutcome != OutcomeNoop {
		t.Fatalf("recovered = %+v, want current noop marker", recovered)
	}
	if _, err := os.Stat(sidecar); err != nil {
		t.Fatalf("sidecar should remain after recovery: %v", err)
	}
}
