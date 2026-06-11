package daemon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/config"
)

type configConnectionTestResponse struct {
	Status     string `json:"status"`
	Code       string `json:"code,omitempty"`
	Provider   string `json:"provider"`
	Model      string `json:"model,omitempty"`
	Detail     string `json:"detail"`
	DurationMS int64  `json:"duration_ms,omitempty"`
}

func (s *Server) handleTestConfig(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil || s.deps.ConfigPath == "" {
		writeError(w, http.StatusInternalServerError, "config path not configured")
		return
	}
	cfg, err := readDaemonConfig(s.deps.ConfigPath, s.deps.Config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := testProviderConnection(r.Context(), cfg)
	writeJSON(w, http.StatusOK, resp)
}

func testProviderConnection(ctx context.Context, cfg *config.Config) configConnectionTestResponse {
	provider := strings.TrimSpace(cfg.Provider)
	if provider == "" {
		provider = "anthropic"
	}
	model := providerModelName(provider, cfg)
	missing := missingProviderConnectionFields(provider, cfg)
	if len(missing) > 0 {
		return configConnectionTestResponse{
			Status:   "needs_setup",
			Code:     "missing_fields",
			Provider: provider,
			Model:    model,
			Detail:   fmt.Sprintf("缺少 %s。", strings.Join(missing, "、")),
		}
	}

	llm, err := providerConnectionClient(provider, cfg)
	if err != nil {
		return configConnectionTestResponse{
			Status:   "error",
			Code:     "unsupported_provider",
			Provider: provider,
			Model:    model,
			Detail:   err.Error(),
		}
	}

	testCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	started := time.Now()
	_, err = llm.Chat(testCtx, "", []client.Message{{Role: "user", Content: "Reply with ok."}}, nil, 8, nil)
	if err != nil {
		classification := classifyProviderConnectionError(err, cfg)
		return configConnectionTestResponse{
			Status:     "error",
			Code:       classification.code,
			Provider:   provider,
			Model:      model,
			Detail:     classification.detail,
			DurationMS: time.Since(started).Milliseconds(),
		}
	}
	return configConnectionTestResponse{
		Status:     "ready",
		Code:       "ok",
		Provider:   provider,
		Model:      model,
		Detail:     "连接成功，模型返回了有效响应。",
		DurationMS: time.Since(started).Milliseconds(),
	}
}

func providerModelName(provider string, cfg *config.Config) string {
	switch provider {
	case "openai":
		return strings.TrimSpace(cfg.OpenAIModel)
	case "ollama":
		return strings.TrimSpace(cfg.OllamaModel)
	default:
		return strings.TrimSpace(cfg.ModelTier)
	}
}

func missingProviderConnectionFields(provider string, cfg *config.Config) []string {
	var missing []string
	switch provider {
	case "openai":
		if strings.TrimSpace(cfg.OpenAIEndpoint) == "" {
			missing = append(missing, "Base URL")
		}
		if strings.TrimSpace(cfg.OpenAIModel) == "" {
			missing = append(missing, "Model")
		}
		if strings.TrimSpace(cfg.OpenAIAPIKey) == "" {
			missing = append(missing, "API key")
		}
	case "ollama":
		if strings.TrimSpace(cfg.OllamaEndpoint) == "" {
			missing = append(missing, "Base URL")
		}
		if strings.TrimSpace(cfg.OllamaModel) == "" {
			missing = append(missing, "Model")
		}
	default:
		if strings.TrimSpace(cfg.Endpoint) == "" {
			missing = append(missing, "Base URL")
		}
		if strings.TrimSpace(cfg.ModelTier) == "" {
			missing = append(missing, "Model")
		}
		if strings.TrimSpace(cfg.APIKey) == "" {
			missing = append(missing, "API key")
		}
	}
	return missing
}

func providerConnectionClient(provider string, cfg *config.Config) (client.LLMClient, error) {
	switch provider {
	case "openai":
		return client.NewOpenAIClient(strings.TrimSpace(cfg.OpenAIAPIKey), strings.TrimSpace(cfg.OpenAIEndpoint), strings.TrimSpace(cfg.OpenAIModel)), nil
	case "ollama":
		return client.NewOllamaClient(strings.TrimSpace(cfg.OllamaEndpoint), strings.TrimSpace(cfg.OllamaModel)), nil
	case "anthropic", "":
		return client.NewAnthropicClient(strings.TrimSpace(cfg.APIKey), strings.TrimSpace(cfg.Endpoint), strings.TrimSpace(cfg.ModelTier)), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", provider)
	}
}

func sanitizeProviderConnectionError(message string, cfg *config.Config) string {
	out := strings.TrimSpace(message)
	for _, secret := range []string{cfg.APIKey, cfg.OpenAIAPIKey} {
		secret = strings.TrimSpace(secret)
		if secret != "" {
			out = strings.ReplaceAll(out, "Bearer "+secret, "Bearer [REDACTED]")
			out = strings.ReplaceAll(out, secret, "[REDACTED]")
		}
	}
	if len(out) > 360 {
		out = out[:360] + "..."
	}
	if out == "" {
		return "provider returned an empty error"
	}
	return out
}

type providerConnectionErrorClassification struct {
	code   string
	detail string
}

func classifyProviderConnectionError(err error, cfg *config.Config) providerConnectionErrorClassification {
	raw := sanitizeProviderConnectionError(err.Error(), cfg)
	lower := strings.ToLower(raw)
	status := providerErrorStatusCode(lower)
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline exceeded") {
		return providerConnectionErrorClassification{
			code:   "timeout",
			detail: "连接超时。请检查 Base URL 是否可访问，或稍后重试。",
		}
	}
	if strings.Contains(lower, "request failed") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "no such host") ||
		strings.Contains(lower, "network is unreachable") ||
		strings.Contains(lower, "server misbehaving") {
		return providerConnectionErrorClassification{
			code:   "network_unreachable",
			detail: "无法连接到 Base URL。请检查地址、端口、代理或本地服务是否已启动。",
		}
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return providerConnectionErrorClassification{
			code:   "auth_failed",
			detail: "认证失败。请检查 API key 是否正确、是否有权限访问该 provider。",
		}
	}
	if status == http.StatusTooManyRequests || strings.Contains(lower, "rate limit") || strings.Contains(lower, "too many requests") {
		return providerConnectionErrorClassification{
			code:   "rate_limited",
			detail: "provider 返回限流。请稍后重试，或检查额度与频率限制。",
		}
	}
	if status == http.StatusNotFound ||
		strings.Contains(lower, "model not found") ||
		strings.Contains(lower, "model_not_found") ||
		strings.Contains(lower, "does not exist") ||
		strings.Contains(lower, "unknown model") {
		return providerConnectionErrorClassification{
			code:   "model_not_found",
			detail: "模型不可用。请检查模型名是否拼写正确，或该 API key 是否有模型访问权限。",
		}
	}
	if strings.Contains(lower, "failed to decode response") ||
		strings.Contains(lower, "no choices in response") ||
		strings.Contains(lower, "unexpected choice type") ||
		strings.Contains(lower, "missing message in choice") ||
		strings.Contains(lower, "invalid character") {
		return providerConnectionErrorClassification{
			code:   "invalid_response",
			detail: "provider 返回格式不兼容。请确认 Base URL 指向兼容的 Chat Completions 或 Messages API。",
		}
	}
	if status >= 500 {
		return providerConnectionErrorClassification{
			code:   "provider_unavailable",
			detail: "provider 暂时不可用或返回服务端错误。请稍后重试。",
		}
	}
	return providerConnectionErrorClassification{
		code:   "provider_error",
		detail: raw,
	}
}

func providerErrorStatusCode(message string) int {
	start := strings.Index(message, "api error (")
	if start < 0 {
		return 0
	}
	start += len("api error (")
	end := strings.Index(message[start:], ")")
	if end < 0 {
		return 0
	}
	status, err := strconv.Atoi(message[start : start+end])
	if err != nil {
		return 0
	}
	return status
}
