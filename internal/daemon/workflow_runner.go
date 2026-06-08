package daemon

import (
	"context"

	"github.com/starclaw/starclaw/internal/agent"
)

func (s *Server) runWorkflowAgent(ctx context.Context, req RunAgentRequest, inv *workflowInvocation, handler agent.EventHandler) (RunAgentResponse, error) {
	if inv == nil {
		return s.runAgent(ctx, req, handler)
	}

	metadata := workflowMetadata(inv)
	for _, step := range inv.Steps {
		step.Metadata = metadata
		s.runStore.UpsertStep(req.RequestID, step)
	}

	switch inv.Type {
	case WorkflowTypeResearch:
		s.runStore.TransitionStep(req.RequestID, "research_plan", WorkflowStepRunning, metadata)
		s.runStore.TransitionStep(req.RequestID, "research_plan", WorkflowStepCompleted, metadata)
		s.runStore.TransitionStep(req.RequestID, "research_execute", WorkflowStepRunning, metadata)
	case WorkflowTypeSwarm:
		s.runStore.TransitionStep(req.RequestID, "role_plan", WorkflowStepRunning, metadata)
		s.runStore.TransitionStep(req.RequestID, "role_plan", WorkflowStepCompleted, metadata)
		s.runStore.TransitionStep(req.RequestID, "synthesis_handoff", WorkflowStepRunning, metadata)
		s.runStore.TransitionStep(req.RequestID, "synthesis_handoff", WorkflowStepCompleted, metadata)
		s.runStore.TransitionStep(req.RequestID, "swarm_execute", WorkflowStepRunning, metadata)
	}

	result, err := s.runAgent(ctx, req, handler)
	s.finishWorkflowSteps(req.RequestID, inv, err, result.Error != "")
	return result, err
}

func (s *Server) finishWorkflowSteps(runID string, inv *workflowInvocation, err error, responseError bool) {
	if inv == nil {
		return
	}
	status := WorkflowStepCompleted
	if err != nil || responseError {
		status = WorkflowStepFailed
	}
	metadata := workflowMetadata(inv)
	switch inv.Type {
	case WorkflowTypeResearch:
		s.runStore.TransitionStep(runID, "research_execute", status, metadata)
		s.runStore.TransitionStep(runID, "research_complete", status, metadata)
	case WorkflowTypeSwarm:
		s.runStore.TransitionStep(runID, "swarm_execute", status, metadata)
		s.runStore.TransitionStep(runID, "swarm_complete", status, metadata)
	}
}
