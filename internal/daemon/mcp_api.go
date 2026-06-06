package daemon

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/mcp"
)

type mcpTestRequest struct {
	Name string `json:"name"`
}

type mcpToolView struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type mcpTestResponse struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	ToolCount int           `json:"tool_count"`
	Tools     []mcpToolView `json:"tools"`
	Error     string        `json:"error,omitempty"`
}

func (s *Server) handleTestMCPServer(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil || s.deps.ConfigPath == "" {
		writeError(w, http.StatusInternalServerError, "config path not configured")
		return
	}
	var req mcpTestRequest
	if !decodeBody(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	cfg, err := readDaemonConfig(s.deps.ConfigPath, s.deps.Config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	server, ok := cfg.MCPServers[name]
	if !ok {
		writeError(w, http.StatusNotFound, "MCP server not found")
		return
	}
	if server.Disabled {
		writeJSON(w, http.StatusOK, mcpTestResponse{Name: name, Status: "disabled", Error: "server is disabled"})
		return
	}

	remoteTools, err := s.testMCPServer(r, name, server)
	if err != nil {
		writeJSON(w, http.StatusOK, mcpTestResponse{Name: name, Status: "error", Error: err.Error()})
		return
	}
	tools := make([]mcpToolView, 0, len(remoteTools))
	for _, tool := range remoteTools {
		tools = append(tools, mcpToolView{
			Name:        tool.Tool.Name,
			Description: tool.Tool.Description,
		})
	}
	writeJSON(w, http.StatusOK, mcpTestResponse{
		Name:      name,
		Status:    "ok",
		ToolCount: len(tools),
		Tools:     tools,
	})
}

func (s *Server) testMCPServer(r *http.Request, name string, server mcp.MCPServerConfig) ([]mcp.RemoteTool, error) {
	if s.deps != nil && s.deps.MCPTester != nil {
		return s.deps.MCPTester(name, server)
	}
	manager := mcp.NewClientManager()
	defer manager.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	return manager.ConnectAll(ctx, map[string]mcp.MCPServerConfig{name: server})
}
