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
	r.registerScheduleRoutes(mux)
	r.registerAgentRoutes(mux)
	r.registerSkillRoutes(mux)
	r.registerConfigRoutes(mux)
	r.registerInstructionsRoutes(mux)
	r.registerSessionRoutes(mux)
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
}

func (r *Router) registerMessageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /message", r.srv.handleMessage)
	mux.HandleFunc("POST /cancel", r.srv.handleCancel)
	mux.HandleFunc("POST /shutdown", r.srv.handleShutdown)
	mux.HandleFunc("GET /events", r.srv.handleEvents)
	mux.HandleFunc("GET /runs", r.srv.handleRuns)
	mux.HandleFunc("GET /runs/{id}", r.srv.handleGetRun)
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

func (r *Router) registerPermissionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /permissions", r.srv.handlePermissions)
	mux.HandleFunc("POST /approval", r.srv.handleApproval)
}
