package daemon

import (
	"encoding/json"
	"testing"

	"github.com/starclaw/starclaw/internal/config"
)

func TestChannelConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"ChannelCLI", ChannelCLI, "cli"},
		{"ChannelHTTP", ChannelHTTP, "http"},
		{"ChannelSchedule", ChannelSchedule, "schedule"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %q, want %q", tt.constant, tt.want)
			}
		})
	}
}

func TestRunAgentRequestJSON(t *testing.T) {
	req := RunAgentRequest{
		Text:    "hello",
		Agent:   "assistant",
		Channel: "cli",
		Sender:  "user1",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded RunAgentRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Text != "hello" {
		t.Errorf("Text = %q, want %q", decoded.Text, "hello")
	}
	if decoded.Agent != "assistant" {
		t.Errorf("Agent = %q, want %q", decoded.Agent, "assistant")
	}
	if decoded.Channel != "cli" {
		t.Errorf("Channel = %q, want %q", decoded.Channel, "cli")
	}
	if decoded.Sender != "user1" {
		t.Errorf("Sender = %q, want %q", decoded.Sender, "user1")
	}
}

func TestRunAgentRequestOmitEmpty(t *testing.T) {
	req := RunAgentRequest{
		Text: "hello",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if _, ok := raw["request_id"]; ok {
		t.Error("expected request_id to be omitted")
	}
	if _, ok := raw["model"]; ok {
		t.Error("expected model to be omitted")
	}
}

func TestRunAgentResponseJSON(t *testing.T) {
	resp := RunAgentResponse{
		SessionID: "sess_001",
		Messages:  []string{"hello", "world"},
		Usage:     map[string]int{"prompt_tokens": 10, "completion_tokens": 20},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded RunAgentResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.SessionID != "sess_001" {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, "sess_001")
	}
	if len(decoded.Messages) != 2 {
		t.Errorf("got %d messages, want 2", len(decoded.Messages))
	}
	if decoded.Usage["prompt_tokens"] != 10 {
		t.Errorf("prompt_tokens = %d, want 10", decoded.Usage["prompt_tokens"])
	}
}

func TestServerDepsDefaults(t *testing.T) {
	deps := ServerDeps{}
	if deps.StarclawDir != "" {
		t.Errorf("expected empty StarclawDir, got %q", deps.StarclawDir)
	}
}

func TestServerDepsFields(t *testing.T) {
	cfg := &config.Config{}
	deps := ServerDeps{
		StarclawDir:     "/home/user/.starclaw",
		ConfigPath:      "/home/user/.starclaw/config.yaml",
		Config:          cfg,
		AgentsDir:       "/home/user/.starclaw/agents",
		SkillsDir:       "/home/user/.starclaw/skills",
		InstructionsDir: "/home/user/.starclaw/instructions",
	}
	if deps.StarclawDir != "/home/user/.starclaw" {
		t.Errorf("StarclawDir = %q", deps.StarclawDir)
	}
	if deps.ConfigPath != "/home/user/.starclaw/config.yaml" {
		t.Errorf("ConfigPath = %q", deps.ConfigPath)
	}
	if deps.Config != cfg {
		t.Error("Config should keep the provided pointer")
	}
	if deps.AgentsDir != "/home/user/.starclaw/agents" {
		t.Errorf("AgentsDir = %q", deps.AgentsDir)
	}
	if deps.SkillsDir != "/home/user/.starclaw/skills" {
		t.Errorf("SkillsDir = %q", deps.SkillsDir)
	}
	if deps.InstructionsDir != "/home/user/.starclaw/instructions" {
		t.Errorf("InstructionsDir = %q", deps.InstructionsDir)
	}
}
