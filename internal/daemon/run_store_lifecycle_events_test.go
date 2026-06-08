package daemon

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

func TestRunStorePublishesLifecycleEvents(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("lifecycle")
	store := NewRunStore(10)
	store.SetEventBus(bus)

	store.Start(RunAgentRequest{
		RequestID: "life-run",
		Text:      "secret prompt body",
		Agent:     "helper",
		Channel:   ChannelHTTP,
		Source:    "test",
		SessionID: "sess-start",
	})
	started := readEvent(t, ch)
	if started.Type != "run_started" {
		t.Fatalf("start event type = %q, want run_started", started.Type)
	}
	startPayload := decodeLifecycleEventPayload(t, started)
	if startPayload["run_id"] != "life-run" || startPayload["status"] != "running" {
		t.Fatalf("start payload = %#v, want run_id life-run running", startPayload)
	}
	if startPayload["agent"] != "helper" || startPayload["channel"] != ChannelHTTP || startPayload["source"] != "test" {
		t.Fatalf("start metadata = %#v, want safe agent/channel/source", startPayload)
	}
	if _, ok := startPayload["prompt"]; ok {
		t.Fatalf("start payload must not include prompt: %#v", startPayload)
	}

	store.Complete("life-run", RunAgentResponse{
		SessionID: "sess-1",
		Usage:     map[string]int{"input_tokens": 3, "output_tokens": 4, "total_tokens": 7},
		BudgetStatus: &agent.TokenBudgetUsage{
			Status:       agent.TokenBudgetStatusOK,
			InputTokens:  3,
			OutputTokens: 4,
			TotalTokens:  7,
		},
		Routing: &agent.RouteRecommendation{
			Complexity: agent.ComplexitySimple,
			Route:      agent.RouteDirect,
			ModelTier:  "small",
			Reason:     "unit",
		},
	}, nil)
	completed := readEvent(t, ch)
	if completed.Type != "run_completed" {
		t.Fatalf("complete event type = %q, want run_completed", completed.Type)
	}
	completePayload := decodeLifecycleEventPayload(t, completed)
	if completePayload["status"] != "completed" || completePayload["session_id"] != "sess-1" {
		t.Fatalf("complete payload = %#v, want completed sess-1", completePayload)
	}
	usage, ok := completePayload["usage"].(map[string]any)
	if !ok || usage["total_tokens"] != float64(7) {
		t.Fatalf("usage payload = %#v, want total_tokens 7", completePayload["usage"])
	}
}

func TestRunStorePublishesRunErrorEvents(t *testing.T) {
	t.Run("go error", func(t *testing.T) {
		bus := NewEventBus()
		ch := bus.Subscribe("lifecycle-error")
		store := NewRunStore(10)
		store.SetEventBus(bus)

		store.Start(RunAgentRequest{RequestID: "err-run", Channel: ChannelHTTP})
		_ = readEvent(t, ch)
		store.Complete("err-run", RunAgentResponse{}, errors.New("provider failed with bearer secret-token"))

		evt := readEvent(t, ch)
		if evt.Type != "run_error" {
			t.Fatalf("event type = %q, want run_error", evt.Type)
		}
		payload := decodeLifecycleEventPayload(t, evt)
		if payload["status"] != "error" || payload["error"] != "[REDACTED]" {
			t.Fatalf("error payload = %#v, want redacted error status", payload)
		}
	})

	t.Run("response error", func(t *testing.T) {
		bus := NewEventBus()
		ch := bus.Subscribe("lifecycle-response-error")
		store := NewRunStore(10)
		store.SetEventBus(bus)

		store.Start(RunAgentRequest{RequestID: "response-err-run", Channel: ChannelHTTP})
		_ = readEvent(t, ch)
		store.Complete("response-err-run", RunAgentResponse{Error: "plain failure"}, nil)

		evt := readEvent(t, ch)
		payload := decodeLifecycleEventPayload(t, evt)
		if evt.Type != "run_error" || payload["error"] != "plain failure" {
			t.Fatalf("event = %#v payload = %#v, want response run_error", evt, payload)
		}
	})
}

func TestRunStoreLifecycleEventsDoNotLeakPayloads(t *testing.T) {
	bus := NewEventBus()
	ch := bus.Subscribe("redaction")
	store := NewRunStore(10)
	store.SetEventBus(bus)

	store.Start(RunAgentRequest{
		RequestID: "leak-run",
		Text:      "phase5 prompt secret",
		Channel:   ChannelHTTP,
		Source:    "test",
	})
	store.AddEvent("leak-run", EventToolCall, map[string]any{
		"args":     `{"api_key":"sk-phase5-secret"}`,
		"response": "phase5 provider response body",
	})
	store.Complete("leak-run", RunAgentResponse{
		SessionID: "sess",
		Error:     "Bearer phase5-token",
	}, nil)

	events := []Event{readEvent(t, ch), readEvent(t, ch)}
	encoded, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal lifecycle events: %v", err)
	}
	assertNoForbiddenLeak(t, "lifecycle bus events", encoded, secretLeakForbiddenValues())
	if strings.Contains(string(encoded), `"text":`) || strings.Contains(string(encoded), `"args":`) {
		t.Fatalf("lifecycle events included raw text/args keys: %s", encoded)
	}
}

func TestRunStoreLifecycleNilBusDoesNotPanic(t *testing.T) {
	store := NewRunStore(10)
	store.Start(RunAgentRequest{RequestID: "nil-bus", Channel: ChannelHTTP})
	store.Complete("nil-bus", RunAgentResponse{SessionID: "sess"}, nil)
	store.SetEventBus(nil)
	store.Start(RunAgentRequest{RequestID: "nil-bus-2", Channel: ChannelHTTP})
	store.Complete("nil-bus-2", RunAgentResponse{SessionID: "sess"}, nil)
}

func decodeLifecycleEventPayload(t *testing.T, evt Event) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
		t.Fatalf("decode lifecycle event %s: %v data=%s", evt.Type, err, evt.Data)
	}
	return payload
}
