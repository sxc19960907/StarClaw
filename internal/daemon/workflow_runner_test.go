package daemon

import (
	"context"
	"testing"
)

func TestWorkflowRunnerRecordsResearchSteps(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	runID := "workflow-research-runner"
	req := RunAgentRequest{Text: "/research compare daemon", RequestID: runID, Channel: ChannelHTTP, Source: "test"}
	inv, err := parseWorkflowInvocation(req.Text)
	if err != nil {
		t.Fatalf("parse workflow: %v", err)
	}
	req.Text = inv.Prompt
	s.runStore.Start(req)

	result, runErr := s.runWorkflowAgent(context.Background(), req, inv, nil)
	s.runStore.Complete(runID, result, runErr)
	if runErr != nil {
		t.Fatalf("runWorkflowAgent() error = %v", runErr)
	}

	record, ok := s.runStore.Get(runID)
	if !ok {
		t.Fatal("expected run record")
	}
	assertWorkflowStepStatus(t, record, "parse_command", WorkflowStepCompleted)
	assertWorkflowStepStatus(t, record, "research_plan", WorkflowStepCompleted)
	assertWorkflowStepStatus(t, record, "research_execute", WorkflowStepCompleted)
	assertWorkflowStepStatus(t, record, "research_complete", WorkflowStepCompleted)
}

func TestWorkflowRunnerRecordsFailedSteps(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	runID := "workflow-failed-runner"
	req := RunAgentRequest{Text: "/swarm ", RequestID: runID, Channel: ChannelHTTP, Source: "test"}
	inv := newSwarmWorkflow("force failure")
	req.Text = ""
	s.runStore.Start(req)

	result, runErr := s.runWorkflowAgent(context.Background(), req, inv, nil)
	s.runStore.Complete(runID, result, runErr)
	if runErr == nil {
		t.Fatal("expected run error")
	}

	record, ok := s.runStore.Get(runID)
	if !ok {
		t.Fatal("expected run record")
	}
	assertWorkflowStepStatus(t, record, "swarm_execute", WorkflowStepFailed)
	assertWorkflowStepStatus(t, record, "swarm_complete", WorkflowStepFailed)
}

func assertWorkflowStepStatus(t *testing.T, record *RunRecord, id, status string) {
	t.Helper()
	for _, step := range record.Steps {
		if step.ID != id {
			continue
		}
		if step.Status != status {
			t.Fatalf("step %q status = %q, want %q", id, step.Status, status)
		}
		if step.Metadata["workflow"] == "" || step.Metadata["command"] == "" || step.Metadata["route_hint"] == "" {
			t.Fatalf("step %q metadata = %#v, want workflow command metadata", id, step.Metadata)
		}
		return
	}
	t.Fatalf("step %q not found in %#v", id, record.Steps)
}
