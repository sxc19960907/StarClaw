package daemon

import (
	"net/http"
)

// Router groups HTTP route registration by module.
type Router struct {
	srv *Server
}

// NewRouter creates a Router bound to the given server.
func NewRouter(srv *Server) *Router {
	return &Router{srv: srv}
}

// RegisterRoutes registers all daemon HTTP routes on the provided mux.
// deps is accepted for future independent use even though the Router
// accesses dependencies through its Server reference today.
func (r *Router) RegisterRoutes(mux *http.ServeMux, deps *ServerDeps) {
	r.registerWebRoutes(mux)
	r.registerHealthRoutes(mux)
	r.registerMessageRoutes(mux)
	r.registerOpenAIRoutes(mux)
	r.registerScheduleRoutes(mux)
	r.registerAgentRoutes(mux)
	r.registerSkillRoutes(mux)
	r.registerConfigRoutes(mux)
	r.registerInstructionsRoutes(mux)
	r.registerSessionRoutes(mux)
	r.registerMemoryRoutes(mux)
	r.registerCouncilRoutes(mux)
	r.registerInboxRoutes(mux)
	r.registerQueueRoutes(mux)
	r.registerIntakeRoutes(mux)
	r.registerPermissionRoutes(mux)
}

// ---------------------------------------------------------------------------
// Per-module registration
// ---------------------------------------------------------------------------

func (r *Router) registerWebRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", r.srv.handleWebRoot)
	mux.HandleFunc("GET /app", r.srv.handleWebAppRedirect)
	mux.HandleFunc("GET /app/", r.srv.handleWebApp)
	mux.HandleFunc("GET /app/assets/", r.srv.handleWebAsset)
}

func (r *Router) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", r.srv.handleHealth)
	mux.HandleFunc("GET /status", r.srv.handleStatus)
	mux.HandleFunc("GET /diagnostics", r.srv.handleDiagnostics)
	mux.HandleFunc("GET /metrics", r.srv.handleMetrics)
	mux.HandleFunc("GET /version", r.srv.handleVersion)
	mux.HandleFunc("GET /update/check", r.srv.handleUpdateCheck)
}

func (r *Router) registerMessageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /message", r.srv.handleMessage)
	mux.HandleFunc("POST /cancel", r.srv.handleCancel)
	mux.HandleFunc("POST /shutdown", r.srv.handleShutdown)
	mux.HandleFunc("GET /events", r.srv.handleEvents)
	mux.HandleFunc("GET /runs", r.srv.handleRuns)
	mux.HandleFunc("GET /runs/{id}", r.srv.handleGetRun)
	mux.HandleFunc("GET /runs/{id}/trace", r.srv.handleGetRunTrace)
	mux.HandleFunc("POST /runs/{id}/control", r.srv.handleRunControl)
	mux.HandleFunc("GET /traces/export", r.srv.handleExportTraces)
}

func (r *Router) registerOpenAIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/chat/completions", r.srv.handleOpenAIChatCompletions)
}

func (r *Router) registerScheduleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /schedules", r.srv.handleListSchedules)
	mux.HandleFunc("GET /schedules/{id}", r.srv.handleGetSchedule)
	mux.HandleFunc("POST /schedules", r.srv.handleCreateSchedule)
	mux.HandleFunc("PATCH /schedules/{id}", r.srv.handlePatchSchedule)
	mux.HandleFunc("DELETE /schedules/{id}", r.srv.handleDeleteSchedule)
}

func (r *Router) registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /agents", r.srv.handleAgents)
	mux.HandleFunc("GET /agents/{name}", r.srv.handleGetAgent)
	mux.HandleFunc("POST /agents", r.srv.handleCreateAgent)
	mux.HandleFunc("PUT /agents/{name}", r.srv.handleUpdateAgent)
	mux.HandleFunc("DELETE /agents/{name}", r.srv.handleDeleteAgent)
}

func (r *Router) registerSkillRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /skills", r.srv.handleSkills)
}

func (r *Router) registerConfigRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /config", r.srv.handleGetConfig)
	mux.HandleFunc("PATCH /config", r.srv.handlePatchConfig)
	mux.HandleFunc("POST /mcp/test", r.srv.handleTestMCPServer)
}

func (r *Router) registerInstructionsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /instructions", r.srv.handleGetInstructions)
	mux.HandleFunc("PUT /instructions", r.srv.handlePutInstructions)
}

func (r *Router) registerSessionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /sessions", r.srv.handleSessions)
	mux.HandleFunc("GET /sessions/{id}", r.srv.handleGetSession)
	mux.HandleFunc("PATCH /sessions/{id}", r.srv.handlePatchSession)
	mux.HandleFunc("DELETE /sessions/{id}", r.srv.handleDeleteSession)
	mux.HandleFunc("GET /sessions/search", r.srv.handleSessionSearch)
}

func (r *Router) registerMemoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /memory", r.srv.handleGetMemory)
	mux.HandleFunc("POST /memory", r.srv.handleAppendMemory)
	mux.HandleFunc("DELETE /memory/{name}", r.srv.handleDeleteMemory)
}

func (r *Router) registerCouncilRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /council", r.srv.handleListCouncilRuns)
	mux.HandleFunc("POST /council", r.srv.handleCreateCouncilRun)
	mux.HandleFunc("GET /council/{id}", r.srv.handleGetCouncilRun)
	mux.HandleFunc("POST /council/{id}/run", r.srv.handleRunCouncilSynthesis)
}

func (r *Router) registerInboxRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /inbox", r.srv.handleListInbox)
	mux.HandleFunc("GET /inbox/providers", r.srv.handleInboxProviders)
	mux.HandleFunc("POST /inbox/webhook", r.srv.handleInboxWebhook)
	mux.HandleFunc("POST /inbox/github", r.srv.handleInboxGitHubWebhook)
	mux.HandleFunc("POST /inbox/{id}/approve", r.srv.handleApproveInboxItem)
	mux.HandleFunc("POST /inbox/{id}/reject", r.srv.handleRejectInboxItem)
	mux.HandleFunc("POST /inbox/{id}/retry", r.srv.handleRetryInboxItem)
}

func (r *Router) registerQueueRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /queue", r.srv.handleListQueueMessages)
	mux.HandleFunc("POST /queue", r.srv.handleCreateQueueMessage)
	mux.HandleFunc("GET /queue/{id}", r.srv.handleGetQueueMessage)
	mux.HandleFunc("POST /queue/claim", r.srv.handleClaimQueueMessages)
	mux.HandleFunc("POST /queue/{id}/ack", r.srv.handleAckQueueMessage)
	mux.HandleFunc("POST /queue/{id}/release", r.srv.handleReleaseQueueMessage)
}

func (r *Router) registerIntakeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /intake/file", r.srv.handleFileIntake)
}

func (r *Router) registerPermissionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /permissions", r.srv.handlePermissions)
	mux.HandleFunc("POST /approval", r.srv.handleApproval)
}
