package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemorySidecarStatusEmptyDir(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)

	status := s.memorySidecarStatus()
	if status.Provider != MemoryProviderDisabled || status.Ready {
		t.Fatalf("status = %#v, want disabled/not ready", status)
	}
	if status.Reason != MemoryReasonNoLocalMemory {
		t.Fatalf("reason = %q, want %q", status.Reason, MemoryReasonNoLocalMemory)
	}
}

func TestMemorySidecarStatusWithMemoryFacts(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	if err := os.MkdirAll(s.memoryDir(), 0o700); err != nil {
		t.Fatalf("mkdir memory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(s.memoryDir(), "MEMORY.md"), []byte("- preference: User likes local-first runtime.\n"), 0o600); err != nil {
		t.Fatalf("write memory: %v", err)
	}

	status := s.memorySidecarStatus()
	if status.Provider != MemoryProviderLocal || !status.Ready {
		t.Fatalf("status = %#v, want local ready", status)
	}
	if status.LocalFacts != 1 || !status.FallbackAvailable {
		t.Fatalf("status facts/fallback = %#v", status)
	}
}

func TestMemorySidecarStatusWithBundleManifest(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	bundleDir := filepath.Join(s.memoryDir(), "bundles", "2026-06-08T00-00-00Z")
	if err := os.MkdirAll(bundleDir, 0o700); err != nil {
		t.Fatalf("mkdir bundle: %v", err)
	}
	manifest := `{"bundle_version":"0.6.0","bundle_created_at":"2026-06-08T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(bundleDir, "manifest.json"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.Symlink(bundleDir, filepath.Join(s.memoryDir(), "current")); err != nil {
		t.Fatalf("symlink current: %v", err)
	}

	status := s.memorySidecarStatus()
	if !status.Ready || status.Bundle == nil || status.Bundle.Version != "0.6.0" {
		t.Fatalf("status = %#v, want ready bundle", status)
	}
}

func TestMemorySidecarStatusMalformedBundle(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	bundleDir := filepath.Join(s.memoryDir(), "bundles", "bad")
	if err := os.MkdirAll(bundleDir, 0o700); err != nil {
		t.Fatalf("mkdir bundle: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, "manifest.json"), []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.Symlink(bundleDir, filepath.Join(s.memoryDir(), "current")); err != nil {
		t.Fatalf("symlink current: %v", err)
	}

	status := s.memorySidecarStatus()
	if status.Provider != MemoryProviderDegraded || status.Reason != MemoryReasonBundleInvalid {
		t.Fatalf("status = %#v, want degraded invalid bundle", status)
	}
}

func TestMemoryRecallMatchesLocalFacts(t *testing.T) {
	deps := newTestServerDeps(t)
	s := newTestServer(t, deps)
	if err := os.MkdirAll(s.memoryDir(), 0o700); err != nil {
		t.Fatalf("mkdir memory: %v", err)
	}
	content := strings.Join([]string{
		"## Preferences",
		"- User likes local-first runtime.",
		"## Decisions",
		"- Use daemon queue before cloud transport.",
	}, "\n")
	if err := os.WriteFile(filepath.Join(s.memoryDir(), "MEMORY.md"), []byte(content), 0o600); err != nil {
		t.Fatalf("write memory: %v", err)
	}

	result := s.recallMemory("local first")
	if result.Outcome != MemoryRecallOutcomeMatched || len(result.Results) != 1 {
		t.Fatalf("recall = %#v, want one match", result)
	}
	if !strings.Contains(result.Results[0].Text, "local-first") {
		t.Fatalf("result = %#v", result.Results[0])
	}
}

func TestMemoryRecallNoMemoryReason(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	result := s.recallMemory("anything")
	if result.Outcome != MemoryRecallOutcomeNoData || result.Reason != MemoryReasonNoLocalMemory {
		t.Fatalf("recall = %#v, want no local memory reason", result)
	}
}

func TestPrivateMemoryBlock(t *testing.T) {
	result := MemoryRecallResult{
		Outcome: MemoryRecallOutcomeMatched,
		Results: []MemoryRecallMatch{
			{Category: "preferences", Text: "User likes local-first runtime.", Entry: "MEMORY.md", Line: 1},
		},
	}
	block := privateMemoryBlock(result)
	if !strings.Contains(block, "<private_memory>") || !strings.Contains(block, "local-first") || !strings.Contains(block, "</private_memory>") {
		t.Fatalf("private memory block = %q", block)
	}
}

func TestMemoryRecallTelemetryIsContentFree(t *testing.T) {
	result := MemoryRecallResult{
		Attempted:       true,
		Provider:        MemoryProviderLocal,
		Outcome:         MemoryRecallOutcomeMatched,
		ContextInjected: true,
		Results: []MemoryRecallMatch{
			{Text: "User likes local-first runtime.", Entry: "MEMORY.md", Line: 1},
		},
	}
	data, err := json.Marshal(memoryPreflightEventData(result))
	if err != nil {
		t.Fatalf("marshal telemetry: %v", err)
	}
	if strings.Contains(string(data), "local-first") || strings.Contains(string(data), "MEMORY.md") {
		t.Fatalf("telemetry leaked content: %s", data)
	}
	if !strings.Contains(string(data), `"results_count":1`) {
		t.Fatalf("telemetry missing result count: %s", data)
	}
}
