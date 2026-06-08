package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/client"
)

func TestTokenBudgetDecision(t *testing.T) {
	tests := []struct {
		name            string
		budget          TokenBudget
		usages          []client.Usage
		projectedInput  int
		projectedOutput int
		wantStop        bool
		wantStatus      string
		wantUnknown     int
		wantDetailPart  string
	}{
		{
			name:            "under budget",
			budget:          TokenBudget{MaxTotalTokens: 100, HardStop: true},
			usages:          []client.Usage{{InputTokens: 20, OutputTokens: 10}},
			projectedOutput: 20,
			wantStatus:      TokenBudgetStatusOK,
		},
		{
			name:            "at budget with projection",
			budget:          TokenBudget{MaxTotalTokens: 100, HardStop: true},
			usages:          []client.Usage{{InputTokens: 50, OutputTokens: 25}},
			projectedOutput: 25,
			wantStop:        true,
			wantStatus:      TokenBudgetStatusExhausted,
			wantDetailPart:  "total token budget exhausted",
		},
		{
			name:           "over output budget",
			budget:         TokenBudget{MaxOutputTokens: 10, HardStop: true},
			usages:         []client.Usage{{InputTokens: 1, OutputTokens: 11}},
			wantStop:       true,
			wantStatus:     TokenBudgetStatusExhausted,
			wantDetailPart: "output token budget exhausted",
		},
		{
			name:           "unknown usage",
			budget:         TokenBudget{MaxTotalTokens: 100, HardStop: true},
			usages:         []client.Usage{{}},
			wantStatus:     TokenBudgetStatusUnknown,
			wantUnknown:    1,
			wantDetailPart: "provider did not return",
		},
		{
			name:           "projected input exceeds budget before usage",
			budget:         TokenBudget{MaxInputTokens: 1, HardStop: true},
			projectedInput: 2,
			wantStop:       true,
			wantStatus:     TokenBudgetStatusExhausted,
			wantDetailPart: "input token budget exhausted",
		},
		{
			name:       "soft budget does not stop",
			budget:     TokenBudget{MaxTotalTokens: 10, HardStop: false},
			usages:     []client.Usage{{InputTokens: 10, OutputTokens: 10}},
			wantStatus: TokenBudgetStatusExhausted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := newTokenBudgetTracker(tt.budget)
			for _, usage := range tt.usages {
				tracker.AddUsage(usage)
			}
			if tt.projectedInput > 0 {
				tracker.SetProjectedInput(tt.projectedInput)
			}
			decision := tracker.Decision(tt.projectedOutput)
			if decision.Stop != tt.wantStop {
				t.Fatalf("Stop = %v, want %v", decision.Stop, tt.wantStop)
			}
			if decision.Status.Status != tt.wantStatus {
				t.Fatalf("Status = %q, want %q; detail=%q", decision.Status.Status, tt.wantStatus, decision.Status.Detail)
			}
			if decision.Status.UnknownTurns != tt.wantUnknown {
				t.Fatalf("UnknownTurns = %d, want %d", decision.Status.UnknownTurns, tt.wantUnknown)
			}
			if tt.wantDetailPart != "" && !strings.Contains(decision.Status.Detail, tt.wantDetailPart) {
				t.Fatalf("Detail = %q, want substring %q", decision.Status.Detail, tt.wantDetailPart)
			}
		})
	}
}

type budgetSequenceClient struct {
	responses []*client.Response
	calls     int
}

func (c *budgetSequenceClient) Chat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions) (*client.Response, error) {
	c.calls++
	if c.calls > len(c.responses) {
		return &client.Response{Content: "unexpected extra call"}, nil
	}
	return c.responses[c.calls-1], nil
}

func TestAgentLoop_BudgetHardStopBeforeFollowupModelCall(t *testing.T) {
	llm := &budgetSequenceClient{
		responses: []*client.Response{
			{
				Content: "use tool",
				ToolUses: []client.ToolUse{{
					ID:    "toolu_1",
					Name:  "echo",
					Input: []byte(`{}`),
				}},
				Usage: client.Usage{InputTokens: 10, OutputTokens: 5},
			},
			{Content: "should not be called"},
		},
	}
	registry := NewToolRegistry()
	registry.Register(TestTool("echo"))
	loop := NewAgentLoop(llm, registry)
	loop.SetMaxTokens(10)
	loop.SetTokenBudget(TokenBudget{MaxTotalTokens: 20, HardStop: true})

	resp, err := loop.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if llm.calls != 1 {
		t.Fatalf("Chat calls = %d, want 1", llm.calls)
	}
	if resp.StopReason != RunStatusBudgetExhausted {
		t.Fatalf("StopReason = %q, want %q", resp.StopReason, RunStatusBudgetExhausted)
	}
	if loop.LastRunStatus().Code != RunStatusBudgetExhausted {
		t.Fatalf("LastRunStatus = %#v, want budget exhausted", loop.LastRunStatus())
	}
	status := loop.LastBudgetStatus()
	if status.Status != TokenBudgetStatusExhausted {
		t.Fatalf("Budget status = %#v, want exhausted", status)
	}
}

func TestAgentLoop_BudgetHardStopBeforeInitialModelCall(t *testing.T) {
	llm := &budgetSequenceClient{}
	loop := NewAgentLoop(llm, NewToolRegistry())
	loop.SetTokenBudget(TokenBudget{MaxInputTokens: 1, HardStop: true})

	resp, err := loop.Run(context.Background(), "this prompt is intentionally longer than one estimated token")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if llm.calls != 0 {
		t.Fatalf("Chat calls = %d, want 0", llm.calls)
	}
	if resp.StopReason != RunStatusBudgetExhausted {
		t.Fatalf("StopReason = %q, want %q", resp.StopReason, RunStatusBudgetExhausted)
	}
	status := loop.LastBudgetStatus()
	if status.Status != TokenBudgetStatusExhausted {
		t.Fatalf("Budget status = %#v, want exhausted", status)
	}
	if !strings.Contains(status.Detail, "input token budget exhausted") {
		t.Fatalf("Detail = %q, want input budget exhausted", status.Detail)
	}
}
