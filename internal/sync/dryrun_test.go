package sync

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestDryRunUploaderWritesOutboxAndAcceptsIDs(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 6, 8, 12, 0, 0, 123000000, time.UTC)
	uploader := &DryRunUploader{
		OutboxDir: dir,
		Now:       func() time.Time { return now },
	}
	batch := BatchRequest{
		ClientVersion: "starclaw/test",
		SyncAt:        now,
		Sessions: []SessionEnvelope{
			{ID: "sess-a", AgentName: "default", Session: json.RawMessage(`{"id":"sess-a"}`)},
			{ID: "sess-b", AgentName: "helper", Session: json.RawMessage(`{"id":"sess-b"}`)},
		},
	}

	resp, err := uploader.Send(context.Background(), batch)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !reflect.DeepEqual(resp.Accepted, []string{"sess-a", "sess-b"}) {
		t.Fatalf("Accepted = %v, want [sess-a sess-b]", resp.Accepted)
	}
	if len(resp.Rejected) != 0 {
		t.Fatalf("Rejected = %v, want empty", resp.Rejected)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("outbox entries = %d, want 1", len(entries))
	}
	path := filepath.Join(dir, entries[0].Name())
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat outbox: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("outbox mode = %o, want 0600", got)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read outbox: %v", err)
	}
	if !json.Valid(body) {
		t.Fatalf("outbox is not valid JSON: %s", body)
	}
	var got BatchRequest
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal outbox: %v", err)
	}
	if got.ClientVersion != batch.ClientVersion || len(got.Sessions) != len(batch.Sessions) {
		t.Fatalf("outbox batch = %+v, want %+v", got, batch)
	}
}

func TestDryRunUploaderRejectsEmptyOutbox(t *testing.T) {
	uploader := &DryRunUploader{}
	if _, err := uploader.Send(context.Background(), BatchRequest{}); err == nil {
		t.Fatal("Send with empty outbox should error")
	}
}

func TestDryRunUploaderHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	uploader := &DryRunUploader{OutboxDir: t.TempDir()}
	if _, err := uploader.Send(ctx, BatchRequest{}); err == nil {
		t.Fatal("Send with cancelled context should error")
	}
}
