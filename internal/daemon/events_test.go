package daemon

import (
	"testing"
)

func TestEventBusSubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("test-1")

	bus.Publish(Event{Type: "test", Data: "hello"})

	select {
	case evt := <-ch:
		if evt.Type != "test" {
			t.Errorf("Type = %q, want %q", evt.Type, "test")
		}
		if evt.Data != "hello" {
			t.Errorf("Data = %q, want %q", evt.Data, "hello")
		}
	default:
		t.Fatal("expected event, got none")
	}
}

func TestEventBusMultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	ch1 := bus.Subscribe("multi-1")
	ch2 := bus.Subscribe("multi-2")

	bus.Publish(Event{Type: "broadcast", Data: "data"})

	// Both subscribers should receive the event
	select {
	case <-ch1:
	default:
		t.Error("subscriber 1 did not receive event")
	}

	select {
	case <-ch2:
	default:
		t.Error("subscriber 2 did not receive event")
	}
}

func TestEventBusUnsubscribe(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("unsub-test")

	bus.Unsubscribe("unsub-test")

	// After unsubscribing, the channel should not receive events
	bus.Publish(Event{Type: "after-unsub", Data: "data"})

	select {
	case <-ch:
		t.Error("received event after unsubscribe")
	default:
		// expected
	}
}

func TestEventBusPublishNonBlocking(t *testing.T) {
	bus := NewEventBus()
	// Subscribe with buffer size 64 (default)
	_ = bus.Subscribe("slow-consumer")

	// Fill up the buffer by publishing many events
	for i := 0; i < 128; i++ {
		bus.Publish(Event{Type: "filler", Data: "data"})
	}

	// Publish should never block; no panic expected
	bus.Publish(Event{Type: "last", Data: "data"})
}

func TestEventBusMultiplePublishOrder(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("order-test")

	bus.Publish(Event{Type: "first", Data: "one"})
	bus.Publish(Event{Type: "second", Data: "two"})

	evt1 := <-ch
	if evt1.Type != "first" || evt1.Data != "one" {
		t.Errorf("expected first event, got %+v", evt1)
	}

	evt2 := <-ch
	if evt2.Type != "second" || evt2.Data != "two" {
		t.Errorf("expected second event, got %+v", evt2)
	}
}

func TestEventBusUnsubscribeNonExistent(t *testing.T) {
	bus := NewEventBus()
	// Unsubscribing a non-existent subscriber should not panic
	bus.Unsubscribe("nonexistent")
}

func TestEventBusSubscribeSameIDReplaces(t *testing.T) {
	bus := NewEventBus()
	ch1 := bus.Subscribe("same-id")
	ch2 := bus.Subscribe("same-id")

	// ch1 should be de-registered; only ch2 gets events
	bus.Publish(Event{Type: "test", Data: "data"})

	select {
	case <-ch1:
		t.Error("old subscriber received event after replacement")
	default:
	}

	select {
	case <-ch2:
	default:
		t.Error("new subscriber did not receive event")
	}
}

func TestEventBusAssignsIDsAndReplaysSinceCursor(t *testing.T) {
	bus := NewEventBus()
	bus.Publish(Event{Type: "first", Data: "one"})
	bus.Publish(Event{Type: "second", Data: "two"})
	bus.Publish(Event{Type: "third", Data: "three"})

	all := bus.EventsSince("")
	if len(all) != 3 {
		t.Fatalf("all events = %d, want 3", len(all))
	}
	if all[0].ID != "1" || all[1].ID != "2" || all[2].ID != "3" {
		t.Fatalf("event IDs = %#v, want 1/2/3", all)
	}

	replayed := bus.EventsSince("1")
	if len(replayed) != 2 || replayed[0].Type != "second" || replayed[1].Type != "third" {
		t.Fatalf("replayed = %#v, want second and third", replayed)
	}
}

func TestEventBusSubscribeWithReplayReturnsMissedThenLive(t *testing.T) {
	bus := NewEventBus()
	bus.Publish(Event{Type: "first", Data: "one"})
	bus.Publish(Event{Type: "second", Data: "two"})

	replayed, ch := bus.SubscribeWithReplay("replay-live", "1")
	defer bus.Unsubscribe("replay-live")
	if len(replayed) != 1 || replayed[0].ID != "2" || replayed[0].Type != "second" {
		t.Fatalf("replayed = %#v, want second id=2", replayed)
	}

	bus.Publish(Event{Type: "third", Data: "three"})
	select {
	case evt := <-ch:
		if evt.ID != "3" || evt.Type != "third" {
			t.Fatalf("live event = %#v, want third id=3", evt)
		}
	default:
		t.Fatal("expected live event after subscribe-with-replay")
	}

	select {
	case evt := <-ch:
		t.Fatalf("unexpected duplicate event after replay/live boundary: %#v", evt)
	default:
	}
}

func TestEventBusSubscribeWithReplayInvalidCursorReplaysHistory(t *testing.T) {
	bus := NewEventBus()
	bus.Publish(Event{Type: "first", Data: "one"})
	bus.Publish(Event{Type: "second", Data: "two"})

	replayed, ch := bus.SubscribeWithReplay("invalid-cursor", "not-a-number")
	defer bus.Unsubscribe("invalid-cursor")
	if len(replayed) != 2 || replayed[0].ID != "1" || replayed[1].ID != "2" {
		t.Fatalf("replayed = %#v, want full history for invalid cursor", replayed)
	}

	select {
	case evt := <-ch:
		t.Fatalf("subscriber channel should not receive replayed history: %#v", evt)
	default:
	}
}
