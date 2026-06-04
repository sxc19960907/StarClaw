package daemon

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type DiagnosticStatus string

const (
	DiagnosticReady      DiagnosticStatus = "ready"
	DiagnosticWarning    DiagnosticStatus = "warning"
	DiagnosticNeedsSetup DiagnosticStatus = "needs_setup"
	DiagnosticError      DiagnosticStatus = "error"
)

type DiagnosticsResponse struct {
	Status         DiagnosticStatus  `json:"status"`
	Summary        string            `json:"summary"`
	WebURL         string            `json:"web_url"`
	LaunchCommand  string            `json:"launch_command"`
	ExecutablePath string            `json:"executable_path,omitempty"`
	StarclawDir    string            `json:"starclaw_dir,omitempty"`
	ConfigPath     string            `json:"config_path,omitempty"`
	AgentsDir      string            `json:"agents_dir,omitempty"`
	SessionsDir    string            `json:"sessions_dir,omitempty"`
	Checks         []DiagnosticCheck `json:"checks"`
}

type DiagnosticCheck struct {
	ID     string           `json:"id"`
	Label  string           `json:"label"`
	Status DiagnosticStatus `json:"status"`
	Detail string           `json:"detail"`
	Action string           `json:"action,omitempty"`
}

func (s *Server) handleDiagnostics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.buildDiagnostics(r.Context()))
}

func (s *Server) buildDiagnostics(ctx context.Context) DiagnosticsResponse {
	starclawDir := s.depsPath(func(deps *ServerDeps) string { return deps.StarclawDir })
	configPath := s.depsPath(func(deps *ServerDeps) string { return deps.ConfigPath })
	agentsDir := s.depsPath(func(deps *ServerDeps) string { return deps.AgentsDir })
	sessionsDir := filepath.Join(starclawDir, "sessions")
	checks := []DiagnosticCheck{
		s.checkConfigFile(),
		s.checkProvider(ctx),
		s.checkDirectoryWritable("storage", "StarClaw directory", starclawDir),
		s.checkDirectoryWritable("sessions", "Sessions directory", sessionsDir),
		s.checkScheduleManager(),
		s.checkToolRegistry(),
		s.checkPermissions(),
	}
	status := highestDiagnosticStatus(checks)
	resp := DiagnosticsResponse{
		Status:        status,
		Summary:       diagnosticSummary(status),
		WebURL:        daemonWebURLForPort(s.port),
		LaunchCommand: daemonLaunchCommand,
		StarclawDir:   starclawDir,
		ConfigPath:    configPath,
		AgentsDir:     agentsDir,
		SessionsDir:   sessionsDir,
		Checks:        checks,
	}
	if executablePath, err := os.Executable(); err == nil {
		resp.ExecutablePath = executablePath
	}
	return resp
}

func (s *Server) depsPath(get func(*ServerDeps) string) string {
	if s.deps == nil {
		return ""
	}
	return get(s.deps)
}

func (s *Server) checkConfigFile() DiagnosticCheck {
	configPath := s.depsPath(func(deps *ServerDeps) string { return deps.ConfigPath })
	if strings.TrimSpace(configPath) == "" {
		return DiagnosticCheck{
			ID:     "config",
			Label:  "Config file",
			Status: DiagnosticError,
			Detail: "Config path is not available.",
			Action: "Restart the daemon after StarClaw config is initialized.",
		}
	}
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DiagnosticCheck{
				ID:     "config",
				Label:  "Config file",
				Status: DiagnosticNeedsSetup,
				Detail: fmt.Sprintf("Config file does not exist at %s.", configPath),
				Action: "Run starclaw setup.",
			}
		}
		return DiagnosticCheck{
			ID:     "config",
			Label:  "Config file",
			Status: DiagnosticError,
			Detail: fmt.Sprintf("Config file cannot be read: %v.", err),
			Action: "Check file permissions and restart the daemon.",
		}
	}
	if info.IsDir() {
		return DiagnosticCheck{
			ID:     "config",
			Label:  "Config file",
			Status: DiagnosticError,
			Detail: fmt.Sprintf("Config path points to a directory: %s.", configPath),
			Action: "Replace it with a StarClaw config file.",
		}
	}
	return DiagnosticCheck{
		ID:     "config",
		Label:  "Config file",
		Status: DiagnosticReady,
		Detail: fmt.Sprintf("Config file is available at %s.", configPath),
	}
}

func (s *Server) checkProvider(ctx context.Context) DiagnosticCheck {
	if s.deps == nil || s.deps.Config == nil {
		return DiagnosticCheck{
			ID:     "provider",
			Label:  "Provider",
			Status: DiagnosticError,
			Detail: "Loaded config is not available.",
			Action: "Restart the daemon after fixing the config file.",
		}
	}
	cfg := s.deps.Config
	provider := strings.TrimSpace(cfg.Provider)
	switch provider {
	case "", "anthropic":
		if strings.TrimSpace(cfg.Endpoint) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "Anthropic endpoint is missing.",
				Action: "Set endpoint in config.yaml.",
			}
		}
		if strings.TrimSpace(cfg.APIKey) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "Anthropic API key is missing.",
				Action: "Run starclaw setup or set api_key.",
			}
		}
		if strings.TrimSpace(cfg.ModelTier) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "Anthropic model tier is missing.",
				Action: "Set model_tier in config.yaml.",
			}
		}
		return DiagnosticCheck{
			ID:     "provider",
			Label:  "Provider",
			Status: DiagnosticReady,
			Detail: "Anthropic provider is configured.",
		}
	case "openai":
		if strings.TrimSpace(cfg.OpenAIEndpoint) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "OpenAI endpoint is missing.",
				Action: "Set openai_endpoint in config.yaml.",
			}
		}
		if strings.TrimSpace(cfg.OpenAIModel) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "OpenAI model is missing.",
				Action: "Set openai_model in config.yaml.",
			}
		}
		if strings.TrimSpace(cfg.OpenAIAPIKey) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "OpenAI API key is missing.",
				Action: "Run starclaw setup or set openai_api_key.",
			}
		}
		return DiagnosticCheck{
			ID:     "provider",
			Label:  "Provider",
			Status: DiagnosticReady,
			Detail: "OpenAI provider is configured.",
		}
	case "ollama":
		if strings.TrimSpace(cfg.OllamaEndpoint) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "Ollama endpoint is missing.",
				Action: "Set ollama_endpoint in config.yaml.",
			}
		}
		if strings.TrimSpace(cfg.OllamaModel) == "" {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticNeedsSetup,
				Detail: "Ollama model is missing.",
				Action: "Set ollama_model in config.yaml.",
			}
		}
		if err := probeOllama(ctx, cfg.OllamaEndpoint); err != nil {
			return DiagnosticCheck{
				ID:     "provider",
				Label:  "Provider",
				Status: DiagnosticWarning,
				Detail: fmt.Sprintf("Ollama is configured but not reachable: %v.", err),
				Action: "Start Ollama or update ollama_endpoint.",
			}
		}
		return DiagnosticCheck{
			ID:     "provider",
			Label:  "Provider",
			Status: DiagnosticReady,
			Detail: "Ollama provider is configured and reachable.",
		}
	default:
		return DiagnosticCheck{
			ID:     "provider",
			Label:  "Provider",
			Status: DiagnosticError,
			Detail: fmt.Sprintf("Unknown provider %q.", provider),
			Action: "Set provider to anthropic, openai, or ollama.",
		}
	}
}

func probeOllama(ctx context.Context, endpoint string) error {
	base, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}
	if base.Scheme == "" || base.Host == "" {
		return fmt.Errorf("endpoint must include scheme and host")
	}
	base.Path = path.Join(base.Path, "/api/tags")
	base.RawQuery = ""
	base.Fragment = ""

	probeCtx, cancel := context.WithTimeout(ctx, 350*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, base.String(), nil)
	if err != nil {
		return fmt.Errorf("create probe request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GET /api/tags returned %s", resp.Status)
	}
	return nil
}

func (s *Server) checkDirectoryWritable(id, label, dir string) DiagnosticCheck {
	if strings.TrimSpace(dir) == "" {
		return DiagnosticCheck{
			ID:     id,
			Label:  label,
			Status: DiagnosticError,
			Detail: "Directory path is not available.",
			Action: "Restart the daemon after StarClaw config is initialized.",
		}
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return DiagnosticCheck{
			ID:     id,
			Label:  label,
			Status: DiagnosticError,
			Detail: fmt.Sprintf("Directory cannot be created: %v.", err),
			Action: "Check filesystem permissions.",
		}
	}
	probe, err := os.CreateTemp(dir, ".starclaw-diagnostic-*.tmp")
	if err != nil {
		return DiagnosticCheck{
			ID:     id,
			Label:  label,
			Status: DiagnosticError,
			Detail: fmt.Sprintf("Directory is not writable: %v.", err),
			Action: "Check filesystem permissions.",
		}
	}
	probePath := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probePath)
		return DiagnosticCheck{
			ID:     id,
			Label:  label,
			Status: DiagnosticError,
			Detail: fmt.Sprintf("Write probe cannot be closed: %v.", err),
			Action: "Check filesystem permissions.",
		}
	}
	if err := os.Remove(probePath); err != nil {
		return DiagnosticCheck{
			ID:     id,
			Label:  label,
			Status: DiagnosticWarning,
			Detail: fmt.Sprintf("Directory is writable but cleanup failed: %v.", err),
			Action: "Remove stale diagnostic temp files if they accumulate.",
		}
	}
	return DiagnosticCheck{
		ID:     id,
		Label:  label,
		Status: DiagnosticReady,
		Detail: fmt.Sprintf("%s is writable.", dir),
	}
}

func (s *Server) checkScheduleManager() DiagnosticCheck {
	if s.deps == nil || s.deps.ScheduleManager == nil {
		return DiagnosticCheck{
			ID:     "schedules",
			Label:  "Schedules",
			Status: DiagnosticError,
			Detail: "Schedule manager is not available.",
			Action: "Restart the daemon.",
		}
	}
	if _, err := s.deps.ScheduleManager.List(); err != nil {
		return DiagnosticCheck{
			ID:     "schedules",
			Label:  "Schedules",
			Status: DiagnosticError,
			Detail: fmt.Sprintf("Schedules cannot be loaded: %v.", err),
			Action: "Check schedules.json permissions and contents.",
		}
	}
	return DiagnosticCheck{
		ID:     "schedules",
		Label:  "Schedules",
		Status: DiagnosticReady,
		Detail: "Schedule persistence is available.",
	}
}

func (s *Server) checkToolRegistry() DiagnosticCheck {
	if s.deps == nil || s.deps.Registry == nil {
		return DiagnosticCheck{
			ID:     "tools",
			Label:  "Tools",
			Status: DiagnosticError,
			Detail: "Tool registry is not available.",
			Action: "Restart the daemon.",
		}
	}
	count := s.deps.Registry.Count()
	if count == 0 {
		return DiagnosticCheck{
			ID:     "tools",
			Label:  "Tools",
			Status: DiagnosticWarning,
			Detail: "No tools are registered.",
			Action: "Restart the daemon and check tool registration.",
		}
	}
	return DiagnosticCheck{
		ID:     "tools",
		Label:  "Tools",
		Status: DiagnosticReady,
		Detail: fmt.Sprintf("%d tools are registered.", count),
	}
}

func (s *Server) checkPermissions() DiagnosticCheck {
	if s.deps == nil || s.deps.Config == nil || s.deps.Config.Permissions == nil {
		return DiagnosticCheck{
			ID:     "permissions",
			Label:  "Permissions",
			Status: DiagnosticWarning,
			Detail: "Permissions config is not present; daemon will rely on built-in defaults.",
			Action: "Add permissions rules to config.yaml for stricter local policy.",
		}
	}
	return DiagnosticCheck{
		ID:     "permissions",
		Label:  "Permissions",
		Status: DiagnosticReady,
		Detail: "Permissions config is present.",
	}
}

func highestDiagnosticStatus(checks []DiagnosticCheck) DiagnosticStatus {
	status := DiagnosticReady
	for _, check := range checks {
		if diagnosticSeverity(check.Status) > diagnosticSeverity(status) {
			status = check.Status
		}
	}
	return status
}

func diagnosticSeverity(status DiagnosticStatus) int {
	switch status {
	case DiagnosticError:
		return 3
	case DiagnosticNeedsSetup:
		return 2
	case DiagnosticWarning:
		return 1
	default:
		return 0
	}
}

func diagnosticSummary(status DiagnosticStatus) string {
	switch status {
	case DiagnosticError:
		return "Runtime diagnostics found a blocking error."
	case DiagnosticNeedsSetup:
		return "Provider setup is incomplete."
	case DiagnosticWarning:
		return "StarClaw can run, but some runtime checks need attention."
	default:
		return "StarClaw is ready to run agents."
	}
}
