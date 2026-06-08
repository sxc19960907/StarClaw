package daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/agents"
	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/schedule"
	"github.com/starclaw/starclaw/internal/session"
	"github.com/starclaw/starclaw/internal/skills"
	"gopkg.in/yaml.v3"
)

// Server is the daemon HTTP server.
type Server struct {
	port           int
	deps           *ServerDeps
	srv            *http.Server
	version        string
	eventBus       *EventBus
	approvalBroker *ApprovalBroker
	ctx            context.Context
	cancel         context.CancelFunc
	startedAt      time.Time
	running        sync.Map // requestID -> context.CancelFunc
	runStore       *RunStore
	councilStore   *CouncilStore
	inboxStore     *InboxStore
}

// NewServer creates a new Server.
func NewServer(port int, deps *ServerDeps, version string) *Server {
	return &Server{
		port:           port,
		deps:           deps,
		version:        version,
		eventBus:       NewEventBus(),
		approvalBroker: NewApprovalBroker(),
		runStore:       NewRunStore(defaultRunStoreLimit),
		councilStore:   NewCouncilStore(defaultCouncilStoreLimit),
		inboxStore:     NewInboxStore(defaultInboxStoreLimit),
	}
}

// Port returns the server's port number.
func (s *Server) Port() int {
	return s.port
}

// Handler returns the HTTP handler for this server. Exported for testing.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	NewRouter(s).RegisterRoutes(mux, s.deps)
	return mux
}

// Start starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	s.ctx = ctx
	s.startedAt = time.Now()

	s.srv = &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", s.port),
		Handler: s.Handler(),
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.srv.Shutdown(shutCtx); err != nil {
			log.Printf("daemon: server shutdown error: %v", err)
		}
	}()

	log.Printf("daemon: starting server on localhost:%d", s.port)
	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("daemon server: %w", err)
	}
	return nil
}

// SetCancelFunc sets a cancel function that handleShutdown will call to
// stop the daemon.
func (s *Server) SetCancelFunc(cancel context.CancelFunc) {
	s.cancel = cancel
}

// ---------------------------------------------------------------------------
// Health / Status
// ---------------------------------------------------------------------------

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": s.version,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startedAt).Seconds()
	activeCount := 0
	s.running.Range(func(_, _ interface{}) bool {
		activeCount++
		return true
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"uptime":        int(uptime),
		"version":       s.version,
		"active_agents": activeCount,
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"metrics": s.runStore.Metrics(),
	})
}

// ---------------------------------------------------------------------------
// Message / Agent execution
// ---------------------------------------------------------------------------

func (s *Server) handleMessage(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil {
		writeError(w, http.StatusInternalServerError, "daemon deps not configured")
		return
	}

	var req RunAgentRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Source == "" {
		req.Source = "starclaw"
	}
	if req.Channel == "" {
		req.Channel = ChannelHTTP
	}
	if req.RequestID == "" {
		req.RequestID = generateRequestID()
	}
	s.runStore.Start(req)

	// Create cancellable context for this request.
	ctx, cancel := context.WithCancel(r.Context())
	s.running.Store(req.RequestID, cancel)
	defer s.running.Delete(req.RequestID)

	// SSE streaming.
	if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		s.handleMessageSSE(w, r, req, ctx)
		return
	}

	// Synchronous JSON response.
	handler := s.recordingHandler(req.RequestID, &httpEventHandler{})
	result, err := s.runAgent(ctx, req, handler)
	s.runStore.Complete(req.RequestID, result, err)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleMessageSSE(w http.ResponseWriter, r *http.Request, req RunAgentRequest, ctx context.Context) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	handler := s.recordingHandler(req.RequestID, &sseEventHandler{w: w, flusher: flusher})
	result, err := s.runAgent(ctx, req, handler)
	s.runStore.Complete(req.RequestID, result, err)
	if err != nil {
		_, _ = fmt.Fprintf(w, "event: error\ndata: %s\n\n", mustJSON(map[string]string{"error": err.Error()}))
		flusher.Flush()
		return
	}
	_, _ = fmt.Fprintf(w, "event: done\ndata: %s\n\n", mustJSON(result))
	flusher.Flush()
}

func (s *Server) runAgent(ctx context.Context, req RunAgentRequest, handler agent.EventHandler) (RunAgentResponse, error) {
	approver := NewDaemonApprovalRequester(s.approvalBroker, s.eventBus, req.Channel, req.RequestID, req.Agent)
	return RunAgentWithApproval(ctx, s.deps, req, handler, approver)
}

func (s *Server) recordingHandler(requestID string, handler agent.EventHandler) agent.EventHandler {
	recorder := &runRecorderHandler{store: s.runStore, id: requestID}
	if handler == nil {
		return recorder
	}
	return NewMultiHandler(handler, recorder)
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"runs": s.runStore.List()})
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	record, ok := s.runStore.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

// ---------------------------------------------------------------------------
// Cancel / Shutdown
// ---------------------------------------------------------------------------

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RequestID string `json:"request_id"`
		Reason    string `json:"reason,omitempty"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.RequestID == "" {
		writeError(w, http.StatusBadRequest, "request_id is required")
		return
	}

	if cancel, ok := s.running.Load(body.RequestID); ok {
		cancel.(context.CancelFunc)()
		s.runStore.AddControlDecision(body.RequestID, RunControlDecision{
			Action: "cancel",
			Status: "cancelled",
			Reason: body.Reason,
		})
		writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled", "run_id": body.RequestID, "action": "cancel"})
		return
	}
	writeError(w, http.StatusNotFound, "request not found")
}

func (s *Server) handleRunControl(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		writeError(w, http.StatusBadRequest, "run id is required")
		return
	}

	var body struct {
		Action   string `json:"action"`
		Reason   string `json:"reason,omitempty"`
		Approved bool   `json:"approved,omitempty"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	action := strings.ToLower(strings.TrimSpace(body.Action))
	if action == "" {
		writeError(w, http.StatusBadRequest, "action is required")
		return
	}

	switch action {
	case "cancel":
		if cancel, ok := s.running.Load(runID); ok {
			cancel.(context.CancelFunc)()
			s.runStore.AddControlDecision(runID, RunControlDecision{Action: action, Status: "cancelled", Reason: body.Reason})
			writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled", "run_id": runID, "action": action})
			return
		}
		if _, ok := s.runStore.Get(runID); ok {
			s.runStore.AddControlDecision(runID, RunControlDecision{Action: action, Status: "not_running", Reason: body.Reason})
			writeError(w, http.StatusConflict, "run is not active")
			return
		}
		writeError(w, http.StatusNotFound, "run not found")
	case "pause", "resume":
		if _, ok := s.runStore.Get(runID); !ok {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		s.runStore.AddControlDecision(runID, RunControlDecision{Action: action, Status: "unsupported", Reason: body.Reason})
		writeError(w, http.StatusConflict, action+" is not supported for this runtime")
	case "replay":
		record, ok := s.runStore.Get(runID)
		if !ok {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		s.runStore.AddControlDecision(runID, RunControlDecision{Action: action, Status: "approval_required", Reason: body.Reason})
		writeJSON(w, http.StatusOK, map[string]any{
			"status": "approval_required",
			"run_id": runID,
			"action": action,
			"replay": map[string]any{
				"source_run_id":     runID,
				"requires_approval": true,
				"approved":          body.Approved,
				"reason":            "Replay can repeat tool calls or external effects.",
				"request":           replayControlRequest(record.Request),
			},
		})
	default:
		writeError(w, http.StatusBadRequest, "unsupported action")
	}
}

func replayControlRequest(req RunAgentRequest) map[string]any {
	out := map[string]any{
		"text_redacted": true,
		"channel":       req.Channel,
	}
	if req.Agent != "" {
		out["agent"] = req.Agent
	}
	if req.SessionID != "" {
		out["session_id"] = req.SessionID
	}
	if req.Source != "" {
		out["source"] = req.Source
	}
	return out
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "shutting_down"})
	if s.cancel != nil {
		log.Println("daemon: shutdown requested via /shutdown")
		go s.cancel()
	}
}

// ---------------------------------------------------------------------------
// Events (SSE)
// ---------------------------------------------------------------------------

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	id := generateRequestID()
	ch := s.eventBus.Subscribe(id)
	defer s.eventBus.Unsubscribe(id)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case evt := <-ch:
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, evt.Data)
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Schedule CRUD
// ---------------------------------------------------------------------------

func (s *Server) scheduleManager() *schedule.Manager {
	if s.deps == nil {
		return nil
	}
	return s.deps.ScheduleManager
}

func (s *Server) handleListSchedules(w http.ResponseWriter, r *http.Request) {
	mgr := s.scheduleManager()
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "schedule manager not configured")
		return
	}
	list, err := mgr.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []schedule.Schedule{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"schedules": list})
}

func (s *Server) handleGetSchedule(w http.ResponseWriter, r *http.Request) {
	mgr := s.scheduleManager()
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "schedule manager not configured")
		return
	}
	id := r.PathValue("id")
	sched, err := mgr.Get(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sched)
}

func (s *Server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	mgr := s.scheduleManager()
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "schedule manager not configured")
		return
	}

	var body struct {
		Agent  string `json:"agent"`
		Cron   string `json:"cron"`
		Prompt string `json:"prompt"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	id, err := mgr.Create(body.Agent, body.Cron, body.Prompt)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "cannot be empty") {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	sched, err := mgr.Get(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sched)
}

func (s *Server) handlePatchSchedule(w http.ResponseWriter, r *http.Request) {
	mgr := s.scheduleManager()
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "schedule manager not configured")
		return
	}

	id := r.PathValue("id")
	var patch struct {
		Cron    *string `json:"cron"`
		Prompt  *string `json:"prompt"`
		Enabled *bool   `json:"enabled"`
	}
	if !decodeBody(w, r, &patch) {
		return
	}
	if patch.Cron == nil && patch.Prompt == nil && patch.Enabled == nil {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	update := &schedule.UpdateOpts{
		Cron:    patch.Cron,
		Prompt:  patch.Prompt,
		Enabled: patch.Enabled,
	}
	if err := mgr.Update(id, update); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "no fields to update") ||
			strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "cannot be empty") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sched, err := mgr.Get(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sched)
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	mgr := s.scheduleManager()
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "schedule manager not configured")
		return
	}
	id := r.PathValue("id")
	if err := mgr.Remove(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---------------------------------------------------------------------------
// Agent CRUD
// ---------------------------------------------------------------------------

func (s *Server) agentsDir() string {
	if s.deps == nil {
		return ""
	}
	return s.deps.AgentsDir
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	dir := s.agentsDir()
	if dir == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"agents": []interface{}{}})
		return
	}
	infos, err := agents.ListAgents(dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if infos == nil {
		infos = []agents.AgentInfo{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"agents": infos})
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := agents.ValidateAgentName(name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	dir := s.agentsDir()
	if dir == "" {
		writeError(w, http.StatusInternalServerError, "agents directory not configured")
		return
	}
	a, err := agents.LoadAgent(dir, name)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	dir := s.agentsDir()
	if dir == "" {
		writeError(w, http.StatusInternalServerError, "agents directory not configured")
		return
	}
	var req agentEditRequest
	if !decodeBody(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "agent name is required")
		return
	}
	agent, err := saveAgentDefinition(dir, name, req, true)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, agent)
}

func (s *Server) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := agents.ValidateAgentName(name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	dir := s.agentsDir()
	if dir == "" {
		writeError(w, http.StatusInternalServerError, "agents directory not configured")
		return
	}
	var req agentEditRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Name) != "" && strings.TrimSpace(req.Name) != name {
		writeError(w, http.StatusBadRequest, "agent name cannot be changed")
		return
	}
	agent, err := saveAgentDefinition(dir, name, req, false)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := agents.ValidateAgentName(name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	dir := s.agentsDir()
	if dir == "" {
		writeError(w, http.StatusInternalServerError, "agents directory not configured")
		return
	}
	// Verify the agent exists.
	agentDir := filepath.Join(dir, name)
	if _, err := os.Stat(filepath.Join(agentDir, "AGENT.md")); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, fmt.Sprintf("agent %q not found", name))
		return
	}
	if err := os.RemoveAll(agentDir); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---------------------------------------------------------------------------
// Skills
// ---------------------------------------------------------------------------

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil || s.deps.SkillsDir == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"skills": []interface{}{}})
		return
	}
	metas, err := skills.ListSkills(skills.SkillSource{
		Dir:    s.deps.SkillsDir,
		Source: skills.SourceGlobal,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if metas == nil {
		metas = []skills.SkillMeta{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"skills": metas})
}

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil || s.deps.ConfigPath == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"config": nil})
		return
	}
	cfg, err := readDaemonConfig(s.deps.ConfigPath, s.deps.Config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"config": newDaemonConfigView(cfg)})
}

func (s *Server) handlePatchConfig(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil || s.deps.ConfigPath == "" {
		writeError(w, http.StatusInternalServerError, "config path not configured")
		return
	}

	var patch providerConfigPatch
	if !decodeBody(w, r, &patch) {
		return
	}

	cfg, err := readDaemonConfig(s.deps.ConfigPath, s.deps.Config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := patch.apply(cfg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("marshal config yaml: %v", err))
		return
	}
	if err := os.WriteFile(s.deps.ConfigPath, out, 0600); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.deps.Config = cfg
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "updated",
		"config": newDaemonConfigView(cfg),
	})
}

// ---------------------------------------------------------------------------
// Instructions
// ---------------------------------------------------------------------------

func (s *Server) instructionsPath() string {
	if s.deps == nil || s.deps.InstructionsDir == "" {
		return ""
	}
	return filepath.Join(s.deps.InstructionsDir, "instructions.md")
}

func (s *Server) handleGetInstructions(w http.ResponseWriter, r *http.Request) {
	path := s.instructionsPath()
	if path == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"content": nil})
		return
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"content": nil})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"content": string(data)})
}

func (s *Server) handlePutInstructions(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content *string `json:"content"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	path := s.instructionsPath()
	if path == "" {
		writeError(w, http.StatusInternalServerError, "instructions directory not configured")
		return
	}

	if body.Content == nil {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := os.WriteFile(path, []byte(*body.Content), 0600); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

func (s *Server) sessionManagerFor(agentName string) *session.Manager {
	if s.deps == nil || s.deps.StarclawDir == "" {
		return nil
	}
	dir := filepath.Join(s.deps.StarclawDir, "sessions")
	if agentName != "" {
		dir = filepath.Join(dir, agentName)
	}
	return session.NewManager(dir)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	agentName := r.URL.Query().Get("agent")
	mgr := s.sessionManagerFor(agentName)
	if mgr == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": []interface{}{}})
		return
	}
	summaries, err := mgr.List()
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": []interface{}{}})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if summaries == nil {
		summaries = []session.SessionSummary{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"sessions": summaries})
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "session id required")
		return
	}
	if id != filepath.Base(id) || strings.ContainsAny(id, `/\`) {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	agentName := r.URL.Query().Get("agent")
	mgr := s.sessionManagerFor(agentName)
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "daemon deps not configured")
		return
	}
	sess, err := mgr.Resume(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, fmt.Sprintf("session %q not found", id))
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handlePatchSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "session id required")
		return
	}
	if id != filepath.Base(id) || strings.ContainsAny(id, `/\`) {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	var patch struct {
		Title    *string `json:"title"`
		Favorite *bool   `json:"favorite"`
	}
	if !decodeBody(w, r, &patch) {
		return
	}
	if patch.Title == nil && patch.Favorite == nil {
		writeError(w, http.StatusBadRequest, "no session fields supplied")
		return
	}

	agentName := r.URL.Query().Get("agent")
	mgr := s.sessionManagerFor(agentName)
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "daemon deps not configured")
		return
	}
	sess, err := mgr.Resume(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, fmt.Sprintf("session %q not found", id))
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if patch.Title != nil {
		title := strings.TrimSpace(*patch.Title)
		if title == "" {
			writeError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		sess.Title = title
	}
	if patch.Favorite != nil {
		sess.Favorite = *patch.Favorite
	}
	if err := mgr.Save(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "session id required")
		return
	}
	if id != filepath.Base(id) || strings.ContainsAny(id, `/\`) {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	agentName := r.URL.Query().Get("agent")
	mgr := s.sessionManagerFor(agentName)
	if mgr == nil {
		writeError(w, http.StatusInternalServerError, "daemon deps not configured")
		return
	}
	if err := mgr.Delete(id); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, fmt.Sprintf("session %q not found", id))
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleSessionSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "q parameter required")
		return
	}

	agentName := r.URL.Query().Get("agent")
	mgr := s.sessionManagerFor(agentName)
	if mgr == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"results": []interface{}{}})
		return
	}

	// List all sessions and filter by title match.
	summaries, err := mgr.List()
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, map[string]interface{}{"results": []interface{}{}})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	q := strings.ToLower(query)
	var results []session.SessionSummary
	for _, s := range summaries {
		if strings.Contains(strings.ToLower(s.Title), q) {
			results = append(results, s)
		}
	}
	if results == nil {
		results = []session.SessionSummary{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

// ---------------------------------------------------------------------------
// Permissions / Approval
// ---------------------------------------------------------------------------

func (s *Server) handlePermissions(w http.ResponseWriter, r *http.Request) {
	overview := map[string]interface{}{
		"configured":         false,
		"allowed_dirs":       []string{},
		"allowed_commands":   []string{},
		"denied_commands":    []string{},
		"network_allowlist":  []string{},
		"sensitive_patterns": []string{},
	}
	if s.deps != nil && s.deps.Config != nil && s.deps.Config.Permissions != nil {
		perms := s.deps.Config.Permissions
		overview["configured"] = true
		overview["allowed_dirs"] = stringSliceOrEmpty(perms.AllowedDirs)
		overview["allowed_commands"] = stringSliceOrEmpty(perms.AllowedCommands)
		overview["denied_commands"] = stringSliceOrEmpty(perms.DeniedCommands)
		overview["network_allowlist"] = stringSliceOrEmpty(perms.NetworkAllowlist)
		overview["sensitive_patterns"] = stringSliceOrEmpty(perms.SensitivePatterns)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"permissions": overview,
	})
}

func (s *Server) handleApproval(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RequestID string           `json:"request_id"`
		Decision  ApprovalDecision `json:"decision"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.RequestID == "" {
		writeError(w, http.StatusBadRequest, "request_id required")
		return
	}
	switch body.Decision {
	case DecisionAllow, DecisionDeny:
	default:
		writeError(w, http.StatusBadRequest, "decision must be allow or deny")
		return
	}

	s.approvalBroker.Resolve(ApprovalResolvedPayload{
		RequestID:  body.RequestID,
		Decision:   body.Decision,
		ResolvedBy: "api",
	})

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---------------------------------------------------------------------------
// Event handler types
// ---------------------------------------------------------------------------

// httpEventHandler is a no-op EventHandler for synchronous HTTP requests.
type httpEventHandler struct{}

func (h *httpEventHandler) OnToolCall(name string, args string) {}
func (h *httpEventHandler) OnToolResult(name string, result agent.ToolResult) {
	log.Printf("http: tool %s completed", name)
}
func (h *httpEventHandler) OnText(text string)         {}
func (h *httpEventHandler) OnUsage(usage client.Usage) {}
func (h *httpEventHandler) OnStreamDelta(delta string) {}
func (h *httpEventHandler) OnPreamble(preamble string) {}

// sseEventHandler streams agent events as SSE to an HTTP response.
type sseEventHandler struct {
	w            http.ResponseWriter
	flusher      http.Flusher
	streamedText bool
}

func (h *sseEventHandler) OnToolCall(name string, args string) {
	data := mustJSON(map[string]string{"tool": name, "status": "running", "args": args})
	_, _ = fmt.Fprintf(h.w, "event: tool_call\ndata: %s\n\n", data)
	h.flusher.Flush()
}

func (h *sseEventHandler) OnToolResult(name string, result agent.ToolResult) {
	status := "completed"
	if result.IsError {
		status = "error"
	}
	data := mustJSON(map[string]interface{}{
		"tool":           name,
		"status":         status,
		"content":        result.Content,
		"is_error":       result.IsError,
		"error_category": string(result.ErrorCategory),
	})
	_, _ = fmt.Fprintf(h.w, "event: tool_result\ndata: %s\n\n", data)
	h.flusher.Flush()
}

func (h *sseEventHandler) OnText(text string) {
	if h.streamedText {
		return
	}
	data := mustJSON(map[string]string{"text": text})
	_, _ = fmt.Fprintf(h.w, "event: text\ndata: %s\n\n", data)
	h.flusher.Flush()
}

func (h *sseEventHandler) OnUsage(usage client.Usage) {
	// Usage is reported via the final "done" event; no per-event emission needed.
}

func (h *sseEventHandler) OnStreamDelta(delta string) {
	if delta == "" {
		return
	}
	h.streamedText = true
	data := mustJSON(map[string]string{"text": delta})
	_, _ = fmt.Fprintf(h.w, "event: text\ndata: %s\n\n", data)
	h.flusher.Flush()
}

func (h *sseEventHandler) OnPreamble(preamble string) {
	data := mustJSON(map[string]string{"preamble": preamble})
	_, _ = fmt.Fprintf(h.w, "event: preamble\ndata: %s\n\n", data)
	h.flusher.Flush()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const maxBodySize = 1 << 20 // 1 MB

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// decodeBody reads a JSON request body with a size limit. Returns false and
// writes an error response if decoding fails.
func decodeBody(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func stringSliceOrEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// mustJSON marshals v to JSON, returning "{}" on error.
func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// generateRequestID creates a unique request identifier.
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "req_" + hex.EncodeToString([]byte(time.Now().Format("150405.000000")))
	}
	return "req_" + hex.EncodeToString(b)
}
