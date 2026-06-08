package sync

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestDiscoverCandidatesFindsDefaultAndNamedSessions(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	writeCandidateFile(t, dir, "", "default-1", "default-1", "", base.Add(time.Minute))
	writeCandidateFile(t, dir, "research", "named-1", "named-1", "", base.Add(2*time.Minute))

	got, skipped, err := DiscoverCandidates(context.Background(), ScannerDeps{StarclawDir: dir}, Config{}, Marker{LastSyncAt: base}, base)
	if err != nil {
		t.Fatalf("DiscoverCandidates: %v", err)
	}
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}
	if ids := candidateIDs(got); !reflect.DeepEqual(ids, []string{"default-1", "named-1"}) {
		t.Fatalf("ids = %v", ids)
	}
	if got[0].AgentName != "" {
		t.Fatalf("default AgentName = %q, want empty", got[0].AgentName)
	}
	if got[1].AgentName != "research" {
		t.Fatalf("named AgentName = %q, want research", got[1].AgentName)
	}
}

func TestDiscoverCandidatesFiltersExcludesDedupeAndSkipsInvalid(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	writeCandidateFile(t, dir, "", "old", "old", "", base.Add(-time.Minute))
	writeCandidateFile(t, dir, "", "local-excluded", "local-excluded", "", base.Add(time.Minute))
	writeCandidateFile(t, dir, "skip-agent", "agent-excluded", "agent-excluded", "", base.Add(2*time.Minute))
	writeCandidateFile(t, dir, "", "remote-excluded", "remote-excluded", "remote", base.Add(3*time.Minute))
	writeCandidateFile(t, dir, "", "dupe", "dupe", "kept-source", base.Add(4*time.Minute))
	writeCandidateFile(t, dir, "worker", "dupe", "dupe", "kept-source", base.Add(5*time.Minute))
	writeRawSessionFile(t, dir, "", "corrupt", []byte("{not-json"))
	writeRawSessionFile(t, dir, "", "missing-updated", []byte(`{"id":"missing-updated"}`))

	cfg := Config{
		ExcludeAgents:  []string{"skip-agent"},
		ExcludeSources: []string{"remote", "local"},
	}
	got, skipped, err := DiscoverCandidates(context.Background(), ScannerDeps{StarclawDir: dir}, cfg, Marker{LastSyncAt: base}, base)
	if err != nil {
		t.Fatalf("DiscoverCandidates: %v", err)
	}
	if skipped != 2 {
		t.Fatalf("skipped = %d, want 2", skipped)
	}
	if ids := candidateIDs(got); !reflect.DeepEqual(ids, []string{"dupe"}) {
		t.Fatalf("ids = %v, want [dupe]", ids)
	}
	if got[0].AgentName != "worker" {
		t.Fatalf("deduped AgentName = %q, want worker", got[0].AgentName)
	}
}

func TestDiscoverCandidatesPermanentFailureNoChurnAndEditedRetry(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	observed := base.Add(2 * time.Minute)
	writeCandidateFile(t, dir, "", "unchanged", "unchanged", "", observed)
	writeCandidateFile(t, dir, "", "edited", "edited", "", observed.Add(time.Minute))

	marker := Marker{
		LastSyncAt: base,
		Failed: map[string]FailedEntry{
			"unchanged": {
				Category:              CategoryPermanent,
				LastObservedUpdatedAt: observed,
			},
			"edited": {
				Category:              CategoryPermanent,
				LastObservedUpdatedAt: observed,
			},
		},
	}

	got, skipped, err := DiscoverCandidates(context.Background(), ScannerDeps{StarclawDir: dir}, Config{}, marker, base)
	if err != nil {
		t.Fatalf("DiscoverCandidates: %v", err)
	}
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}
	if ids := candidateIDs(got); !reflect.DeepEqual(ids, []string{"edited"}) {
		t.Fatalf("ids = %v, want [edited]", ids)
	}
}

func TestDiscoverCandidatesIncludesDueTransientRetriesDeterministically(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	due := base.Add(-time.Minute)
	future := base.Add(time.Minute)
	writeCandidateFile(t, dir, "", "due-local", "due-local", "", base.Add(-time.Hour))
	writeCandidateFile(t, dir, "", "due-remote", "due-remote", "remote", base.Add(-time.Hour))

	marker := Marker{
		LastSyncAt: base,
		Failed: map[string]FailedEntry{
			"due-local": {
				Category:      CategoryTransient,
				LastAttemptAt: base.Add(-2 * time.Hour),
				NextAttemptAt: &due,
			},
			"due-missing": {
				Category:      CategoryTransient,
				LastAttemptAt: base.Add(-3 * time.Hour),
				NextAttemptAt: &due,
			},
			"not-due": {
				Category:      CategoryTransient,
				LastAttemptAt: base.Add(-4 * time.Hour),
				NextAttemptAt: &future,
			},
			"due-remote": {
				Category:      CategoryTransient,
				LastAttemptAt: base.Add(-5 * time.Hour),
				NextAttemptAt: &due,
			},
		},
	}

	got, skipped, err := DiscoverCandidates(context.Background(), ScannerDeps{StarclawDir: dir}, Config{ExcludeSources: []string{"remote"}}, marker, base)
	if err != nil {
		t.Fatalf("DiscoverCandidates: %v", err)
	}
	if skipped != 0 {
		t.Fatalf("skipped = %d, want 0", skipped)
	}
	if ids := candidateIDs(got); !reflect.DeepEqual(ids, []string{"due-missing", "due-local"}) {
		t.Fatalf("ids = %v, want [due-missing due-local]", ids)
	}
}

func TestDiscoverCandidatesReturnsContextCancellation(t *testing.T) {
	dir := t.TempDir()
	writeCandidateFile(t, dir, "", "one", "one", "", time.Now().UTC())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := DiscoverCandidates(ctx, ScannerDeps{StarclawDir: dir}, Config{}, EmptyMarker(), time.Now().UTC())
	if err == nil {
		t.Fatal("DiscoverCandidates should return context cancellation")
	}
}

func candidateIDs(candidates []Candidate) []string {
	ids := make([]string, 0, len(candidates))
	for _, c := range candidates {
		ids = append(ids, c.SessionID)
	}
	return ids
}

func writeCandidateFile(t *testing.T, starclawDir, agent, filename, id, source string, updatedAt time.Time) {
	t.Helper()
	body := `{"id":"` + id + `","updated_at":"` + updatedAt.Format(time.RFC3339Nano) + `","source":"` + source + `","messages":[]}`
	writeRawSessionFile(t, starclawDir, agent, filename, []byte(body))
}

func writeRawSessionFile(t *testing.T, starclawDir, agent, filename string, body []byte) {
	t.Helper()
	dir := filepath.Join(starclawDir, "sessions")
	if agent != "" {
		dir = filepath.Join(dir, agent)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename+".json"), body, 0600); err != nil {
		t.Fatalf("write session: %v", err)
	}
}
