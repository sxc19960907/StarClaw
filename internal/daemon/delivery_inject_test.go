package daemon

import (
	"strings"
	"testing"
)

func TestFormatDeliveryFailure(t *testing.T) {
	t.Run("permanent phrases reactive removal", func(t *testing.T) {
		got := formatDeliveryFailure(ReplyDeliveryResultPayload{
			OK: false, Channel: "slack", ThreadID: "C1-99.1", Class: ClassPermanent, Reason: "bot was kicked",
		})
		if !strings.Contains(got, "reply to slack C1 FAILED") {
			t.Fatalf("slack channel label missing: %q", got)
		}
		if !strings.Contains(got, "bot was kicked") || !strings.Contains(got, "user did not see it") {
			t.Fatalf("permanent line missing failure facts: %q", got)
		}
		if !strings.Contains(strings.ToLower(got), "will not receive") {
			t.Fatalf("permanent line should carry re-add implication: %q", got)
		}
	})

	t.Run("transient is cautious", func(t *testing.T) {
		got := formatDeliveryFailure(ReplyDeliveryResultPayload{
			OK: false, Channel: "slack", Class: ClassTransient, Reason: "gateway closed",
		})
		if strings.Contains(strings.ToLower(got), "will not receive") {
			t.Fatalf("transient must not assert removal: %q", got)
		}
		if !strings.Contains(got, "may not have been delivered") || !strings.Contains(got, "retry is in progress") {
			t.Fatalf("transient phrasing missing: %q", got)
		}
	})

	t.Run("empty reason defaults", func(t *testing.T) {
		got := formatDeliveryFailure(ReplyDeliveryResultPayload{Channel: "telegram", Class: ClassTransient})
		if !strings.Contains(got, "delivery failed") {
			t.Fatalf("default reason missing: %q", got)
		}
	})
}

func TestHandleReplyDeliveryResult(t *testing.T) {
	store := NewSystemEventStore(20)
	idx := NewReplyRouteIndex(8)
	idx.Put("m-ok", "route-ok")
	idx.Put("m-fail", "route-fail")

	consumer := newDeliveryResultConsumer(store, idx)
	consumer(ReplyDeliveryResultPayload{OK: true, Channel: "slack"}, "m-ok")
	if got := store.Drain("route-ok"); len(got) != 0 {
		t.Fatalf("success must be silent, got %#v", got)
	}

	consumer(ReplyDeliveryResultPayload{
		OK: false, Channel: "slack", ThreadID: "C1-99.1", Class: ClassPermanent, Reason: "bot was kicked",
	}, "m-fail")
	got := store.Drain("route-fail")
	if len(got) != 1 {
		t.Fatalf("events len = %d, want 1: %#v", len(got), got)
	}
	if !got[0].Trusted {
		t.Fatal("delivery failure event should be trusted")
	}
	if got[0].ContextKey != "delivery-fail:slack:C1-99.1" {
		t.Fatalf("context key = %q", got[0].ContextKey)
	}
	if !strings.Contains(got[0].Text, "FAILED") {
		t.Fatalf("event text = %q", got[0].Text)
	}

	consumer(ReplyDeliveryResultPayload{OK: false, Channel: "slack", Class: ClassPermanent}, "missing")
	if got := store.Drain("route-fail"); len(got) != 0 {
		t.Fatalf("missing route leaked event: %#v", got)
	}
}

func TestHandleReplyDeliveryResultNilStoresNoop(t *testing.T) {
	HandleReplyDeliveryResult(nil, nil, ReplyDeliveryResultPayload{OK: false}, "m")
}
