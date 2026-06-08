package daemon

import (
	"fmt"
	"strings"
)

const (
	WorkflowTypeResearch = "research"
	WorkflowTypeSwarm    = "swarm"
)

type workflowInvocation struct {
	Type      string
	Command   string
	Goal      string
	RouteHint string
	Prompt    string
	Steps     []WorkflowStepState
}

func parseWorkflowInvocation(text string) (*workflowInvocation, error) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return nil, nil
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return nil, nil
	}
	command := strings.ToLower(fields[0])
	goal := strings.TrimSpace(strings.TrimPrefix(trimmed, fields[0]))

	switch command {
	case "/research":
		if goal == "" {
			return nil, fmt.Errorf("/research goal is required")
		}
		return newResearchWorkflow(goal), nil
	case "/swarm":
		if goal == "" {
			return nil, fmt.Errorf("/swarm goal is required")
		}
		return newSwarmWorkflow(goal), nil
	default:
		return nil, nil
	}
}

func newResearchWorkflow(goal string) *workflowInvocation {
	return &workflowInvocation{
		Type:      WorkflowTypeResearch,
		Command:   "/research",
		Goal:      goal,
		RouteHint: "research",
		Prompt: strings.TrimSpace(fmt.Sprintf(`Run this as a StarClaw research workflow.

Goal:
%s

Produce an evidence-backed research brief. Separate verified facts, assumptions, options, open questions, and suggested follow-up checks. Use tools when local evidence is needed, keep claims scoped to available evidence, and preserve any required approval boundaries.`, goal)),
		Steps: []WorkflowStepState{
			{ID: "parse_command", Title: "Parse workflow command", Status: WorkflowStepCompleted, Sequence: 1},
			{ID: "research_plan", Title: "Plan research approach", Status: WorkflowStepPlanned, Sequence: 2},
			{ID: "research_execute", Title: "Execute research workflow", Status: WorkflowStepPlanned, Sequence: 3},
			{ID: "research_complete", Title: "Complete research brief", Status: WorkflowStepPlanned, Sequence: 4},
		},
	}
}

func newSwarmWorkflow(goal string) *workflowInvocation {
	return &workflowInvocation{
		Type:      WorkflowTypeSwarm,
		Command:   "/swarm",
		Goal:      goal,
		RouteHint: "council",
		Prompt: strings.TrimSpace(fmt.Sprintf(`Run this as a StarClaw multi-agent swarm workflow.

Goal:
%s

Coordinate planner, researcher, and reviewer perspectives. Keep role contributions distinct, synthesize the concrete execution path, call out risks and verification steps, and preserve any required approval boundaries before taking irreversible action.`, goal)),
		Steps: []WorkflowStepState{
			{ID: "parse_command", Title: "Parse workflow command", Status: WorkflowStepCompleted, Sequence: 1},
			{ID: "role_plan", Title: "Plan role contributions", Status: WorkflowStepPlanned, Sequence: 2},
			{ID: "synthesis_handoff", Title: "Synthesize swarm handoff", Status: WorkflowStepPlanned, Sequence: 3},
			{ID: "swarm_execute", Title: "Execute swarm workflow", Status: WorkflowStepPlanned, Sequence: 4},
			{ID: "swarm_complete", Title: "Complete swarm synthesis", Status: WorkflowStepPlanned, Sequence: 5},
		},
	}
}

func workflowMetadata(inv *workflowInvocation) map[string]any {
	if inv == nil {
		return nil
	}
	return map[string]any{
		"workflow":   inv.Type,
		"command":    inv.Command,
		"route_hint": inv.RouteHint,
	}
}
