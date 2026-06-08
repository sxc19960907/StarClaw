package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/share"
)

func TestListPublishedFilesTool(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	store := share.NewStore(config.StarclawDir())
	activeDir := filepath.Join(config.StarclawDir(), "web", "active-id")
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("mkdir active dir: %v", err)
	}
	if _, err := store.Record(share.Artifact{
		ID:        "active-id",
		Filename:  "active.html",
		LocalPath: filepath.Join(activeDir, "active.html"),
		URL:       "http://localhost:7533/web/active-id/active.html",
		SizeBytes: 12,
		Status:    share.StatusActive,
		Purpose:   "demo",
	}); err != nil {
		t.Fatalf("record active: %v", err)
	}
	retractedDir := filepath.Join(config.StarclawDir(), "web", "retracted-id")
	if err := os.MkdirAll(retractedDir, 0o755); err != nil {
		t.Fatalf("mkdir retracted dir: %v", err)
	}
	if _, err := store.Record(share.Artifact{
		ID:        "retracted-id",
		Filename:  "old.html",
		LocalPath: filepath.Join(retractedDir, "old.html"),
		URL:       "http://localhost:7533/web/retracted-id/old.html",
		SizeBytes: 9,
		Status:    share.StatusActive,
	}); err != nil {
		t.Fatalf("record retracted: %v", err)
	}
	if _, _, err := store.Retract("retracted-id"); err != nil {
		t.Fatalf("retract setup: %v", err)
	}

	tool := NewListPublishedFilesTool()
	result, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.IsError {
		t.Fatalf("Run error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "active-id") {
		t.Fatalf("expected active id, got: %s", result.Content)
	}
	if strings.Contains(result.Content, "retracted-id") {
		t.Fatalf("default list should hide retracted records: %s", result.Content)
	}

	result, err = tool.Run(context.Background(), `{"include_retracted":true}`)
	if err != nil {
		t.Fatalf("Run include retracted: %v", err)
	}
	if !strings.Contains(result.Content, "retracted-id") || !strings.Contains(result.Content, "status=retracted") {
		t.Fatalf("expected retracted record, got: %s", result.Content)
	}
}

func TestRetractPublishedFileTool(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)
	starclawDir := config.StarclawDir()
	artifactDir := filepath.Join(starclawDir, "web", "artifact-id")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(artifactDir, "file.html"), []byte("html"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	store := share.NewStore(starclawDir)
	if _, err := store.Record(share.Artifact{
		ID:        "artifact-id",
		Filename:  "file.html",
		LocalPath: filepath.Join(artifactDir, "file.html"),
		URL:       "http://localhost:7533/web/artifact-id/file.html",
		SizeBytes: 4,
	}); err != nil {
		t.Fatalf("record artifact: %v", err)
	}

	tool := NewRetractPublishedFileTool()
	result, err := tool.Run(context.Background(), `{"id":"artifact-id"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.IsError {
		t.Fatalf("Run error: %s", result.Content)
	}
	if !strings.Contains(result.Content, "Retracted") {
		t.Fatalf("expected retracted message, got: %s", result.Content)
	}
	if _, err := os.Stat(artifactDir); !os.IsNotExist(err) {
		t.Fatalf("artifact dir should be removed, stat err: %v", err)
	}

	result, err = tool.Run(context.Background(), `{"id":"artifact-id"}`)
	if err != nil {
		t.Fatalf("Run second: %v", err)
	}
	if !strings.Contains(result.Content, "Already retracted") {
		t.Fatalf("expected idempotent message, got: %s", result.Content)
	}
}

func TestRetractPublishedFileToolMissingID(t *testing.T) {
	tool := NewRetractPublishedFileTool()
	result, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.IsError || !strings.Contains(result.Content, "id is required") {
		t.Fatalf("expected id validation, got: %#v", result)
	}
}
