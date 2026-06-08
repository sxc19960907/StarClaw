package daemon

import (
	"errors"
	"testing"
)

func TestMailboxStoreEnqueueOrdersPriorityThenFIFO(t *testing.T) {
	store := NewMailboxStore(10)
	first, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Text: "normal one", Priority: QueuePriorityNormal})
	if err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	second, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Text: "urgent", Priority: QueuePriorityHigh})
	if err != nil {
		t.Fatalf("enqueue second: %v", err)
	}
	third, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Text: "normal two", Priority: QueuePriorityNormal})
	if err != nil {
		t.Fatalf("enqueue third: %v", err)
	}

	claimed, err := store.Claim("route-a", 10)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	got := idsOfQueuedMessages(claimed)
	want := []string{second.ID, first.ID, third.ID}
	if !sameStringSlice(got, want) {
		t.Fatalf("claim order = %#v, want %#v", got, want)
	}
}

func TestMailboxStoreCapacity(t *testing.T) {
	store := NewMailboxStore(1)
	if _, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Text: "one"}); err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	if _, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Text: "two"}); !errors.Is(err, ErrMailboxFull) {
		t.Fatalf("enqueue second error = %v, want ErrMailboxFull", err)
	}
}

func TestMailboxStoreDeduplicatesExternalIDPerRoute(t *testing.T) {
	store := NewMailboxStore(10)
	first, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Source: "github", ExternalID: "delivery-1", Text: "one"})
	if err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	second, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Source: "github", ExternalID: "delivery-1", Text: "two"})
	if err != nil {
		t.Fatalf("enqueue duplicate: %v", err)
	}
	if !second.Duplicate || second.ID != first.ID {
		t.Fatalf("duplicate result = %#v, first %#v", second, first)
	}
	third, err := store.Enqueue(QueuedMessage{RouteKey: "route-b", Source: "github", ExternalID: "delivery-1", Text: "three"})
	if err != nil {
		t.Fatalf("enqueue route-b: %v", err)
	}
	if third.Duplicate || third.ID == first.ID {
		t.Fatalf("dedup crossed route boundary: third=%#v first=%#v", third, first)
	}
}

func TestMailboxStoreSnapshotIsDefensive(t *testing.T) {
	store := NewMailboxStore(10)
	msg, err := store.Enqueue(QueuedMessage{
		RouteKey: "route-a",
		Text:     "one",
		Metadata: map[string]string{"token": "secret-token", "safe": "value"},
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	snapshot := store.List("")
	if len(snapshot) != 1 {
		t.Fatalf("snapshot count = %d, want 1", len(snapshot))
	}
	snapshot[0].Text = "changed"
	snapshot[0].Metadata["safe"] = "changed"

	got, ok := store.Get(msg.ID)
	if !ok {
		t.Fatal("expected message")
	}
	if got.Text != "one" || got.Metadata["safe"] != "value" {
		t.Fatalf("store mutated through snapshot: %#v", got)
	}
	if _, ok := got.Metadata["token"]; ok {
		t.Fatalf("sensitive metadata key was not sanitized: %#v", got.Metadata)
	}
}

func TestMailboxStoreClaimAckReleaseLifecycle(t *testing.T) {
	store := NewMailboxStore(10)
	first, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Text: "one"})
	if err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	second, err := store.Enqueue(QueuedMessage{RouteKey: "route-a", Text: "two"})
	if err != nil {
		t.Fatalf("enqueue second: %v", err)
	}

	claimed, err := store.Claim("route-a", 1)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if len(claimed) != 1 || claimed[0].ID != first.ID || claimed[0].Status != QueueStatusClaimed || claimed[0].ClaimID == "" {
		t.Fatalf("claimed = %#v, want first claimed with claim id", claimed)
	}

	next, err := store.Claim("route-a", 10)
	if err != nil {
		t.Fatalf("claim next: %v", err)
	}
	if len(next) != 1 || next[0].ID != second.ID {
		t.Fatalf("next claim = %#v, want second only", next)
	}

	if !store.Release(first.ID, claimed[0].ClaimID) {
		t.Fatal("expected release to succeed")
	}
	reclaimed, err := store.Claim("route-a", 1)
	if err != nil {
		t.Fatalf("reclaim: %v", err)
	}
	if len(reclaimed) != 1 || reclaimed[0].ID != first.ID || reclaimed[0].Attempt != 2 {
		t.Fatalf("reclaimed = %#v, want first attempt 2", reclaimed)
	}
	if !store.Ack(first.ID, reclaimed[0].ClaimID) {
		t.Fatal("expected ack to succeed")
	}
	if got, ok := store.Get(first.ID); !ok || got.Status != QueueStatusAcknowledged {
		t.Fatalf("acked message = %#v ok=%v", got, ok)
	}
	if later, err := store.Claim("route-a", 10); err != nil || len(later) != 0 {
		t.Fatalf("later claim = %#v err=%v, want none", later, err)
	}
}

func idsOfQueuedMessages(messages []QueuedMessage) []string {
	out := make([]string, 0, len(messages))
	for _, msg := range messages {
		out = append(out, msg.ID)
	}
	return out
}

func sameStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
