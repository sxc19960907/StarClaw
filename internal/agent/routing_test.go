package agent

import (
	"errors"
	"testing"
)

func TestRecommendRoute(t *testing.T) {
	tests := []struct {
		name       string
		input      RoutingInput
		complexity string
		route      string
		modelTier  string
	}{
		{
			name:       "simple prompt",
			input:      RoutingInput{Prompt: "Summarize this note"},
			complexity: ComplexitySimple,
			route:      RouteDirect,
			modelTier:  "small",
		},
		{
			name:       "evidence heavy prompt",
			input:      RoutingInput{Prompt: "Research this claim and cite sources"},
			complexity: ComplexityEvidenceHeavy,
			route:      RouteResearch,
			modelTier:  "medium",
		},
		{
			name:       "council worthy prompt",
			input:      RoutingInput{Prompt: "Run a council tradeoff review for this architecture decision"},
			complexity: ComplexityCouncilWorthy,
			route:      RouteCouncil,
			modelTier:  "high",
		},
		{
			name:       "delivery sensitive prompt",
			input:      RoutingInput{Prompt: "Send the final update email to the customer"},
			complexity: ComplexityDeliverySensitive,
			route:      RouteDelivery,
			modelTier:  "medium",
		},
		{
			name:       "budget constrained prompt",
			input:      RoutingInput{Prompt: "Research and cite sources", TokenBudget: TokenBudget{MaxTotalTokens: 100, HardStop: true}},
			complexity: ComplexityBudgetConstrained,
			route:      RouteBudget,
			modelTier:  "small",
		},
		{
			name:       "tool requested evidence",
			input:      RoutingInput{Prompt: "Find the answer", RequestedTools: []string{"grep", "file_read"}},
			complexity: ComplexityEvidenceHeavy,
			route:      RouteResearch,
			modelTier:  "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RecommendRoute(tt.input)
			if got.Complexity != tt.complexity {
				t.Fatalf("Complexity = %q, want %q", got.Complexity, tt.complexity)
			}
			if got.Route != tt.route {
				t.Fatalf("Route = %q, want %q", got.Route, tt.route)
			}
			if got.ModelTier != tt.modelTier {
				t.Fatalf("ModelTier = %q, want %q", got.ModelTier, tt.modelTier)
			}
			if got.Reason == "" {
				t.Fatal("Reason should be set")
			}
		})
	}
}

func TestRecommendFallback(t *testing.T) {
	tests := []struct {
		name      string
		input     FallbackInput
		wantNil   bool
		reason    string
		route     string
		modelTier string
	}{
		{
			name:      "provider error",
			input:     FallbackInput{ProviderError: errors.New("provider unavailable"), CurrentRoute: RouteResearch},
			reason:    FallbackProviderError,
			route:     RouteDirect,
			modelTier: "small",
		},
		{
			name: "budget exhaustion",
			input: FallbackInput{BudgetStatus: TokenBudgetUsage{
				Status: TokenBudgetStatusExhausted,
				Detail: "total token budget exhausted",
			}},
			reason:    FallbackBudgetExhausted,
			route:     RouteBudget,
			modelTier: "small",
		},
		{
			name:      "repeated failure",
			input:     FallbackInput{PreviousFailures: 2, CurrentRoute: RouteDirect},
			reason:    FallbackRepeatedFailure,
			route:     RouteCouncil,
			modelTier: "high",
		},
		{
			name:    "no fallback",
			input:   FallbackInput{CurrentRoute: RouteDirect},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RecommendFallback(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("fallback = %#v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("fallback = nil")
			}
			if got.Reason != tt.reason {
				t.Fatalf("Reason = %q, want %q", got.Reason, tt.reason)
			}
			if got.Route != tt.route {
				t.Fatalf("Route = %q, want %q", got.Route, tt.route)
			}
			if got.ModelTier != tt.modelTier {
				t.Fatalf("ModelTier = %q, want %q", got.ModelTier, tt.modelTier)
			}
		})
	}
}
