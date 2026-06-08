package daemon

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/starclaw/starclaw/internal/agents"
)

type suggestionResponse struct {
	Text            string `json:"text"`
	SuggestedAtUnix int64  `json:"suggested_at_unix"`
	Suggestion      string `json:"suggestion,omitempty"`
}

func (s *Server) validateSuggestionRoute(w http.ResponseWriter, r *http.Request) (string, bool) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name != "" {
		if err := agents.ValidateAgentName(name); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return "", false
		}
		if _, err := agents.LoadAgent(s.agentsDir(), name); err != nil {
			writeError(w, http.StatusNotFound, "agent not found")
			return "", false
		}
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "session id required")
		return "", false
	}
	if id != filepath.Base(id) || strings.ContainsAny(id, `/\`) {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return "", false
	}
	return id, true
}

func (s *Server) handleGetSuggestion(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := s.validateSuggestionRoute(w, r)
	if !ok {
		return
	}
	cur, present := s.suggestions.Get(sessionID)
	if !present || strings.TrimSpace(cur.Text) == "" {
		writeError(w, http.StatusNotFound, "no suggestion available")
		return
	}
	writeJSON(w, http.StatusOK, suggestionResponse{
		Text:            cur.Text,
		SuggestedAtUnix: cur.SuggestedAt.Unix(),
	})
}

func (s *Server) handleAcceptSuggestion(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := s.validateSuggestionRoute(w, r)
	if !ok {
		return
	}
	cur, present := s.suggestions.Get(sessionID)
	if !present || strings.TrimSpace(cur.Text) == "" {
		writeError(w, http.StatusNotFound, "no suggestion available")
		return
	}
	s.suggestions.MarkAccepted(sessionID)
	s.suggestions.Clear(sessionID)
	writeJSON(w, http.StatusOK, suggestionResponse{
		Text:            cur.Text,
		SuggestedAtUnix: cur.SuggestedAt.Unix(),
		Suggestion:      cur.Text,
	})
}
