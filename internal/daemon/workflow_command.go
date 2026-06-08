package daemon

import (
	"fmt"
	"strings"

	"github.com/starclaw/starclaw/internal/cloudflow"
)

const (
	WorkflowTypeResearch = "research"
	WorkflowTypeSwarm    = "swarm"
	WorkflowTypeAuto     = "auto"
)

type workflowInvocation struct {
	Type      string
	Command   string
	Goal      string
	Strategy  string
	RouteHint string
	Prompt    string
	Steps     []WorkflowStepState
}

func parseWorkflowInvocation(text string) (*workflowInvocation, error) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return nil, nil
	}
	cmd := cloudflow.ParseSlash(trimmed)
	if cmd == nil {
		fields := strings.Fields(trimmed)
		if len(fields) == 1 {
			switch strings.ToLower(fields[0]) {
			case "/research":
				return nil, fmt.Errorf("/research goal is required")
			case "/swarm":
				return nil, fmt.Errorf("/swarm goal is required")
			case "/dag":
				return nil, fmt.Errorf("/dag goal is required")
			}
		}
		return nil, nil
	}
	switch cmd.Type {
	case cloudflow.TypeResearch:
		return newResearchWorkflow(cmd.Query, cmd.Strategy), nil
	case cloudflow.TypeSwarm:
		return newSwarmWorkflow(cmd.Query), nil
	case cloudflow.TypeAuto:
		return newAutoWorkflow(cmd.Query), nil
	default:
		return nil, nil
	}
}

func newResearchWorkflow(goal, strategy string) *workflowInvocation {
	if strategy == "" {
		strategy = "standard"
	}
	return &workflowInvocation{
		Type:      WorkflowTypeResearch,
		Command:   "/research",
		Goal:      goal,
		Strategy:  strategy,
		RouteHint: "research",
		Prompt: strings.TrimSpace(fmt.Sprintf(`Run this as a StarClaw research workflow.

Goal:
%s

Research strategy: %s

Produce an evidence-backed research brief. Separate verified facts, assumptions, options, open questions, and suggested follow-up checks. Use tools when local evidence is needed, keep claims scoped to available evidence, and preserve any required approval boundaries.`, goal, strategy)),
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

func newAutoWorkflow(goal string) *workflowInvocation {
	return &workflowInvocation{
		Type:      WorkflowTypeAuto,
		Command:   "/dag",
		Goal:      goal,
		RouteHint: "auto",
		Prompt: strings.TrimSpace(fmt.Sprintf(`Run this as a StarClaw auto-orchestration workflow.

Goal:
%s

Decompose the task into a clear dependency graph, identify which steps can run independently, execute the local steps available through StarClaw, and synthesize a final result with verification status and remaining blockers.`, goal)),
		Steps: []WorkflowStepState{
			{ID: "parse_command", Title: "Parse workflow command", Status: WorkflowStepCompleted, Sequence: 1},
			{ID: "dag_plan", Title: "Plan task graph", Status: WorkflowStepPlanned, Sequence: 2},
			{ID: "dag_execute", Title: "Execute local graph", Status: WorkflowStepPlanned, Sequence: 3},
			{ID: "dag_complete", Title: "Complete orchestration", Status: WorkflowStepPlanned, Sequence: 4},
		},
	}
}

func workflowMetadata(inv *workflowInvocation) map[string]any {
	if inv == nil {
		return nil
	}
	meta := map[string]any{
		"workflow":   inv.Type,
		"command":    inv.Command,
		"route_hint": inv.RouteHint,
	}
	if inv.Strategy != "" {
		meta["strategy"] = inv.Strategy
	}
	return meta
}
