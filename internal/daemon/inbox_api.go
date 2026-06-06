package daemon

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

type inboxProviderView struct {
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	Endpoint         string   `json:"endpoint"`
	Configured       bool     `json:"configured"`
	SecretConfigured bool     `json:"secret_configured"`
	SupportedEvents  []string `json:"supported_events"`
	Description      string   `json:"description"`
}

type githubWebhookPayload struct {
	Action string `json:"action"`
	Issue  struct {
		ID      int64  `json:"id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"issue"`
	Comment struct {
		ID      int64  `json:"id"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"comment"`
	Repository struct {
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
}

func (s *Server) handleListInbox(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.inboxStore.List()})
}

func (s *Server) handleInboxProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"providers": []inboxProviderView{
		{
			Name:             "Local webhook",
			Kind:             "webhook",
			Endpoint:         "/inbox/webhook",
			Configured:       true,
			SecretConfigured: false,
			SupportedEvents:  []string{"generic"},
			Description:      "Local JSON intake for tests and manual channel bridges.",
		},
		{
			Name:             "GitHub",
			Kind:             "github",
			Endpoint:         "/inbox/github",
			Configured:       true,
			SecretConfigured: githubWebhookSecret() != "",
			SupportedEvents:  []string{"issues", "issue_comment"},
			Description:      "GitHub issue and comment webhooks enter the guarded Inbox.",
		},
	}})
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

func (s *Server) handleInboxGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read webhook body")
		return
	}
	if err := verifyGitHubSignature(raw, r.Header.Get("X-Hub-Signature-256")); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var payload githubWebhookPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid GitHub webhook payload")
		return
	}
	item, status, err := inboxItemFromGitHubWebhook(r.Header.Get("X-GitHub-Event"), r.Header.Get("X-GitHub-Delivery"), payload)
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

func inboxItemFromGitHubWebhook(event, delivery string, payload githubWebhookPayload) (InboxItem, int, error) {
	event = strings.TrimSpace(event)
	action := strings.TrimSpace(payload.Action)
	repo := strings.TrimSpace(payload.Repository.FullName)
	if event == "" {
		return InboxItem{}, 0, errors.New("X-GitHub-Event is required")
	}
	if repo == "" {
		return InboxItem{}, 0, errors.New("repository.full_name is required")
	}
	switch event {
	case "issues":
		if payload.Issue.ID == 0 {
			return InboxItem{}, 0, errors.New("issue.id is required")
		}
		title := strings.TrimSpace(payload.Issue.Title)
		body := strings.TrimSpace(payload.Issue.Body)
		text := strings.TrimSpace(fmt.Sprintf("GitHub issue %s#%d %s\n\n%s", repo, payload.Issue.Number, title, body))
		if body == "" {
			text = strings.TrimSpace(fmt.Sprintf("GitHub issue %s#%d %s", repo, payload.Issue.Number, title))
		}
		return InboxItem{
			Provider:   "github",
			ExternalID: fmt.Sprintf("issue:%s:%d:%s", repo, payload.Issue.ID, action),
			Sender:     firstNonEmpty(payload.Issue.User.Login, payload.Sender.Login),
			Text:       text,
			Status:     "pending",
			Metadata: cleanInboxMetadata(map[string]string{
				"event":      event,
				"action":     action,
				"delivery":   delivery,
				"repository": repo,
				"issue":      fmt.Sprintf("%d", payload.Issue.Number),
				"html_url":   payload.Issue.HTMLURL,
				"repo_url":   payload.Repository.HTMLURL,
				"sender":     payload.Sender.Login,
			}),
		}, http.StatusCreated, nil
	case "issue_comment":
		if payload.Comment.ID == 0 {
			return InboxItem{}, 0, errors.New("comment.id is required")
		}
		text := strings.TrimSpace(fmt.Sprintf("GitHub comment on %s#%d\n\n%s", repo, payload.Issue.Number, strings.TrimSpace(payload.Comment.Body)))
		return InboxItem{
			Provider:   "github",
			ExternalID: fmt.Sprintf("issue_comment:%s:%d:%s", repo, payload.Comment.ID, action),
			Sender:     firstNonEmpty(payload.Comment.User.Login, payload.Sender.Login),
			Text:       text,
			Status:     "pending",
			Metadata: cleanInboxMetadata(map[string]string{
				"event":      event,
				"action":     action,
				"delivery":   delivery,
				"repository": repo,
				"issue":      fmt.Sprintf("%d", payload.Issue.Number),
				"html_url":   payload.Comment.HTMLURL,
				"repo_url":   payload.Repository.HTMLURL,
				"sender":     payload.Sender.Login,
			}),
		}, http.StatusCreated, nil
	default:
		return InboxItem{}, 0, fmt.Errorf("unsupported GitHub event %q", event)
	}
}

func verifyGitHubSignature(body []byte, signature string) error {
	secret := githubWebhookSecret()
	if secret == "" {
		return nil
	}
	signature = strings.TrimSpace(signature)
	if !strings.HasPrefix(signature, "sha256=") {
		return errors.New("invalid GitHub webhook signature")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("invalid GitHub webhook signature")
	}
	return nil
}

func githubWebhookSecret() string {
	return strings.TrimSpace(os.Getenv("STARCLAW_GITHUB_WEBHOOK_SECRET"))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
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
