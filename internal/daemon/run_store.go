package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/client"
)

const defaultRunStoreLimit = 100

const (
	WorkflowStepPlanned         = "planned"
	WorkflowStepRunning         = "running"
	WorkflowStepBlocked         = "blocked"
	WorkflowStepWaitingApproval = "waiting_approval"
	WorkflowStepCompleted       = "completed"
	WorkflowStepFailed          = "failed"
	WorkflowStepCancelled       = "cancelled"
	WorkflowStepSkipped         = "skipped"
)

type RunEvent struct {
	Type string         `json:"type"`
	At   time.Time      `json:"at"`
	Data map[string]any `json:"data"`
}

type RunControlDecision struct {
	Action string    `json:"action"`
	Status string    `json:"status"`
	Reason string    `json:"reason,omitempty"`
	At     time.Time `json:"at"`
}

type WorkflowStepState struct {
	ID        string         `json:"id"`
	Title     string         `json:"title,omitempty"`
	Status    string         `json:"status"`
	Sequence  int            `json:"sequence,omitempty"`
	ParentID  string         `json:"parent_id,omitempty"`
	Attempt   int            `json:"attempt,omitempty"`
	StartedAt *time.Time     `json:"started_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`
	EndedAt   *time.Time     `json:"ended_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type RunRecord struct {
	ID               string                     `json:"id"`
	Status           string                     `json:"status"`
	Agent            string                     `json:"agent,omitempty"`
	Channel          string                     `json:"channel,omitempty"`
	Prompt           string                     `json:"prompt,omitempty"`
	SessionID        string                     `json:"session_id,omitempty"`
	StartedAt        time.Time                  `json:"started_at"`
	EndedAt          *time.Time                 `json:"ended_at,omitempty"`
	Request          RunAgentRequest            `json:"request"`
	Response         *RunAgentResponse          `json:"response,omitempty"`
	Usage            map[string]int             `json:"usage,omitempty"`
	Budget           *agent.TokenBudgetUsage    `json:"budget_status,omitempty"`
	Routing          *agent.RouteRecommendation `json:"routing,omitempty"`
	Fallback         *agent.FallbackDecision    `json:"fallback,omitempty"`
	Error            string                     `json:"error,omitempty"`
	Events           []RunEvent                 `json:"events,omitempty"`
	StructuredEvents []StructuredRunEvent       `json:"structured_events,omitempty"`
	Control          []RunControlDecision       `json:"control,omitempty"`
	Steps            []WorkflowStepState        `json:"steps,omitempty"`
}

type RunSummary struct {
	ID          string               `json:"id"`
	Status      string               `json:"status"`
	Agent       string               `json:"agent,omitempty"`
	Channel     string               `json:"channel,omitempty"`
	Prompt      string               `json:"prompt,omitempty"`
	SessionID   string               `json:"session_id,omitempty"`
	Source      string               `json:"source,omitempty"`
	Control     []RunControlDecision `json:"control,omitempty"`
	Steps       []WorkflowStepState  `json:"steps,omitempty"`
	TraceEvents int                  `json:"trace_events,omitempty"`
	StartedAt   time.Time            `json:"started_at"`
	EndedAt     *time.Time           `json:"ended_at,omitempty"`
}

type RunStore struct {
	mu          sync.RWMutex
	limit       int
	order       []string
	records     map[string]*RunRecord
	eventSeq    map[string]int
	persistPath string
	persistErr  error
}

type runStoreEnvelope struct {
	SchemaVersion string       `json:"schema_version"`
	Records       []*RunRecord `json:"records"`
}

func NewRunStore(limit int) *RunStore {
	if limit <= 0 {
		limit = defaultRunStoreLimit
	}
	return &RunStore{
		limit:    limit,
		records:  make(map[string]*RunRecord),
		eventSeq: make(map[string]int),
	}
}

func NewPersistentRunStore(limit int, path string) (*RunStore, error) {
	store := NewRunStore(limit)
	store.persistPath = path
	if path == "" {
		return store, nil
	}
	if err := store.load(); err != nil {
		store.persistErr = err
		return store, err
	}
	return store, nil
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
	s.addStructuredEventLocked(record.ID, "run_started", map[string]any{
		"agent":   req.Agent,
		"channel": req.Channel,
		"source":  req.Source,
	})
	for len(s.order) > s.limit {
		last := s.order[len(s.order)-1]
		s.order = s.order[:len(s.order)-1]
		delete(s.records, last)
		delete(s.eventSeq, last)
	}
	s.persistLocked()
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
	if record.Status == "cancelled" {
		s.persistLocked()
		return
	}
	if response.BudgetStatus != nil {
		s.addStructuredEventLocked(id, "budget_status", map[string]any{
			"status":        response.BudgetStatus.Status,
			"input_tokens":  response.BudgetStatus.InputTokens,
			"output_tokens": response.BudgetStatus.OutputTokens,
			"total_tokens":  response.BudgetStatus.TotalTokens,
			"unknown_turns": response.BudgetStatus.UnknownTurns,
			"detail":        response.BudgetStatus.Detail,
		})
	}
	if response.Routing != nil {
		s.addStructuredEventLocked(id, "routing_selected", map[string]any{
			"complexity": response.Routing.Complexity,
			"route":      response.Routing.Route,
			"model_tier": response.Routing.ModelTier,
			"reason":     response.Routing.Reason,
		})
	}
	if response.Fallback != nil {
		s.addStructuredEventLocked(id, "fallback_decision", map[string]any{
			"reason":     response.Fallback.Reason,
			"route":      response.Fallback.Route,
			"model_tier": response.Fallback.ModelTier,
			"detail":     response.Fallback.Detail,
		})
	}
	if err != nil {
		record.Status = "error"
		record.Error = err.Error()
		s.addStructuredEventLocked(id, "run_error", map[string]any{"error": err.Error()})
		s.persistLocked()
		return
	}
	if response.Error != "" {
		record.Status = "error"
		record.Error = response.Error
		s.addStructuredEventLocked(id, "run_error", map[string]any{"error": response.Error})
		s.persistLocked()
		return
	}
	record.Status = "completed"
	s.addStructuredEventLocked(id, "run_completed", map[string]any{
		"status": record.Status,
		"usage":  response.Usage,
	})
	s.persistLocked()
}

func (s *RunStore) AddEvent(id, eventType string, data map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[id]
	if record == nil {
		return
	}
	record.Events = append(record.Events, RunEvent{Type: eventType, At: time.Now(), Data: data})
	s.addStructuredEventLocked(id, eventType, data)
	s.persistLocked()
}

func (s *RunStore) AddControlDecision(id string, decision RunControlDecision) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[id]
	if record == nil {
		return false
	}
	if decision.At.IsZero() {
		decision.At = time.Now()
	}
	record.Control = append(record.Control, decision)
	if decision.Status == "cancelled" {
		record.Status = "cancelled"
		now := decision.At
		record.EndedAt = &now
	}
	s.addStructuredEventLocked(id, "control_decision", map[string]any{
		"action": decision.Action,
		"status": decision.Status,
		"reason": decision.Reason,
	})
	s.persistLocked()
	return true
}

func (s *RunStore) UpsertStep(runID string, step WorkflowStepState) bool {
	if step.ID == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[runID]
	if record == nil {
		return false
	}
	now := time.Now()
	normalizeWorkflowStep(&step, now)
	for idx := range record.Steps {
		if record.Steps[idx].ID != step.ID {
			continue
		}
		if step.StartedAt == nil {
			step.StartedAt = record.Steps[idx].StartedAt
		}
		if step.EndedAt == nil {
			step.EndedAt = record.Steps[idx].EndedAt
		}
		record.Steps[idx] = cloneWorkflowStep(step)
		s.addStructuredEventLocked(runID, "workflow_step", workflowStepEventData(step))
		s.persistLocked()
		return true
	}
	record.Steps = append(record.Steps, cloneWorkflowStep(step))
	sortWorkflowSteps(record.Steps)
	s.addStructuredEventLocked(runID, "workflow_step", workflowStepEventData(step))
	s.persistLocked()
	return true
}

func (s *RunStore) TransitionStep(runID, stepID, status string, metadata map[string]any) bool {
	if stepID == "" || status == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.records[runID]
	if record == nil {
		return false
	}
	now := time.Now()
	for idx := range record.Steps {
		if record.Steps[idx].ID != stepID {
			continue
		}
		step := record.Steps[idx]
		step.Status = status
		step.UpdatedAt = now
		if status == WorkflowStepRunning && step.StartedAt == nil {
			started := now
			step.StartedAt = &started
		}
		if workflowStepTerminal(status) {
			ended := now
			step.EndedAt = &ended
		}
		if step.Attempt <= 0 {
			step.Attempt = 1
		}
		step.Metadata = mergeWorkflowStepMetadata(step.Metadata, metadata)
		record.Steps[idx] = cloneWorkflowStep(step)
		s.addStructuredEventLocked(runID, "workflow_step", workflowStepEventData(step))
		s.persistLocked()
		return true
	}
	step := WorkflowStepState{ID: stepID, Status: status, Metadata: metadata}
	normalizeWorkflowStep(&step, now)
	if status == WorkflowStepRunning {
		started := now
		step.StartedAt = &started
	}
	if workflowStepTerminal(status) {
		ended := now
		step.EndedAt = &ended
	}
	record.Steps = append(record.Steps, cloneWorkflowStep(step))
	sortWorkflowSteps(record.Steps)
	s.addStructuredEventLocked(runID, "workflow_step", workflowStepEventData(step))
	s.persistLocked()
	return true
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
			ID:          record.ID,
			Status:      record.Status,
			Agent:       record.Agent,
			Channel:     record.Channel,
			Prompt:      record.Prompt,
			SessionID:   record.SessionID,
			Source:      record.Request.Source,
			Control:     append([]RunControlDecision(nil), record.Control...),
			Steps:       cloneWorkflowSteps(record.Steps),
			TraceEvents: len(record.StructuredEvents),
			StartedAt:   record.StartedAt,
			EndedAt:     record.EndedAt,
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
	if record.StructuredEvents != nil {
		copyRecord.StructuredEvents = append([]StructuredRunEvent(nil), record.StructuredEvents...)
	}
	if record.Control != nil {
		copyRecord.Control = append([]RunControlDecision(nil), record.Control...)
	}
	if record.Steps != nil {
		copyRecord.Steps = make([]WorkflowStepState, 0, len(record.Steps))
		for _, step := range record.Steps {
			copyRecord.Steps = append(copyRecord.Steps, cloneWorkflowStep(step))
		}
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

func (s *RunStore) Metrics() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	statusCounts := map[string]int{}
	eventCounts := map[string]int{}
	var totalInput, totalOutput int
	for _, id := range s.order {
		record := s.records[id]
		if record == nil {
			continue
		}
		statusCounts[record.Status]++
		if record.Usage != nil {
			totalInput += record.Usage["input_tokens"]
			totalOutput += record.Usage["output_tokens"]
		}
		for _, evt := range record.StructuredEvents {
			eventCounts[evt.Type]++
		}
	}
	return map[string]any{
		"runs_total":          len(s.records),
		"runs_by_status":      statusCounts,
		"events_by_type":      eventCounts,
		"tokens_input_total":  totalInput,
		"tokens_output_total": totalOutput,
		"schema_version":      structuredEventSchemaVersion,
		"stored_run_limit":    s.limit,
	}
}

func (s *RunStore) addStructuredEventLocked(id, eventType string, data map[string]any) {
	record := s.records[id]
	if record == nil {
		return
	}
	s.eventSeq[id]++
	record.StructuredEvents = append(record.StructuredEvents,
		newStructuredRunEvent(id, eventType, eventPhase(eventType), time.Now(), data, s.eventSeq[id]))
}

func (s *RunStore) PersistError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.persistErr
}

func (s *RunStore) load() error {
	data, err := os.ReadFile(s.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read run store: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	var envelope runStoreEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("parse run store: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.order = nil
	s.records = make(map[string]*RunRecord)
	s.eventSeq = make(map[string]int)
	for _, record := range envelope.Records {
		if record == nil || record.ID == "" {
			continue
		}
		if _, exists := s.records[record.ID]; exists {
			continue
		}
		s.records[record.ID] = record
		s.order = append(s.order, record.ID)
		s.eventSeq[record.ID] = maxStructuredEventSeq(record.StructuredEvents)
	}
	for len(s.order) > s.limit {
		last := s.order[len(s.order)-1]
		s.order = s.order[:len(s.order)-1]
		delete(s.records, last)
		delete(s.eventSeq, last)
	}
	return nil
}

func (s *RunStore) persistLocked() {
	if s.persistPath == "" {
		return
	}
	if err := s.saveLocked(); err != nil {
		s.persistErr = err
	}
}

func (s *RunStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.persistPath), 0o700); err != nil {
		return fmt.Errorf("create run store directory: %w", err)
	}
	envelope := runStoreEnvelope{
		SchemaVersion: structuredEventSchemaVersion,
		Records:       make([]*RunRecord, 0, len(s.order)),
	}
	for _, id := range s.order {
		record := s.records[id]
		if record == nil {
			continue
		}
		copyRecord := *record
		envelope.Records = append(envelope.Records, &copyRecord)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.persistPath), filepath.Base(s.persistPath)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create run store temp file: %w", err)
	}
	tmpPath := tmp.Name()
	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(envelope); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("encode run store: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close run store temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.persistPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace run store: %w", err)
	}
	s.persistErr = nil
	return nil
}

func maxStructuredEventSeq(events []StructuredRunEvent) int {
	maxSeq := 0
	for _, evt := range events {
		idx := strings.LastIndex(evt.ID, "-")
		if idx < 0 || idx == len(evt.ID)-1 {
			continue
		}
		seq, err := strconv.Atoi(evt.ID[idx+1:])
		if err == nil && seq > maxSeq {
			maxSeq = seq
		}
	}
	if maxSeq == 0 {
		return len(events)
	}
	return maxSeq
}

func normalizeWorkflowStep(step *WorkflowStepState, now time.Time) {
	if step.Status == "" {
		step.Status = WorkflowStepPlanned
	}
	if step.Attempt <= 0 {
		step.Attempt = 1
	}
	if step.UpdatedAt.IsZero() {
		step.UpdatedAt = now
	}
	step.Metadata = cloneMetadata(step.Metadata)
	if step.Status == WorkflowStepRunning && step.StartedAt == nil {
		started := step.UpdatedAt
		step.StartedAt = &started
	}
	if workflowStepTerminal(step.Status) && step.EndedAt == nil {
		ended := step.UpdatedAt
		step.EndedAt = &ended
	}
}

func workflowStepTerminal(status string) bool {
	switch status {
	case WorkflowStepCompleted, WorkflowStepFailed, WorkflowStepCancelled, WorkflowStepSkipped:
		return true
	default:
		return false
	}
}

func sortWorkflowSteps(steps []WorkflowStepState) {
	sort.SliceStable(steps, func(i, j int) bool {
		left, right := steps[i], steps[j]
		if left.Sequence == 0 || right.Sequence == 0 {
			return false
		}
		return left.Sequence < right.Sequence
	})
}

func cloneWorkflowStep(step WorkflowStepState) WorkflowStepState {
	copyStep := step
	copyStep.Metadata = cloneMetadata(step.Metadata)
	return copyStep
}

func cloneWorkflowSteps(steps []WorkflowStepState) []WorkflowStepState {
	if steps == nil {
		return nil
	}
	out := make([]WorkflowStepState, 0, len(steps))
	for _, step := range steps {
		out = append(out, cloneWorkflowStep(step))
	}
	return out
}

func cloneMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}
	copyMap := make(map[string]any, len(metadata))
	for key, value := range metadata {
		copyMap[key] = redactScalar(value)
	}
	return copyMap
}

func mergeWorkflowStepMetadata(existing, updates map[string]any) map[string]any {
	merged := cloneMetadata(existing)
	if len(updates) == 0 {
		return merged
	}
	if merged == nil {
		merged = make(map[string]any, len(updates))
	}
	for key, value := range updates {
		merged[key] = value
	}
	return merged
}

func workflowStepEventData(step WorkflowStepState) map[string]any {
	data := map[string]any{
		"step_id": step.ID,
		"status":  step.Status,
		"attempt": step.Attempt,
	}
	if step.Title != "" {
		data["title"] = step.Title
	}
	if step.Sequence != 0 {
		data["sequence"] = step.Sequence
	}
	if step.ParentID != "" {
		data["parent_id"] = step.ParentID
	}
	if len(step.Metadata) > 0 {
		data["metadata"] = cloneMetadata(step.Metadata)
	}
	return data
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
