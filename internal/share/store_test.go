package share

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRecordListRetract(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	artifactDir := filepath.Join(dir, "web", "share_1")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(artifactDir, "report.html"), []byte("html"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	created, err := store.Record(Artifact{
		ID:        "share_1",
		Filename:  "report.html",
		LocalPath: filepath.Join(artifactDir, "report.html"),
		URL:       "http://localhost:7533/web/share_1/report.html",
		SizeBytes: 4,
		Purpose:   "review",
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if created.Status != StatusActive || !created.CreatedAt.Equal(now) {
		t.Fatalf("created = %#v", created)
	}

	active, err := store.List(false)
	if err != nil {
		t.Fatalf("List active: %v", err)
	}
	if len(active) != 1 || active[0].ID != "share_1" {
		t.Fatalf("active = %#v", active)
	}

	retracted, already, err := store.Retract("share_1")
	if err != nil {
		t.Fatalf("Retract: %v", err)
	}
	if already {
		t.Fatal("first retract should not be already")
	}
	if retracted.Status != StatusRetracted || retracted.RetractedAt == nil {
		t.Fatalf("retracted = %#v", retracted)
	}
	if _, err := os.Stat(artifactDir); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("artifact dir still exists or unexpected error: %v", err)
	}

	active, err = store.List(false)
	if err != nil {
		t.Fatalf("List after retract: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("active after retract = %#v", active)
	}
	all, err := store.List(true)
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if len(all) != 1 || all[0].Status != StatusRetracted {
		t.Fatalf("all = %#v", all)
	}

	_, already, err = store.Retract("share_1")
	if err != nil {
		t.Fatalf("second Retract: %v", err)
	}
	if !already {
		t.Fatal("second retract should be already")
	}
}

func TestStoreRetractMissing(t *testing.T) {
	store := NewStore(t.TempDir())
	_, _, err := store.Retract("missing")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("error = %v, want fs.ErrNotExist", err)
	}
}

func TestStoreCorruptManifestSidecar(t *testing.T) {
	dir := t.TempDir()
	manifest := filepath.Join(dir, "web", "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifest), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(manifest, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt manifest: %v", err)
	}
	store := NewStore(dir)
	got, err := store.List(true)
	if err != nil {
		t.Fatalf("List corrupt: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got = %#v, want empty", got)
	}
	if _, err := os.Stat(manifest + ".corrupt.bak"); err != nil {
		t.Fatalf("expected corrupt sidecar: %v", err)
	}
}
