package daemon

import (
	"fmt"
	"net/http"

	"github.com/starclaw/starclaw/internal/update"
)

type versionResponse struct {
	Version         string `json:"version"`
	Platform        string `json:"platform"`
	WebURL          string `json:"web_url"`
	UpdateSupported bool   `json:"update_supported"`
	UpdateCommand   string `json:"update_command"`
	Status          string `json:"status"`
	Message         string `json:"message"`
}

type updateCheckResponse struct {
	Version         string `json:"version"`
	Platform        string `json:"platform"`
	UpdateSupported bool   `json:"update_supported"`
	UpdateCommand   string `json:"update_command"`
	Status          string `json:"status"`
	Message         string `json:"message"`
	LatestVersion   string `json:"latest_version,omitempty"`
	ReleaseURL      string `json:"release_url,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
}

func (s *Server) versionInfo(status, message string) versionResponse {
	return versionResponse{
		Version:         s.version,
		Platform:        update.PlatformInfo(),
		WebURL:          daemonWebURLForPort(s.port),
		UpdateSupported: update.IsSemver(s.version),
		UpdateCommand:   "starclaw update --check",
		Status:          status,
		Message:         message,
	}
}

func daemonWebURLForPort(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/app/", port)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	status := "release"
	message := "Release build; update checks are available."
	if !update.IsSemver(s.version) {
		status = "development"
		message = "Development build - update checks require a release version."
	}
	writeJSON(w, http.StatusOK, s.versionInfo(status, message))
}

func (s *Server) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	if !update.IsSemver(s.version) {
		info := s.versionInfo("development", "Development build - update checks require a release version.")
		writeJSON(w, http.StatusOK, updateCheckResponse{
			Version:         info.Version,
			Platform:        info.Platform,
			UpdateSupported: info.UpdateSupported,
			UpdateCommand:   info.UpdateCommand,
			Status:          info.Status,
			Message:         info.Message,
		})
		return
	}

	release, found, err := update.CheckForUpdate(s.version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update check failed: "+err.Error())
		return
	}
	info := s.versionInfo("current", "You're already on the latest version.")
	resp := updateCheckResponse{
		Version:         info.Version,
		Platform:        info.Platform,
		UpdateSupported: info.UpdateSupported,
		UpdateCommand:   info.UpdateCommand,
		Status:          info.Status,
		Message:         info.Message,
	}
	if found && release != nil {
		resp.Status = "available"
		resp.Message = "Update available."
		resp.LatestVersion = release.TagName
		resp.ReleaseURL = release.HTMLURL
		resp.PublishedAt = release.PublishedAt
	}
	writeJSON(w, http.StatusOK, resp)
}
