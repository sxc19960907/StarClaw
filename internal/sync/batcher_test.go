package sync

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestBuildBatchesStripsThinkingBeforeSizeCheck(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{{
		SessionID: "s1",
		UpdatedAt: now.Add(-time.Minute),
	}}
	body := []byte(`{"messages":[{"role":"assistant","content":[{"type":"thinking","thinking":"` + repeatByte('x', 200) + `"},{"type":"text","text":"ok"}]}]}`)
	marker := EmptyMarker()

	batches, err := BuildBatches(context.Background(), candidates, mapLoader(map[string][]byte{"s1": body}), Config{
		SingleSessionMaxBytes: 100,
	}, &marker, now)
	if err != nil {
		t.Fatalf("BuildBatches: %v", err)
	}
	if len(batches) != 1 || len(batches[0].Sessions) != 1 {
		t.Fatalf("batches = %+v, want one session", batches)
	}
	if _, failed := marker.Failed["s1"]; failed {
		t.Fatalf("marker failed should not include s1: %+v", marker.Failed["s1"])
	}
	if containsJSONType(t, batches[0].Sessions[0].Session, "thinking") {
		t.Fatalf("thinking block was not stripped: %s", batches[0].Sessions[0].Session)
	}
}

func TestBuildBatchesRecordsLoadAndSizeFailures(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{SessionID: "load-fails", UpdatedAt: now.Add(-2 * time.Minute)},
		{SessionID: "too-large", UpdatedAt: now.Add(-time.Minute)},
		{SessionID: "ok", UpdatedAt: now},
	}
	loader := func(_ string, id string) ([]byte, error) {
		if id == "load-fails" {
			return nil, errors.New("boom")
		}
		if id == "too-large" {
			return []byte(`{"messages":[{"role":"assistant","content":[{"type":"text","text":"` + repeatByte('x', 80) + `"}]}]}`), nil
		}
		return []byte(`{"ok":true}`), nil
	}
	marker := EmptyMarker()

	batches, err := BuildBatches(context.Background(), candidates, loader, Config{SingleSessionMaxBytes: 70}, &marker, now)
	if err != nil {
		t.Fatalf("BuildBatches: %v", err)
	}
	if len(batches) != 1 || len(batches[0].Sessions) != 1 || batches[0].Sessions[0].ID != "ok" {
		t.Fatalf("batches = %+v, want only ok", batches)
	}
	assertFailure(t, marker, "load-fails", "load_error", 0, now)
	if got := marker.Failed["too-large"]; got.Reason != "size_limit_exceeded" || got.SizeBytes == 0 {
		t.Fatalf("too-large failure = %+v, want size_limit_exceeded with size", got)
	}
	if marker.Failed["too-large"].LastObservedUpdatedAt != candidates[1].UpdatedAt {
		t.Fatalf("too-large LastObservedUpdatedAt = %v", marker.Failed["too-large"].LastObservedUpdatedAt)
	}
}

func TestBuildBatchesSplitsByMaxSessionsAndBytes(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{SessionID: "a", UpdatedAt: now},
		{SessionID: "b", UpdatedAt: now},
		{SessionID: "c", UpdatedAt: now},
		{SessionID: "d", UpdatedAt: now},
	}
	bodies := map[string][]byte{
		"a": []byte(`{"v":"` + repeatByte('a', 10) + `"}`),
		"b": []byte(`{"v":"` + repeatByte('b', 10) + `"}`),
		"c": []byte(`{"v":"` + repeatByte('c', 40) + `"}`),
		"d": []byte(`{"v":"` + repeatByte('d', 10) + `"}`),
	}

	batches, err := BuildBatches(context.Background(), candidates, mapLoader(bodies), Config{
		BatchMaxSessions: 2,
		BatchMaxBytes:    55,
	}, nil, now)
	if err != nil {
		t.Fatalf("BuildBatches: %v", err)
	}
	got := batchIDs(batches)
	want := [][]string{{"a", "b"}, {"c"}, {"d"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("batch ids = %v, want %v", got, want)
	}
}

func TestBuildBatchesReturnsContextCancellation(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	batches, err := BuildBatches(ctx, []Candidate{{SessionID: "s1"}}, mapLoader(nil), Config{}, nil, now)
	if err == nil {
		t.Fatal("BuildBatches should return context cancellation")
	}
	if len(batches) != 0 {
		t.Fatalf("batches = %+v, want none", batches)
	}
}

func mapLoader(bodies map[string][]byte) SessionLoader {
	return func(_ string, id string) ([]byte, error) {
		body, ok := bodies[id]
		if !ok {
			return nil, errors.New("missing")
		}
		return append([]byte(nil), body...), nil
	}
}

func assertFailure(t *testing.T, marker Marker, id, reason string, size uint64, now time.Time) {
	t.Helper()
	got, ok := marker.Failed[id]
	if !ok {
		t.Fatalf("failure %q missing", id)
	}
	if got.Reason != reason || got.Category != ClassifyReason(reason) || got.Attempts != 1 || got.SizeBytes != size {
		t.Fatalf("failure %q = %+v", id, got)
	}
	if !got.FirstAttemptAt.Equal(now) || !got.LastAttemptAt.Equal(now) {
		t.Fatalf("failure %q timestamps = %+v, want %v", id, got, now)
	}
}

func containsJSONType(t *testing.T, body json.RawMessage, blockType string) bool {
	t.Helper()
	var walk func(any) bool
	walk = func(v any) bool {
		switch typed := v.(type) {
		case map[string]any:
			if typed["type"] == blockType {
				return true
			}
			for _, child := range typed {
				if walk(child) {
					return true
				}
			}
		case []any:
			for _, child := range typed {
				if walk(child) {
					return true
				}
			}
		}
		return false
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	return walk(decoded)
}

func batchIDs(batches []BatchRequest) [][]string {
	out := make([][]string, 0, len(batches))
	for _, batch := range batches {
		ids := make([]string, 0, len(batch.Sessions))
		for _, session := range batch.Sessions {
			ids = append(ids, session.ID)
		}
		out = append(out, ids)
	}
	return out
}

func repeatByte(b byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return string(out)
}
