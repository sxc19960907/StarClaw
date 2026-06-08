package daemon

import (
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

const defaultRunStoreLimit = 100

type RunEvent struct {
	Type string         `json:"type"`
	At   time.Time      `json:"at"`
	Data map[string]any `json:"data"`
}

type RunRecord struct {
	ID        string                     `json:"id"`
	Status    string                     `json:"status"`
	Agent     string                     `json:"agent,omitempty"`
	Channel   string                     `json:"channel,omitempty"`
	Prompt    string                     `json:"prompt,omitempty"`
	SessionID string                     `json:"session_id,omitempty"`
	StartedAt time.Time                  `json:"started_at"`
	EndedAt   *time.Time                 `json:"ended_at,omitempty"`
	Request   RunAgentRequest            `json:"request"`
	Response  *RunAgentResponse          `json:"response,omitempty"`
	Usage     map[string]int             `json:"usage,omitempty"`
	Budget    *agent.TokenBudgetUsage    `json:"budget_status,omitempty"`
	Routing   *agent.RouteRecommendation `json:"routing,omitempty"`
	Fallback  *agent.FallbackDecision    `json:"fallback,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Events    []RunEvent                 `json:"events,omitempty"`
}

type RunSummary struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Agent     string     `json:"agent,omitempty"`
	Channel   string     `json:"channel,omitempty"`
	Prompt    string     `json:"prompt,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

type RunStore struct {
	mu      sync.RWMutex
	limit   int
	order   []string
	records map[string]*RunRecord
}

func NewRunStore(limit int) *RunStore {
	if limit <= 0 {
		limit = defaultRunStoreLimit
	}
	return &RunStore{
		limit:   limit,
		records: make(map[string]*RunRecord),
	}
}

func (s *RunStore) Start(req RunAgentRequest) *RunRecord {
	record := &RunRecord{
		ID:        req.RequestID,
		Status:    "running",
		Agent:     req.Agent,
		Channel:   req.Channel,
		Prompt:    req.Text,
		StartedAt: time.Now(),
		Request:   req,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[record.ID]; !exists {
		s.order = append([]string{record.ID}, s.order...)
	}
	s.records[record.ID] = record
	for len(s.order) > s.limit {
		last := s.order[len(s.order)-1]
		s.order = s.order[:len(s.order)-1]
		delete(s.records, last)
	}
	return record
}

func (s *RunStore) Complete(id string, response RunAgentResponse, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[id]
	if record == nil {
		return
	}
	now := time.Now()
	record.EndedAt = &now
	record.Response = &response
	record.SessionID = response.SessionID
	record.Usage = response.Usage
	record.Budget = response.BudgetStatus
	record.Routing = response.Routing
	record.Fallback = response.Fallback
	if err != nil {
		record.Status = "error"
		record.Error = err.Error()
		return
	}
	if response.Error != "" {
		record.Status = "error"
		record.Error = response.Error
		return
	}
	record.Status = "completed"
}

func (s *RunStore) AddEvent(id, eventType string, data map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[id]
	if record == nil {
		return
	}
	record.Events = append(record.Events, RunEvent{Type: eventType, At: time.Now(), Data: data})
}

func (s *RunStore) List() []RunSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summaries := make([]RunSummary, 0, len(s.order))
	for _, id := range s.order {
		record := s.records[id]
		if record == nil {
			continue
		}
		summaries = append(summaries, RunSummary{
			ID:        record.ID,
			Status:    record.Status,
			Agent:     record.Agent,
			Channel:   record.Channel,
			Prompt:    record.Prompt,
			SessionID: record.SessionID,
			StartedAt: record.StartedAt,
			EndedAt:   record.EndedAt,
		})
	}
	return summaries
}

func (s *RunStore) Get(id string) (*RunRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record := s.records[id]
	if record == nil {
		return nil, false
	}
	copyRecord := *record
	if record.Events != nil {
		copyRecord.Events = append([]RunEvent(nil), record.Events...)
	}
	if record.Usage != nil {
		copyRecord.Usage = make(map[string]int, len(record.Usage))
		for key, value := range record.Usage {
			copyRecord.Usage[key] = value
		}
	}
	if record.Budget != nil {
		budget := *record.Budget
		copyRecord.Budget = &budget
	}
	if record.Routing != nil {
		routing := *record.Routing
		copyRecord.Routing = &routing
	}
	if record.Fallback != nil {
		fallback := *record.Fallback
		copyRecord.Fallback = &fallback
	}
	return &copyRecord, true
}

type runRecorderHandler struct {
	store *RunStore
	id    string
}

func (h *runRecorderHandler) OnToolCall(name string, args string) {
	h.add(EventToolCall, map[string]any{"tool": name, "args": args, "status": "running"})
}

func (h *runRecorderHandler) OnToolResult(name string, result agent.ToolResult) {
	status := "completed"
	if result.IsError {
		status = "error"
	}
	h.add(EventToolResult, map[string]any{
		"tool":           name,
		"status":         status,
		"content":        result.Content,
		"is_error":       result.IsError,
		"error_category": string(result.ErrorCategory),
	})
}

func (h *runRecorderHandler) OnText(text string) {
	h.add(EventText, map[string]any{"text": text})
}

func (h *runRecorderHandler) OnUsage(usage client.Usage) {
	h.add(EventUsage, map[string]any{
		"input_tokens":  usage.InputTokens,
		"output_tokens": usage.OutputTokens,
	})
}

func (h *runRecorderHandler) OnRunStatus(code string, detail string) {
	h.add(EventRunStatus, map[string]any{"code": code, "detail": detail})
}

func (h *runRecorderHandler) OnBudgetStatus(status agent.TokenBudgetUsage) {
	data := map[string]any{
		"status":        status.Status,
		"input_tokens":  status.InputTokens,
		"output_tokens": status.OutputTokens,
		"total_tokens":  status.TotalTokens,
	}
	if status.UnknownTurns > 0 {
		data["unknown_turns"] = status.UnknownTurns
	}
	if status.Detail != "" {
		data["detail"] = status.Detail
	}
	h.add(EventBudgetStatus, data)
}

func (h *runRecorderHandler) OnStreamDelta(delta string) {
	h.add(EventStreamDelta, map[string]any{"delta": delta})
}

func (h *runRecorderHandler) OnPreamble(preamble string) {
	if preamble == "" {
		return
	}
	h.add(EventPreamble, map[string]any{"preamble": preamble})
}

func (h *runRecorderHandler) add(eventType string, data map[string]any) {
	if h.store == nil {
		return
	}
	h.store.AddEvent(h.id, eventType, data)
}
