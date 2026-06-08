package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
)

func TestPersistentRunStoreRecoversRunRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.json")
	store, err := NewPersistentRunStore(10, path)
	if err != nil {
		t.Fatalf("new persistent store: %v", err)
	}
	store.Start(RunAgentRequest{RequestID: "persist-run", Text: "remember this prompt", Agent: "helper", Channel: ChannelHTTP, Source: "test"})
	store.AddEvent("persist-run", EventUsage, map[string]any{"input_tokens": 3, "output_tokens": 4})
	store.Complete("persist-run", RunAgentResponse{
		SessionID:    "sess-1",
		Usage:        map[string]int{"input_tokens": 3, "output_tokens": 4, "total_tokens": 7},
		BudgetStatus: &agent.TokenBudgetUsage{Status: agent.TokenBudgetStatusOK, InputTokens: 3, OutputTokens: 4, TotalTokens: 7},
		Routing:      &agent.RouteRecommendation{Complexity: agent.ComplexitySimple, Route: agent.RouteDirect, ModelTier: "small", Reason: "test"},
	}, nil)

	recovered, err := NewPersistentRunStore(10, path)
	if err != nil {
		t.Fatalf("recover persistent store: %v", err)
	}
	record, ok := recovered.Get("persist-run")
	if !ok {
		t.Fatal("expected recovered run")
	}
	if record.Status != "completed" || record.SessionID != "sess-1" {
		t.Fatalf("recovered record = %#v, want completed sess-1", record)
	}
	if record.Usage["input_tokens"] != 3 || record.Budget == nil || record.Routing == nil {
		t.Fatalf("recovered metadata = usage %#v budget %#v routing %#v", record.Usage, record.Budget, record.Routing)
	}
	if len(record.StructuredEvents) == 0 || len(record.Events) == 0 {
		t.Fatalf("recovered events = structured %d legacy %d, want both", len(record.StructuredEvents), len(record.Events))
	}
}

func TestPersistentRunStoreRecoversControlDecisionsAndEventSequence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.json")
	store, err := NewPersistentRunStore(10, path)
	if err != nil {
		t.Fatalf("new persistent store: %v", err)
	}
	store.Start(RunAgentRequest{RequestID: "control-run", Channel: ChannelHTTP})
	if !store.AddControlDecision("control-run", RunControlDecision{Action: "replay", Status: "approval_required", Reason: "review first"}) {
		t.Fatal("expected control decision")
	}

	recovered, err := NewPersistentRunStore(10, path)
	if err != nil {
		t.Fatalf("recover persistent store: %v", err)
	}
	recovered.AddEvent("control-run", EventRunStatus, map[string]any{"status": "reviewed"})
	record, ok := recovered.Get("control-run")
	if !ok {
		t.Fatal("expected recovered run")
	}
	if len(record.Control) != 1 || record.Control[0].Action != "replay" || record.Control[0].Status != "approval_required" {
		t.Fatalf("control = %#v, want replay approval_required", record.Control)
	}
	if len(record.StructuredEvents) != 3 {
		t.Fatalf("structured events = %d, want 3", len(record.StructuredEvents))
	}
	if got := record.StructuredEvents[2].ID; got != "control-run-000003" {
		t.Fatalf("event id after recovery = %q, want control-run-000003", got)
	}
}

func TestPersistentRunStoreCorruptFileReturnsSafeEmptyStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.json")
	if err := os.WriteFile(path, []byte(`{"records":`), 0o600); err != nil {
		t.Fatalf("write corrupt store: %v", err)
	}

	store, err := NewPersistentRunStore(10, path)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if len(store.List()) != 0 {
		t.Fatalf("recovered runs = %d, want 0", len(store.List()))
	}
	if store.PersistError() == nil {
		t.Fatal("expected persisted load error to be recorded")
	}
}

func TestPersistentRunStoreEnforcesLimitOnLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.json")
	store, err := NewPersistentRunStore(3, path)
	if err != nil {
		t.Fatalf("new persistent store: %v", err)
	}
	for _, id := range []string{"run-1", "run-2", "run-3", "run-4", "run-5"} {
		store.Start(RunAgentRequest{RequestID: id, Channel: ChannelHTTP})
	}

	recovered, err := NewPersistentRunStore(3, path)
	if err != nil {
		t.Fatalf("recover persistent store: %v", err)
	}
	summaries := recovered.List()
	if len(summaries) != 3 {
		t.Fatalf("summaries = %d, want 3", len(summaries))
	}
	for _, id := range []string{"run-5", "run-4", "run-3"} {
		if _, ok := recovered.Get(id); !ok {
			t.Fatalf("expected %s to be retained", id)
		}
	}
	for _, id := range []string{"run-2", "run-1"} {
		if _, ok := recovered.Get(id); ok {
			t.Fatalf("expected %s to be pruned", id)
		}
	}
}
