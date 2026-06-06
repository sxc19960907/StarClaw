package daemon

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultCouncilStoreLimit = 50

type CouncilStore struct {
	mu      sync.Mutex
	limit   int
	records []CouncilRun
}

type CouncilRun struct {
	ID          string                `json:"id"`
	Goal        string                `json:"goal"`
	Status      string                `json:"status"`
	Agent       string                `json:"agent,omitempty"`
	Roles       []CouncilContribution `json:"roles"`
	Synthesis   string                `json:"synthesis"`
	CreatedAt   time.Time             `json:"created_at"`
	CompletedAt time.Time             `json:"completed_at"`
	Error       string                `json:"error,omitempty"`
}

type CouncilContribution struct {
	Role    string `json:"role"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Notes   string `json:"notes"`
}

type createCouncilRequest struct {
	Goal  string `json:"goal"`
	Agent string `json:"agent,omitempty"`
}

func NewCouncilStore(limit int) *CouncilStore {
	if limit <= 0 {
		limit = defaultCouncilStoreLimit
	}
	return &CouncilStore{limit: limit}
}

func (s *CouncilStore) Add(run CouncilRun) CouncilRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append([]CouncilRun{run}, s.records...)
	if len(s.records) > s.limit {
		s.records = s.records[:s.limit]
	}
	return run
}

func (s *CouncilStore) List() []CouncilRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]CouncilRun, len(s.records))
	copy(out, s.records)
	return out
}

func (s *CouncilStore) Get(id string) (CouncilRun, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, run := range s.records {
		if run.ID == id {
			return run, true
		}
	}
	return CouncilRun{}, false
}

func (s *Server) handleListCouncilRuns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"runs": s.councilStore.List()})
}

func (s *Server) handleGetCouncilRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := s.councilStore.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "council run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleCreateCouncilRun(w http.ResponseWriter, r *http.Request) {
	var req createCouncilRequest
	if !decodeBody(w, r, &req) {
		return
	}
	req.Goal = strings.TrimSpace(req.Goal)
	req.Agent = strings.TrimSpace(req.Agent)
	if req.Goal == "" {
		writeError(w, http.StatusBadRequest, "goal is required")
		return
	}
	if len([]rune(req.Goal)) > 4000 {
		writeError(w, http.StatusBadRequest, "goal is too long")
		return
	}

	run := buildCouncilRun(req)
	s.councilStore.Add(run)
	writeJSON(w, http.StatusCreated, run)
}

func buildCouncilRun(req createCouncilRequest) CouncilRun {
	now := time.Now()
	id := "council_" + generateRequestID()
	goal := req.Goal
	roles := []CouncilContribution{
		{
			Role:    "planner",
			Status:  "completed",
			Summary: "Define the smallest useful path.",
			Notes:   fmt.Sprintf("Scope the request around %q. Identify the user-visible outcome first, then split implementation into independently checkable steps.", goal),
		},
		{
			Role:    "researcher",
			Status:  "completed",
			Summary: "Gather local evidence before changing code.",
			Notes:   "Read the relevant task artifacts, current implementation, tests, and project specs. Prefer existing StarClaw patterns over new abstractions.",
		},
		{
			Role:    "reviewer",
			Status:  "completed",
			Summary: "Hold the approval and quality boundary.",
			Notes:   "Verify behavior with focused tests, then run the broader daemon/UI checks. Keep code changes bounded and preserve existing CLI/package names.",
		},
	}
	return CouncilRun{
		ID:          id,
		Goal:        goal,
		Status:      "completed",
		Agent:       req.Agent,
		Roles:       roles,
		Synthesis:   councilSynthesis(goal, roles),
		CreatedAt:   now,
		CompletedAt: now,
	}
}

func councilSynthesis(goal string, roles []CouncilContribution) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Council synthesis for: %s\n\n", goal)
	sb.WriteString("1. Confirm the narrow product outcome before broadening scope.\n")
	sb.WriteString("2. Use local evidence and current StarClaw patterns as the implementation source of truth.\n")
	sb.WriteString("3. Implement the smallest shippable slice, then validate with unit tests plus Web UI smoke where the surface changes.\n")
	sb.WriteString("4. Keep future automation behind explicit approval until the workflow has real state and audit trails.\n\n")
	sb.WriteString("Role notes:\n")
	for _, role := range roles {
		fmt.Fprintf(&sb, "- %s: %s\n", role.Role, role.Summary)
	}
	return sb.String()
}
