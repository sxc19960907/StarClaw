package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/tools"
)

type fileIntakeRequest struct {
	Path       string `json:"path"`
	Mode       string `json:"mode"`
	MaxChars   int    `json:"max_chars,omitempty"`
	MaxEntries int    `json:"max_entries,omitempty"`
}

type fileIntakeResponse struct {
	Mode    string `json:"mode"`
	Path    string `json:"path"`
	Status  string `json:"status"`
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

func (s *Server) handleFileIntake(w http.ResponseWriter, r *http.Request) {
	var req fileIntakeRequest
	if !decodeBody(w, r, &req) {
		return
	}
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "auto"
	}
	if mode == "auto" {
		mode = fileIntakeAutoMode(req.Path)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var (
		content string
		isError bool
	)
	switch mode {
	case "document_text":
		args, err := json.Marshal(map[string]any{
			"path":      req.Path,
			"max_chars": req.MaxChars,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "encode document intake args")
			return
		}
		result, err := (&tools.DocumentTextTool{}).Run(ctx, string(args))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		content = result.Content
		isError = result.IsError
	case "archive_inspect":
		args, err := json.Marshal(map[string]any{
			"path":        req.Path,
			"max_entries": req.MaxEntries,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "encode archive intake args")
			return
		}
		result, err := (&tools.ArchiveInspectTool{}).Run(ctx, string(args))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		content = result.Content
		isError = result.IsError
	default:
		writeError(w, http.StatusBadRequest, "mode must be auto, document_text, or archive_inspect")
		return
	}

	status := "ok"
	if isError {
		status = "error"
	}
	writeJSON(w, http.StatusOK, fileIntakeResponse{
		Mode:    mode,
		Path:    req.Path,
		Status:  status,
		Content: content,
		IsError: isError,
	})
}

func fileIntakeAutoMode(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".zip"), strings.HasSuffix(lower, ".tar"), strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "archive_inspect"
	default:
		return "document_text"
	}
}
