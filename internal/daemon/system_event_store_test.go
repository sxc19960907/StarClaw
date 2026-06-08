package daemon

import (
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

func TestSystemEventStoreEnqueueDrainRouteIsolation(t *testing.T) {
	store := NewSystemEventStore(3)
	store.Enqueue("route-a", agent.SystemEvent{Text: "a1"})
	store.Enqueue("route-b", agent.SystemEvent{Text: "b1"})
	store.Enqueue("route-a", agent.SystemEvent{Text: "a2"})

	a := store.Drain("route-a")
	if len(a) != 2 || a[0].Text != "a1" || a[1].Text != "a2" {
		t.Fatalf("route-a events = %#v", a)
	}
	if again := store.Drain("route-a"); len(again) != 0 {
		t.Fatalf("route-a drained twice = %#v", again)
	}
	b := store.Drain("route-b")
	if len(b) != 1 || b[0].Text != "b1" {
		t.Fatalf("route-b events = %#v", b)
	}
}

func TestSystemEventStoreCollapseAndEvict(t *testing.T) {
	store := NewSystemEventStore(2)
	store.Enqueue("route", agent.SystemEvent{Text: "old", ContextKey: "same"})
	store.Enqueue("route", agent.SystemEvent{Text: "new", ContextKey: "same"})
	store.Enqueue("route", agent.SystemEvent{Text: "tail"})

	events := store.Drain("route")
	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2: %#v", len(events), events)
	}
	if events[0].Text != "new" || events[1].Text != "tail" {
		t.Fatalf("events = %#v", events)
	}

	store.Enqueue("route", agent.SystemEvent{Text: "one"})
	store.Enqueue("route", agent.SystemEvent{Text: "two"})
	store.Enqueue("route", agent.SystemEvent{Text: "three"})
	events = store.Drain("route")
	if len(events) != 2 || events[0].Text != "two" || events[1].Text != "three" {
		t.Fatalf("evicted events = %#v", events)
	}
}

func TestSystemEventStoreForgetAndNoop(t *testing.T) {
	var nilStore *SystemEventStore
	nilStore.Enqueue("route", agent.SystemEvent{Text: "ignored"})
	if got := nilStore.Drain("route"); len(got) != 0 {
		t.Fatalf("nil drain = %#v", got)
	}
	nilStore.Forget("route")

	store := NewSystemEventStore(0)
	store.Enqueue("", agent.SystemEvent{Text: "ignored"})
	if got := store.Drain(""); len(got) != 0 {
		t.Fatalf("empty route drain = %#v", got)
	}
	store.Enqueue("route", agent.SystemEvent{Text: "event"})
	store.Forget("route")
	if got := store.Drain("route"); len(got) != 0 {
		t.Fatalf("forgotten events = %#v", got)
	}
}
