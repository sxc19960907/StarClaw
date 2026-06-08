package daemon

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const maxQueueTextBytes = 1 << 20

type queueCreateRequest struct {
	RouteKey   string            `json:"route_key,omitempty"`
	SessionID  string            `json:"session_id,omitempty"`
	Source     string            `json:"source,omitempty"`
	ExternalID string            `json:"external_id,omitempty"`
	Sender     string            `json:"sender,omitempty"`
	Agent      string            `json:"agent,omitempty"`
	Text       string            `json:"text"`
	Priority   *int              `json:"priority,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type queueClaimRequest struct {
	RouteKey string `json:"route_key"`
	Limit    int    `json:"limit,omitempty"`
}

type queueClaimTransitionRequest struct {
	ClaimID string `json:"claim_id"`
}

func (s *Server) handleCreateQueueMessage(w http.ResponseWriter, r *http.Request) {
	if s.mailboxStore == nil {
		writeError(w, http.StatusServiceUnavailable, "mailbox store unavailable")
		return
	}
	var req queueCreateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxQueueTextBytes+4096)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	priority := QueuePriorityNormal
	if req.Priority != nil {
		priority = *req.Priority
	}
	msg, status, err := queuedMessageFromCreateRequest(req, priority)
	if err != nil {
		if errors.Is(err, errQueueTextTooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	stored, err := s.mailboxStore.Enqueue(msg)
	if errors.Is(err, ErrMailboxFull) {
		writeError(w, http.StatusServiceUnavailable, "mailbox capacity reached for this route")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if stored.Duplicate {
		writeJSON(w, http.StatusOK, map[string]any{"message": stored, "duplicate": true})
		return
	}
	if s.replyRoutes != nil && stored.ExternalID != "" && stored.RouteKey != "" {
		s.replyRoutes.Put(stored.ExternalID, stored.RouteKey)
	}
	writeJSON(w, status, map[string]any{"message": stored, "duplicate": false})
}

func (s *Server) handleListQueueMessages(w http.ResponseWriter, r *http.Request) {
	if s.mailboxStore == nil {
		writeError(w, http.StatusServiceUnavailable, "mailbox store unavailable")
		return
	}
	routeKey := strings.TrimSpace(r.URL.Query().Get("route_key"))
	writeJSON(w, http.StatusOK, map[string]any{"messages": s.mailboxStore.List(routeKey)})
}

func (s *Server) handleGetQueueMessage(w http.ResponseWriter, r *http.Request) {
	if s.mailboxStore == nil {
		writeError(w, http.StatusServiceUnavailable, "mailbox store unavailable")
		return
	}
	msg, ok := s.mailboxStore.Get(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "queue message not found")
		return
	}
	writeJSON(w, http.StatusOK, msg)
}

func (s *Server) handleClaimQueueMessages(w http.ResponseWriter, r *http.Request) {
	if s.mailboxStore == nil {
		writeError(w, http.StatusServiceUnavailable, "mailbox store unavailable")
		return
	}
	var req queueClaimRequest
	if !decodeBody(w, r, &req) {
		return
	}
	messages, err := s.mailboxStore.Claim(req.RouteKey, req.Limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (s *Server) handleAckQueueMessage(w http.ResponseWriter, r *http.Request) {
	s.handleQueueClaimTransition(w, r, "ack")
}

func (s *Server) handleReleaseQueueMessage(w http.ResponseWriter, r *http.Request) {
	s.handleQueueClaimTransition(w, r, "release")
}

func (s *Server) handleQueueClaimTransition(w http.ResponseWriter, r *http.Request, action string) {
	if s.mailboxStore == nil {
		writeError(w, http.StatusServiceUnavailable, "mailbox store unavailable")
		return
	}
	var req queueClaimTransitionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	claimID := strings.TrimSpace(req.ClaimID)
	if id == "" || claimID == "" {
		writeError(w, http.StatusBadRequest, "id and claim_id are required")
		return
	}
	ok := false
	switch action {
	case "ack":
		ok = s.mailboxStore.Ack(id, claimID)
	case "release":
		ok = s.mailboxStore.Release(id, claimID)
	}
	if !ok {
		writeError(w, http.StatusConflict, "queue claim transition rejected")
		return
	}
	msg, _ := s.mailboxStore.Get(id)
	writeJSON(w, http.StatusOK, map[string]any{"message": msg})
}

var errQueueTextTooLarge = errors.New("message text exceeds 1 MB cap")

func queuedMessageFromCreateRequest(req queueCreateRequest, priority int) (QueuedMessage, int, error) {
	routeKey := strings.TrimSpace(req.RouteKey)
	sessionID := strings.TrimSpace(req.SessionID)
	text := strings.TrimSpace(req.Text)
	if routeKey == "" && sessionID == "" {
		return QueuedMessage{}, 0, errors.New("route_key or session_id is required")
	}
	if text == "" {
		return QueuedMessage{}, 0, errors.New("text is required")
	}
	if len(text) > maxQueueTextBytes {
		return QueuedMessage{}, 0, errQueueTextTooLarge
	}
	if priority < QueuePriorityHigh {
		return QueuedMessage{}, 0, errors.New("priority must be positive")
	}
	return QueuedMessage{
		RouteKey:   routeKey,
		SessionID:  sessionID,
		Source:     req.Source,
		ExternalID: req.ExternalID,
		Sender:     req.Sender,
		Agent:      req.Agent,
		Text:       text,
		Priority:   priority,
		Metadata:   req.Metadata,
	}, http.StatusAccepted, nil
}
