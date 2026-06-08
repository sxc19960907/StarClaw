package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueueAPIRoutes(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var created struct {
		Message QueuedMessage `json:"message"`
	}
	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"hello","source":"webhook","external_id":"evt-1"}`, http.StatusAccepted, &created)
	if created.Message.ID == "" || created.Message.RouteKey != "route-a" || created.Message.Status != QueueStatusQueued {
		t.Fatalf("created = %#v", created)
	}

	var list struct {
		Messages []QueuedMessage `json:"messages"`
	}
	getJSON(t, ts.URL+"/queue?route_key=route-a", http.StatusOK, &list)
	if len(list.Messages) != 1 || list.Messages[0].ID != created.Message.ID {
		t.Fatalf("list = %#v", list.Messages)
	}

	var detail QueuedMessage
	getJSON(t, ts.URL+"/queue/"+created.Message.ID, http.StatusOK, &detail)
	if detail.ID != created.Message.ID {
		t.Fatalf("detail = %#v", detail)
	}
}

func TestQueueAPIDeduplicatesExternalID(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var first struct {
		Message QueuedMessage `json:"message"`
	}
	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"first","source":"github","external_id":"delivery-1"}`, http.StatusAccepted, &first)

	var second struct {
		Message   QueuedMessage `json:"message"`
		Duplicate bool          `json:"duplicate"`
	}
	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"second","source":"github","external_id":"delivery-1"}`, http.StatusOK, &second)
	if !second.Duplicate || second.Message.ID != first.Message.ID {
		t.Fatalf("duplicate response = %#v first=%#v", second, first)
	}

	var list struct {
		Messages []QueuedMessage `json:"messages"`
	}
	getJSON(t, ts.URL+"/queue?route_key=route-a", http.StatusOK, &list)
	if len(list.Messages) != 1 {
		t.Fatalf("messages = %#v, want one deduped message", list.Messages)
	}
}

func TestQueueAPIValidation(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	postJSON(t, ts.URL+"/queue", `{"text":"hello"}`, http.StatusBadRequest, &map[string]string{})
	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":""}`, http.StatusBadRequest, &map[string]string{})
	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"hello","priority":-1}`, http.StatusBadRequest, &map[string]string{})
}

func TestQueueAPICapacity(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	s.mailboxStore = NewMailboxStore(1)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"one"}`, http.StatusAccepted, &map[string]any{})
	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"two"}`, http.StatusServiceUnavailable, &map[string]string{})
}

func TestQueueAPIClaimAckRelease(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"one"}`, http.StatusAccepted, &map[string]any{})

	var claimed struct {
		Messages []QueuedMessage `json:"messages"`
	}
	postJSON(t, ts.URL+"/queue/claim", `{"route_key":"route-a","limit":1}`, http.StatusOK, &claimed)
	if len(claimed.Messages) != 1 || claimed.Messages[0].ClaimID == "" {
		t.Fatalf("claimed = %#v", claimed.Messages)
	}
	msg := claimed.Messages[0]

	postJSON(t, ts.URL+"/queue/"+msg.ID+"/release", `{"claim_id":"`+msg.ClaimID+`"}`, http.StatusOK, &map[string]any{})
	postJSON(t, ts.URL+"/queue/claim", `{"route_key":"route-a","limit":1}`, http.StatusOK, &claimed)
	if len(claimed.Messages) != 1 || claimed.Messages[0].Attempt != 2 {
		t.Fatalf("reclaimed = %#v, want attempt 2", claimed.Messages)
	}
	msg = claimed.Messages[0]

	postJSON(t, ts.URL+"/queue/"+msg.ID+"/ack", `{"claim_id":"`+msg.ClaimID+`"}`, http.StatusOK, &map[string]any{})
	var detail QueuedMessage
	getJSON(t, ts.URL+"/queue/"+msg.ID, http.StatusOK, &detail)
	if detail.Status != QueueStatusAcknowledged {
		t.Fatalf("detail status = %q, want acknowledged", detail.Status)
	}
}

func TestQueueAPIOversizedText(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body, err := json.Marshal(map[string]string{
		"route_key": "route-a",
		"text":      strings.Repeat("x", maxQueueTextBytes+1),
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	resp, err := http.Post(ts.URL+"/queue", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("POST /queue: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", resp.StatusCode)
	}
}
