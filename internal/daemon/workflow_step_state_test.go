package daemon

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowStepStateUpsertAndTransition(t *testing.T) {
	store := NewRunStore(10)
	store.Start(RunAgentRequest{RequestID: "workflow-run", Channel: ChannelHTTP})

	if !store.UpsertStep("workflow-run", WorkflowStepState{
		ID:       "step-2",
		Title:    "Second step",
		Sequence: 2,
		Metadata: map[string]any{"kind": "analysis"},
	}) {
		t.Fatal("expected step upsert")
	}
	if !store.UpsertStep("workflow-run", WorkflowStepState{
		ID:       "step-1",
		Title:    "First step",
		Sequence: 1,
	}) {
		t.Fatal("expected step upsert")
	}
	if !store.TransitionStep("workflow-run", "step-1", WorkflowStepRunning, map[string]any{"phase": "started"}) {
		t.Fatal("expected running transition")
	}
	if !store.TransitionStep("workflow-run", "step-1", WorkflowStepCompleted, map[string]any{"result": "ok"}) {
		t.Fatal("expected completed transition")
	}

	record, ok := store.Get("workflow-run")
	if !ok {
		t.Fatal("expected run record")
	}
	if record.Status != "running" {
		t.Fatalf("run status = %q, want running", record.Status)
	}
	if len(record.Steps) != 2 {
		t.Fatalf("steps = %d, want 2", len(record.Steps))
	}
	if record.Steps[0].ID != "step-1" || record.Steps[1].ID != "step-2" {
		t.Fatalf("step order = %#v, want sequence order", record.Steps)
	}
	step := record.Steps[0]
	if step.Status != WorkflowStepCompleted || step.Attempt != 1 || step.StartedAt == nil || step.EndedAt == nil {
		t.Fatalf("step = %#v, want completed with attempt/start/end", step)
	}
	if step.Metadata["phase"] != "started" || step.Metadata["result"] != "ok" {
		t.Fatalf("metadata = %#v, want merged transition metadata", step.Metadata)
	}
}

func TestWorkflowStepStatePersistsAndRecovers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runs.json")
	store, err := NewPersistentRunStore(10, path)
	if err != nil {
		t.Fatalf("new persistent store: %v", err)
	}
	store.Start(RunAgentRequest{RequestID: "workflow-persist", Channel: ChannelHTTP})
	store.UpsertStep("workflow-persist", WorkflowStepState{ID: "plan", Title: "Plan", Status: WorkflowStepRunning})
	store.TransitionStep("workflow-persist", "plan", WorkflowStepWaitingApproval, map[string]any{"approval": "operator"})

	recovered, err := NewPersistentRunStore(10, path)
	if err != nil {
		t.Fatalf("recover persistent store: %v", err)
	}
	record, ok := recovered.Get("workflow-persist")
	if !ok {
		t.Fatal("expected recovered run")
	}
	if len(record.Steps) != 1 {
		t.Fatalf("steps = %d, want 1", len(record.Steps))
	}
	step := record.Steps[0]
	if step.ID != "plan" || step.Status != WorkflowStepWaitingApproval || step.StartedAt == nil {
		t.Fatalf("recovered step = %#v, want waiting_approval with start time", step)
	}

	recovered.TransitionStep("workflow-persist", "plan", WorkflowStepCompleted, nil)
	record, ok = recovered.Get("workflow-persist")
	if !ok {
		t.Fatal("expected recovered run")
	}
	last := record.StructuredEvents[len(record.StructuredEvents)-1]
	if last.ID != "workflow-persist-000004" {
		t.Fatalf("event id after recovery = %q, want workflow-persist-000004", last.ID)
	}
}

func TestWorkflowStepEventsRedactMetadataAndMetricsAggregateOnly(t *testing.T) {
	store := NewRunStore(10)
	store.Start(RunAgentRequest{RequestID: "workflow-redact", Text: "secret prompt body", Channel: ChannelHTTP})
	store.UpsertStep("workflow-redact", WorkflowStepState{
		ID:     "danger",
		Status: WorkflowStepRunning,
		Metadata: map[string]any{
			"prompt": "secret prompt body",
			"args":   `{"api_key":"sk-secret"}`,
			"token":  "sk-secret",
			"safe":   "ok",
		},
	})

	record, ok := store.Get("workflow-redact")
	if !ok {
		t.Fatal("expected run record")
	}
	encodedRecord, err := json.Marshal(record.Steps)
	if err != nil {
		t.Fatalf("marshal steps: %v", err)
	}
	for _, forbidden := range []string{"secret prompt body", "sk-secret"} {
		if strings.Contains(string(encodedRecord), forbidden) {
			t.Fatalf("workflow step state leaked %q: %s", forbidden, encodedRecord)
		}
	}
	last := record.StructuredEvents[len(record.StructuredEvents)-1]
	if last.Type != "workflow_step" || last.Phase != "workflow" {
		t.Fatalf("last event = %#v, want workflow_step/workflow", last)
	}
	encoded, err := json.Marshal(last)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	body := string(encoded)
	for _, forbidden := range []string{"secret prompt body", "sk-secret"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("workflow step event leaked %q: %s", forbidden, body)
		}
	}

	metrics := store.Metrics()
	events := metrics["events_by_type"].(map[string]int)
	if events["workflow_step"] != 1 {
		t.Fatalf("workflow_step event count = %d, want 1", events["workflow_step"])
	}
	encodedMetrics, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("marshal metrics: %v", err)
	}
	if strings.Contains(string(encodedMetrics), "danger") || strings.Contains(string(encodedMetrics), "secret prompt body") {
		t.Fatalf("metrics leaked step metadata: %s", encodedMetrics)
	}
}
