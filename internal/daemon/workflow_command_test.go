package daemon

import (
	"strings"
	"testing"
)

func TestParseWorkflowInvocation(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantType  string
		wantGoal  string
		wantError bool
	}{
		{name: "ordinary prompt ignored", text: "research this without slash"},
		{name: "research command", text: "/research compare Kocoro daemon", wantType: WorkflowTypeResearch, wantGoal: "compare Kocoro daemon"},
		{name: "research strategy", text: "/research deep compare Kocoro daemon", wantType: WorkflowTypeResearch, wantGoal: "compare Kocoro daemon"},
		{name: "research command trims goal", text: "  /research   compare Kocoro daemon  ", wantType: WorkflowTypeResearch, wantGoal: "compare Kocoro daemon"},
		{name: "swarm command", text: "/swarm plan phase six", wantType: WorkflowTypeSwarm, wantGoal: "plan phase six"},
		{name: "dag command", text: "/dag plan phase seven", wantType: WorkflowTypeAuto, wantGoal: "plan phase seven"},
		{name: "case-insensitive command", text: "/SWARM plan phase six", wantType: WorkflowTypeSwarm, wantGoal: "plan phase six"},
		{name: "unknown slash command ignored", text: "/unknown keep compatibility"},
		{name: "empty research goal", text: "/research", wantError: true},
		{name: "empty swarm goal", text: "/swarm   ", wantError: true},
		{name: "empty dag goal", text: "/dag   ", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWorkflowInvocation(tt.text)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseWorkflowInvocation() error = %v", err)
			}
			if tt.wantType == "" {
				if got != nil {
					t.Fatalf("workflow = %#v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("workflow = nil")
			}
			if got.Type != tt.wantType || got.Goal != tt.wantGoal {
				t.Fatalf("workflow = %#v, want type %q goal %q", got, tt.wantType, tt.wantGoal)
			}
			if got.Command == "" || got.RouteHint == "" || len(got.Steps) == 0 {
				t.Fatalf("workflow missing command, route, or steps: %#v", got)
			}
			if tt.text == "/research deep compare Kocoro daemon" && got.Strategy != "deep" {
				t.Fatalf("strategy = %q, want deep", got.Strategy)
			}
			if !strings.Contains(got.Prompt, tt.wantGoal) {
				t.Fatalf("prompt %q does not contain goal %q", got.Prompt, tt.wantGoal)
			}
		})
	}
}
