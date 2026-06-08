package agent

import (
	"strings"
)

const (
	ComplexitySimple            = "simple"
	ComplexityEvidenceHeavy     = "evidence_heavy"
	ComplexityCouncilWorthy     = "council_worthy"
	ComplexityDeliverySensitive = "delivery_sensitive"
	ComplexityBudgetConstrained = "budget_constrained"

	RouteDirect   = "direct"
	RouteResearch = "research"
	RouteCouncil  = "council"
	RouteDelivery = "delivery_review"
	RouteBudget   = "budget_guard"

	FallbackProviderError   = "provider_error"
	FallbackBudgetExhausted = "budget_exhausted"
	FallbackRepeatedFailure = "repeated_failure"
)

// RoutingInput contains local-only signals used to classify a run before any
// provider call is made.
type RoutingInput struct {
	Prompt           string
	RequestedTools   []string
	TokenBudget      TokenBudget
	PreviousFailures int
}

// RouteRecommendation describes a deterministic route/model recommendation.
type RouteRecommendation struct {
	Complexity string `json:"complexity"`
	Route      string `json:"route"`
	ModelTier  string `json:"model_tier"`
	Reason     string `json:"reason"`
}

// FallbackInput contains local failure signals used to recommend a next route.
type FallbackInput struct {
	ProviderError    error
	BudgetStatus     TokenBudgetUsage
	PreviousFailures int
	CurrentRoute     string
}

// FallbackDecision describes a transparent fallback recommendation.
type FallbackDecision struct {
	Reason    string `json:"reason"`
	Route     string `json:"route"`
	ModelTier string `json:"model_tier"`
	Detail    string `json:"detail,omitempty"`
}

// RecommendRoute classifies prompt shape and local constraints without making
// remote calls.
func RecommendRoute(input RoutingInput) RouteRecommendation {
	prompt := strings.ToLower(input.Prompt)
	tools := strings.ToLower(strings.Join(input.RequestedTools, " "))

	switch {
	case input.TokenBudget.Enabled() && input.TokenBudget.HardStop:
		return RouteRecommendation{Complexity: ComplexityBudgetConstrained, Route: RouteBudget, ModelTier: "small", Reason: "hard token budget is configured"}
	case containsAny(prompt, "send", "email", "post", "publish", "deploy", "payment", "submit", "delete", "external", "notify"):
		return RouteRecommendation{Complexity: ComplexityDeliverySensitive, Route: RouteDelivery, ModelTier: "medium", Reason: "prompt includes external or high-risk delivery language"}
	case containsAny(prompt, "council", "debate", "multi-agent", "tradeoff", "architecture decision", "strategy", "risk review"):
		return RouteRecommendation{Complexity: ComplexityCouncilWorthy, Route: RouteCouncil, ModelTier: "high", Reason: "prompt asks for multi-perspective reasoning"}
	case containsAny(prompt, "research", "cite", "citation", "evidence", "sources", "compare", "benchmark", "verify") ||
		containsAny(tools, "web", "http", "grep", "file_read"):
		return RouteRecommendation{Complexity: ComplexityEvidenceHeavy, Route: RouteResearch, ModelTier: "medium", Reason: "prompt or tools indicate evidence gathering"}
	default:
		return RouteRecommendation{Complexity: ComplexitySimple, Route: RouteDirect, ModelTier: "small", Reason: "prompt can use direct execution"}
	}
}

// RecommendFallback maps local failure states to a deterministic next route.
func RecommendFallback(input FallbackInput) *FallbackDecision {
	if input.BudgetStatus.Status == TokenBudgetStatusExhausted {
		return &FallbackDecision{
			Reason:    FallbackBudgetExhausted,
			Route:     RouteBudget,
			ModelTier: "small",
			Detail:    input.BudgetStatus.Detail,
		}
	}
	if input.ProviderError != nil {
		return &FallbackDecision{
			Reason:    FallbackProviderError,
			Route:     RouteDirect,
			ModelTier: "small",
			Detail:    input.ProviderError.Error(),
		}
	}
	if input.PreviousFailures >= 2 {
		return &FallbackDecision{
			Reason:    FallbackRepeatedFailure,
			Route:     RouteCouncil,
			ModelTier: "high",
			Detail:    "same-class run failed repeatedly",
		}
	}
	return nil
}

func containsAny(s string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
