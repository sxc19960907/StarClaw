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
