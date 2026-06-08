package daemon

import (
	"net/http"
	"strings"
)

type cloudLifecycleActionRequest struct {
	Action string `json:"action"`
}

func (s *Server) handleGetCloudLifecycle(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.cloudLifecycleStatus())
}

func (s *Server) handlePostCloudLifecycle(w http.ResponseWriter, r *http.Request) {
	var req cloudLifecycleActionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if s.cloudLifecycle == nil {
		writeError(w, http.StatusServiceUnavailable, "cloud lifecycle controller unavailable")
		return
	}
	switch action {
	case "start":
		s.cloudLifecycle.Start(r.Context())
	case "stop":
		s.cloudLifecycle.Stop()
	case "restart":
		s.cloudLifecycle.Restart(r.Context())
	default:
		writeError(w, http.StatusBadRequest, "action must be start, stop, or restart")
		return
	}
	writeJSON(w, http.StatusOK, s.cloudLifecycle.Status())
}

func (s *Server) cloudLifecycleStatus() CloudLifecycleStatus {
	if s.cloudLifecycle == nil {
		return CloudLifecycleStatus{Note: defaultCloudLifecycleNote}
	}
	return s.cloudLifecycle.Status()
}
