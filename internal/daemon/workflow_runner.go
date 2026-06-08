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
	case WorkflowTypeAuto:
		s.runStore.TransitionStep(req.RequestID, "dag_plan", WorkflowStepRunning, metadata)
		s.runStore.TransitionStep(req.RequestID, "dag_plan", WorkflowStepCompleted, metadata)
		s.runStore.TransitionStep(req.RequestID, "dag_execute", WorkflowStepRunning, metadata)
	}

	result, err := s.runAgent(ctx, req, handler)
	applyWorkflowRouting(inv, &result)
	s.finishWorkflowSteps(req.RequestID, inv, err, result.Error != "")
	return result, err
}

func applyWorkflowRouting(inv *workflowInvocation, result *RunAgentResponse) {
	if inv == nil || result == nil || inv.RouteHint == "" {
		return
	}
	if result.Routing == nil {
		result.Routing = &agent.RouteRecommendation{}
	}
	switch inv.RouteHint {
	case "research":
		result.Routing.Complexity = agent.ComplexityEvidenceHeavy
		result.Routing.Route = agent.RouteResearch
		result.Routing.ModelTier = "medium"
	case "council":
		result.Routing.Complexity = agent.ComplexityCouncilWorthy
		result.Routing.Route = agent.RouteCouncil
		result.Routing.ModelTier = "high"
	case "auto":
		result.Routing.Complexity = agent.ComplexityCouncilWorthy
		result.Routing.Route = agent.RouteCouncil
		result.Routing.ModelTier = "high"
	default:
		result.Routing.Route = inv.RouteHint
	}
	result.Routing.Reason = "workflow command route hint"
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
	case WorkflowTypeAuto:
		s.runStore.TransitionStep(runID, "dag_execute", status, metadata)
		s.runStore.TransitionStep(runID, "dag_complete", status, metadata)
	}
}
