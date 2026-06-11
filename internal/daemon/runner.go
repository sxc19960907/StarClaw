package daemon

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/agents"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/session"
	"github.com/starclaw/starclaw/internal/tools"
)

// RunAgent executes an agent turn based on the request.
// For named agents, loads the agent config and resumes its session.
// For the default agent (empty Agent), creates a new session per run.
// It runs the agent loop with the provided tools, saves the session
// after completion, and forwards events (tool calls, results, text)
// to the handler.
func RunAgent(ctx context.Context, deps *ServerDeps, req RunAgentRequest, handler agent.EventHandler) (RunAgentResponse, error) {
	return RunAgentWithApproval(ctx, deps, req, handler, nil)
}

// RunAgentWithApproval executes an agent turn and optionally wires tool approval.
func RunAgentWithApproval(ctx context.Context, deps *ServerDeps, req RunAgentRequest, handler agent.EventHandler, approver agent.ApprovalRequester) (RunAgentResponse, error) {
	// --- Validation ---
	if strings.TrimSpace(req.Text) == "" {
		return RunAgentResponse{}, fmt.Errorf("text is required")
	}
	if deps == nil {
		return RunAgentResponse{}, fmt.Errorf("deps is required")
	}
	if deps.Registry == nil {
		return RunAgentResponse{}, fmt.Errorf("registry not configured in deps")
	}
	ctx = tools.WithBrowserUseLease(ctx)
	defer tools.BrowserUseLeaseFrom(ctx).ReleaseOnly()

	// --- Agent resolution ---
	agentName := strings.TrimSpace(req.Agent)
	var agentCfg *agents.Agent
	if agentName != "" {
		var err error
		agentCfg, err = agents.LoadAgent(deps.AgentsDir, agentName)
		if err != nil {
			return RunAgentResponse{}, fmt.Errorf("failed to load agent %q: %w", agentName, err)
		}
	}
	effectiveCfg := effectiveRunConfig(deps, agentCfg)
	llmClient := deps.LLMClient
	if deps.LLMClientFactory != nil {
		llmClient = deps.LLMClientFactory(effectiveCfg)
	}
	if llmClient == nil {
		return RunAgentResponse{}, fmt.Errorf("LLM client not configured in deps")
	}
	registry := filteredRegistry(deps.Registry, effectiveCfg.Tools)
	routing := agent.RecommendRoute(agent.RoutingInput{
		Prompt:      req.Text,
		TokenBudget: agent.TokenBudgetFromAgent(effectiveCfg.Agent),
	})

	// --- Session setup ---
	sessionsDir := sessionsDirFor(deps, agentName)
	sessMgr := session.NewManager(sessionsDir)

	var sess *session.Session
	switch {
	case req.SessionID != "":
		// Resume a specific session by ID.
		if _, err := sessMgr.Resume(req.SessionID); err != nil {
			return RunAgentResponse{}, fmt.Errorf("session not found: %s", req.SessionID)
		}
		sess = sessMgr.Current()
	case req.NewSession:
		// Force a new session.
		sess = sessMgr.NewSession()
	case agentName != "":
		// Named agent: try to resume the latest session so conversations
		// persist across runs. Fall back to a new session on first run.
		if _, err := sessMgr.ResumeLatest(); err != nil || sessMgr.Current() == nil {
			sess = sessMgr.NewSession()
		} else {
			sess = sessMgr.Current()
		}
	default:
		// Default agent: always create a new session.
		sess = sessMgr.NewSession()
	}
	if sess == nil {
		return RunAgentResponse{}, fmt.Errorf("failed to initialize session")
	}

	// If this is a named agent and the session is new, set a descriptive title.
	if agentName != "" && sess.Title == "New session" {
		sess.Title = "Agent: " + agentName
	}
	if setter, ok := handler.(interface{ SetSessionID(string) }); ok {
		setter.SetSessionID(sess.ID)
	}

	// --- Create and configure agent loop ---
	loop := agent.NewAgentLoop(llmClient, registry)
	loop.SetMaxIterations(effectiveCfg.Agent.MaxIterations)
	loop.SetMaxTokens(effectiveCfg.Agent.MaxTokens)
	loop.SetResultTruncation(effectiveCfg.Tools.ResultTruncation)
	loop.SetConfigDir(deps.StarclawDir)
	loop.SetContextWindow(effectiveCfg.Agent.ContextWindow)
	loop.SetTokenBudget(agent.TokenBudgetFromAgent(effectiveCfg.Agent))
	loop.SetPermissions(effectiveCfg.Permissions)
	loop.SetThinking(agent.ThinkingConfigFromAgent(effectiveCfg.Agent))
	loop.SetReasoningEffort(effectiveCfg.Agent.ReasoningEffort)
	specificModel := strings.TrimSpace(req.Model)
	if specificModel == "" {
		specificModel = effectiveCfg.Agent.Model
	}
	loop.SetSpecificModel(specificModel)
	loop.SetEnableStreaming(req.EnableStreaming)
	loop.SetSession(sess)
	loop.SetSessionManager(sessMgr)
	if approver != nil {
		loop.SetApprovalRequester(approver)
	}
	if req.PauseController != nil {
		loop.SetPauseController(req.PauseController)
	}
	if deps.MemoryPreflight != nil {
		loop.SetMemoryPreflightProvider(deps.MemoryPreflight)
	}
	if agentCfg != nil {
		agentDir := filepath.Join(deps.AgentsDir, agentName)
		loop.SwitchAgent(agentCfg.Prompt, agentDir)
		if agentCfg.Memory != "" {
			loop.SetMemory(agentCfg.Memory)
		}
	}
	if handler != nil {
		loop.SetEventHandler(handler)
	}

	// --- Run the agent loop ---
	resp, runErr := loop.Run(ctx, req.Text)

	// --- Save session after completion ---
	if saveErr := sessMgr.Save(); saveErr != nil {
		log.Printf("daemon: failed to save session: %v", saveErr)
	}

	// --- Build response ---
	response := RunAgentResponse{
		SessionID: sess.ID,
		Routing:   &routing,
	}
	budgetStatus := loop.LastBudgetStatus()
	if budgetStatus.Status != agent.TokenBudgetStatusDisabled {
		response.BudgetStatus = &budgetStatus
	}
	if resp != nil {
		if resp.Content != "" {
			response.Messages = []string{resp.Content}
		}
		if resp.Usage.InputTokens > 0 || resp.Usage.OutputTokens > 0 {
			response.Usage = map[string]int{
				"input_tokens":  resp.Usage.InputTokens,
				"output_tokens": resp.Usage.OutputTokens,
				"total_tokens":  resp.Usage.InputTokens + resp.Usage.OutputTokens,
			}
		}
	}
	if runErr != nil {
		response.Error = runErr.Error()
	}
	response.Fallback = agent.RecommendFallback(agent.FallbackInput{
		ProviderError: runErr,
		BudgetStatus:  budgetStatus,
		CurrentRoute:  routing.Route,
	})

	return response, nil
}

func effectiveRunConfig(deps *ServerDeps, agentCfg *agents.Agent) *config.Config {
	var base *config.Config
	if deps.Config != nil {
		base = deps.Config
	} else {
		base = defaultRunConfig()
	}
	if agentCfg != nil {
		return config.MergeAgentConfig(base, agentCfg)
	}
	return base
}

func defaultRunConfig() *config.Config {
	return &config.Config{
		Agent: config.AgentConfig{
			MaxIterations: 25,
			MaxTokens:     8192,
		},
		Tools: config.ToolsConfig{
			ResultTruncation: 30000,
		},
	}
}

func filteredRegistry(registry *agent.ToolRegistry, cfg config.ToolsConfig) *agent.ToolRegistry {
	filtered := registry.Clone()
	if len(cfg.Allowed) > 0 {
		filtered = filtered.FilterByAllow(cfg.Allowed)
	}
	if len(cfg.Denied) > 0 {
		filtered = filtered.FilterByDeny(cfg.Denied)
	}
	return filtered
}

// sessionsDirFor returns the directory for session storage under the given
// daemon server dependencies. Named agents get a subdirectory so their
// sessions are isolated from the default agent's sessions.
func sessionsDirFor(deps *ServerDeps, agentName string) string {
	base := filepath.Join(deps.StarclawDir, "sessions")
	if agentName != "" {
		return filepath.Join(base, agentName)
	}
	return base
}
