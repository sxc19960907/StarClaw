package daemon

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

const ChannelInbox = "inbox"

type inboxWebhookRequest struct {
	Provider   string            `json:"provider,omitempty"`
	ExternalID string            `json:"external_id"`
	Sender     string            `json:"sender,omitempty"`
	Text       string            `json:"text"`
	Agent      string            `json:"agent,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type inboxActionRequest struct {
	Agent string `json:"agent,omitempty"`
}

func (s *Server) handleListInbox(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.inboxStore.List()})
}

func (s *Server) handleInboxWebhook(w http.ResponseWriter, r *http.Request) {
	var req inboxWebhookRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, status, err := inboxItemFromWebhook(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	stored, duplicate := s.inboxStore.Upsert(item)
	if duplicate {
		writeJSON(w, http.StatusOK, map[string]any{"item": stored, "duplicate": true})
		return
	}
	writeJSON(w, status, map[string]any{"item": stored, "duplicate": false})
}

func (s *Server) handleApproveInboxItem(w http.ResponseWriter, r *http.Request) {
	s.handleInboxRunAction(w, r, "approved")
}

func (s *Server) handleRetryInboxItem(w http.ResponseWriter, r *http.Request) {
	s.handleInboxRunAction(w, r, "retrying")
}

func (s *Server) handleRejectInboxItem(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	item, err := s.inboxStore.Update(id, func(item *InboxItem) error {
		if item.Status == "running" {
			return errors.New("cannot reject a running inbox item")
		}
		item.Status = "rejected"
		item.Error = ""
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleInboxRunAction(w http.ResponseWriter, r *http.Request, transition string) {
	if s.deps == nil {
		writeError(w, http.StatusInternalServerError, "daemon deps not configured")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	var req inboxActionRequest
	if r.Body != nil && r.ContentLength != 0 {
		if !decodeBody(w, r, &req) {
			return
		}
	}
	item, ok := s.inboxStore.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "inbox item not found")
		return
	}
	if transition == "approved" && item.Status != "pending" {
		writeError(w, http.StatusBadRequest, "only pending inbox items can be approved")
		return
	}
	if transition == "retrying" && item.Status != "failed" {
		writeError(w, http.StatusBadRequest, "only failed inbox items can be retried")
		return
	}
	agentName := strings.TrimSpace(req.Agent)
	if agentName == "" {
		agentName = item.Agent
	}
	runID := "inbox_run_" + generateRequestID()
	started, err := s.inboxStore.Update(id, func(item *InboxItem) error {
		item.Status = "running"
		item.Agent = agentName
		item.RunID = runID
		item.Error = ""
		return nil
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	runReq := RunAgentRequest{
		Text:       started.Text,
		Agent:      agentName,
		Source:     started.Provider,
		Channel:    ChannelInbox,
		Sender:     started.Sender,
		NewSession: true,
		RequestID:  runID,
	}
	s.runStore.Start(runReq)
	handler := s.recordingHandler(runID, &httpEventHandler{})
	result, runErr := s.runAgent(context.Background(), runReq, handler)
	s.runStore.Complete(runID, result, runErr)

	finalStatus := "completed"
	errText := ""
	if runErr != nil {
		finalStatus = "failed"
		errText = runErr.Error()
	} else if result.Error != "" {
		finalStatus = "failed"
		errText = result.Error
	}
	updated, updateErr := s.inboxStore.Update(id, func(item *InboxItem) error {
		item.Status = finalStatus
		item.RunID = runID
		item.Error = errText
		return nil
	})
	if updateErr != nil {
		writeError(w, http.StatusInternalServerError, updateErr.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": updated, "run": result})
}

func inboxItemFromWebhook(req inboxWebhookRequest) (InboxItem, int, error) {
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "webhook"
	}
	externalID := strings.TrimSpace(req.ExternalID)
	text := strings.TrimSpace(req.Text)
	if externalID == "" {
		return InboxItem{}, 0, errors.New("external_id is required")
	}
	if text == "" {
		return InboxItem{}, 0, errors.New("text is required")
	}
	if len([]rune(text)) > 8000 {
		return InboxItem{}, 0, errors.New("text is too long")
	}
	return InboxItem{
		Provider:   provider,
		ExternalID: externalID,
		Sender:     strings.TrimSpace(req.Sender),
		Text:       text,
		Status:     "pending",
		Agent:      strings.TrimSpace(req.Agent),
		Metadata:   cleanInboxMetadata(req.Metadata),
	}, http.StatusCreated, nil
}

func cleanInboxMetadata(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
