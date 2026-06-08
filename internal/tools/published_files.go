package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/share"
)

type ListPublishedFilesTool struct{}

type listPublishedFilesArgs struct {
	IncludeRetracted bool `json:"include_retracted,omitempty"`
}

func NewListPublishedFilesTool() *ListPublishedFilesTool {
	return &ListPublishedFilesTool{}
}

func (t *ListPublishedFilesTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "list_published_files",
		Description: "List local files previously published with publish_to_web. Read-only. By default only active files are shown; set include_retracted=true to include retracted records.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"include_retracted": map[string]any{
					"type":        "boolean",
					"description": "Include records already retracted from the local web directory.",
				},
			},
		},
	}
}

func (t *ListPublishedFilesTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args listPublishedFilesArgs
	if strings.TrimSpace(argsJSON) != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
		}
	}
	_ = ctx
	store, err := newShareStore()
	if err != nil {
		return agent.TransientError(err.Error()), nil
	}
	artifacts, err := store.List(args.IncludeRetracted)
	if err != nil {
		return agent.TransientError(fmt.Sprintf("failed to list published files: %v", err)), nil
	}
	if len(artifacts) == 0 {
		if args.IncludeRetracted {
			return agent.ToolResult{Content: "No published files are tracked locally."}, nil
		}
		return agent.ToolResult{Content: "No active published files are tracked locally."}, nil
	}
	var sb strings.Builder
	for i, a := range artifacts {
		fmt.Fprintf(&sb, "[%d] %s\n", i+1, a.ID)
		fmt.Fprintf(&sb, "    %s (%d bytes, status=%s) created %s\n", a.Filename, a.SizeBytes, a.Status, a.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
		fmt.Fprintf(&sb, "    URL: %s\n", a.URL)
		fmt.Fprintf(&sb, "    Local path: %s\n", a.LocalPath)
		if a.Purpose != "" {
			fmt.Fprintf(&sb, "    Purpose: %s\n", a.Purpose)
		}
		if a.RetractedAt != nil {
			fmt.Fprintf(&sb, "    Retracted: %s\n", a.RetractedAt.Format("2006-01-02 15:04:05 UTC"))
		}
	}
	return agent.ToolResult{Content: strings.TrimRight(sb.String(), "\n")}, nil
}

func (t *ListPublishedFilesTool) RequiresApproval() bool     { return false }
func (t *ListPublishedFilesTool) IsReadOnlyCall(string) bool { return true }

type RetractPublishedFileTool struct{}

type retractPublishedFileArgs struct {
	ID string `json:"id"`
}

func NewRetractPublishedFileTool() *RetractPublishedFileTool {
	return &RetractPublishedFileTool{}
}

func (t *RetractPublishedFileTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "retract_published_file",
		Description: "Retract a local file previously published with publish_to_web. Removes the local web artifact directory and marks its manifest record retracted. Requires the artifact id from list_published_files or the publish_to_web result.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{
					"type":        "string",
					"description": "Artifact id returned by publish_to_web or list_published_files.",
				},
			},
		},
		Required: []string{"id"},
	}
}

func (t *RetractPublishedFileTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args retractPublishedFileArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}
	args.ID = strings.TrimSpace(args.ID)
	if args.ID == "" {
		return agent.ValidationError("id is required"), nil
	}
	_ = ctx
	store, err := newShareStore()
	if err != nil {
		return agent.TransientError(err.Error()), nil
	}
	artifact, already, err := store.Retract(args.ID)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return agent.ValidationError(fmt.Sprintf("published file not found: %s", args.ID)), nil
		}
		return agent.TransientError(fmt.Sprintf("failed to retract published file: %v", err)), nil
	}
	if already {
		return agent.ToolResult{Content: fmt.Sprintf("Already retracted.\nID: %s\nFilename: %s", artifact.ID, artifact.Filename)}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Retracted.\nID: %s\nFilename: %s\nLocal directory removed: %s", artifact.ID, artifact.Filename, filepath.Join(config.StarclawDir(), "web", artifact.ID))}, nil
}

func (t *RetractPublishedFileTool) RequiresApproval() bool { return true }

func newShareStore() (*share.Store, error) {
	starclawDir := config.StarclawDir()
	if starclawDir == "" {
		return nil, fmt.Errorf("cannot resolve StarClaw data directory")
	}
	return share.NewStore(starclawDir), nil
}
