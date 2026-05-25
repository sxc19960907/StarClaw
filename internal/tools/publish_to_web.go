package tools

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
)

// PublishToWebTool copies a local file to ~/.starclaw/web/ and returns a URL
// where the content can be viewed via the StarClaw daemon's HTTP server.
type PublishToWebTool struct{}

// NewPublishToWebTool creates a new PublishToWebTool.
func NewPublishToWebTool() *PublishToWebTool {
	return &PublishToWebTool{}
}

type publishArgs struct {
	Path    string `json:"path"`
	Purpose string `json:"purpose,omitempty"`
}

func (t *PublishToWebTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name: "publish_to_web",
		Description: "Publishes a local file to a web-accessible location.\n\n" +
			"Copies the file to ~/.starclaw/web/ and returns a URL that can be " +
			"used to view the content in a browser.\n\n" +
			"The daemon serves files from ~/.starclaw/web/ on http://localhost:7533.\n" +
			"If the daemon is not running, the file is still copied and can be " +
			"accessed directly from the local filesystem path shown in the result.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Local file path to publish (absolute or relative).",
				},
				"purpose": map[string]any{
					"type":        "string",
					"description": "Why the file needs to be published (optional, e.g. 'share report with user').",
				},
			},
		},
		Required: []string{"path"},
	}
}

func (t *PublishToWebTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args publishArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	path := ExpandHome(args.Path)
	if err := IsSafePath(path); err != nil {
		return agent.PermissionError(fmt.Sprintf("unsafe path: %v", err)), nil
	}

	// Verify the file exists and is a regular file
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ValidationError(fmt.Sprintf("file not found: %s", path)), nil
		}
		if os.IsPermission(err) {
			return agent.PermissionError(fmt.Sprintf("permission denied: %s", path)), nil
		}
		return agent.ValidationError(fmt.Sprintf("cannot access file: %v", err)), nil
	}
	if info.IsDir() {
		return agent.ValidationError(fmt.Sprintf("path is a directory, not a file: %s", path)), nil
	}

	// Generate a unique publication ID
	id, err := generateID()
	if err != nil {
		return agent.TransientError(fmt.Sprintf("failed to generate publication ID: %v", err)), nil
	}

	// Create the destination directory
	starclawDir := config.StarclawDir()
	if starclawDir == "" {
		return agent.TransientError("cannot resolve StarClaw data directory"), nil
	}

	pubDir := filepath.Join(starclawDir, "web", id)
	if err := os.MkdirAll(pubDir, 0755); err != nil {
		return agent.TransientError(fmt.Sprintf("failed to create output directory: %v", err)), nil
	}

	// Copy the file
	filename := filepath.Base(path)
	dst := filepath.Join(pubDir, filename)
	if err := copyFile(path, dst); err != nil {
		return agent.TransientError(fmt.Sprintf("failed to copy file: %v", err)), nil
	}

	// Build result
	url := fmt.Sprintf("http://localhost:7533/web/%s/%s", id, filename)
	result := fmt.Sprintf(
		"Published.\nURL: %s\nLocal path: %s\nSize: %d bytes",
		url, dst, info.Size(),
	)
	if args.Purpose != "" {
		result += fmt.Sprintf("\nPurpose: %s", args.Purpose)
	}
	result += "\n\nIf the StarClaw daemon is not running, access the file directly from the local path above."

	return agent.ToolResult{Content: result}, nil
}

func (t *PublishToWebTool) RequiresApproval() bool { return true }

// generateID produces a hex-encoded random identifier for a publication.
func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// copyFile copies a file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Stat the source to get permissions
	srcInfo, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}
