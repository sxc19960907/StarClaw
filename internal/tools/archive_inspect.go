package tools

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// ArchiveInspectTool lists entries in local zip, tar, and tar.gz archives.
type ArchiveInspectTool struct{}

type archiveInspectArgs struct {
	Path       string `json:"path"`
	MaxEntries int    `json:"max_entries,omitempty"`
}

type archiveEntryInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Mode  string `json:"mode,omitempty"`
	IsDir bool   `json:"is_dir"`
}

type archiveInspectOutput struct {
	Path      string             `json:"path"`
	Format    string             `json:"format"`
	Entries   []archiveEntryInfo `json:"entries"`
	Truncated bool               `json:"truncated"`
}

func (t *ArchiveInspectTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "archive_inspect",
		Description: "Inspect zip, tar, or tar.gz archives without extracting them. Returns structured entry names, sizes, and directory flags.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":        map[string]any{"type": "string", "description": "Absolute or relative archive path"},
				"max_entries": map[string]any{"type": "integer", "description": "Maximum entries to return (default: 200)"},
			},
		},
		Required: []string{"path"},
	}
}

func (t *ArchiveInspectTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args archiveInspectArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}
	if strings.TrimSpace(args.Path) == "" {
		return agent.ValidationError("path is required"), nil
	}
	args.Path = ExpandHome(args.Path)
	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	maxEntries := args.MaxEntries
	if maxEntries <= 0 {
		maxEntries = 200
	}

	format := archiveFormat(args.Path)
	entries, truncated, err := inspectArchive(args.Path, format, maxEntries)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("error inspecting archive: %v", err), IsError: true}, nil
	}

	payload, err := json.MarshalIndent(archiveInspectOutput{
		Path:      args.Path,
		Format:    format,
		Entries:   entries,
		Truncated: truncated,
	}, "", "  ")
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("error formatting archive result: %v", err), IsError: true}, nil
	}
	return agent.ToolResult{Content: string(payload)}, nil
}

func (t *ArchiveInspectTool) RequiresApproval() bool { return true }

func (t *ArchiveInspectTool) IsReadOnlyCall(string) bool { return true }

func (t *ArchiveInspectTool) IsSafeArgs(argsJSON string) bool {
	var args archiveInspectArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return IsPathUnderCWD(args.Path)
}

func archiveFormat(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".tar"):
		return "tar"
	case strings.HasSuffix(lower, ".zip"), strings.HasSuffix(lower, ".docx"), strings.HasSuffix(lower, ".xlsx"), strings.HasSuffix(lower, ".pptx"):
		return "zip"
	default:
		return ""
	}
}

func inspectArchive(path, format string, maxEntries int) ([]archiveEntryInfo, bool, error) {
	switch format {
	case "zip":
		return inspectZipArchive(path, maxEntries)
	case "tar", "tar.gz":
		return inspectTarArchive(path, format, maxEntries)
	default:
		return nil, false, fmt.Errorf("unsupported archive format")
	}
}

func inspectZipArchive(path string, maxEntries int) ([]archiveEntryInfo, bool, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, false, fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	var entries []archiveEntryInfo
	for i, f := range zr.File {
		if i >= maxEntries {
			return entries, true, nil
		}
		entries = append(entries, archiveEntryInfo{
			Name:  f.Name,
			Size:  int64(f.UncompressedSize64),
			Mode:  f.Mode().String(),
			IsDir: f.FileInfo().IsDir(),
		})
	}
	return entries, false, nil
}

func inspectTarArchive(path, format string, maxEntries int) ([]archiveEntryInfo, bool, error) {
	r, closeFn, err := openTarReader(path, format)
	if err != nil {
		return nil, false, err
	}
	defer closeFn()

	var entries []archiveEntryInfo
	for {
		hdr, err := r.Next()
		if err == io.EOF {
			return entries, false, nil
		}
		if err != nil {
			return nil, false, fmt.Errorf("read tar entry: %w", err)
		}
		if len(entries) >= maxEntries {
			return entries, true, nil
		}
		entries = append(entries, archiveEntryInfo{
			Name:  hdr.Name,
			Size:  hdr.Size,
			Mode:  hdr.FileInfo().Mode().String(),
			IsDir: hdr.FileInfo().IsDir(),
		})
	}
}

func openTarReader(path, format string) (*tar.Reader, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, func() {}, fmt.Errorf("open archive: %w", err)
	}
	closeFn := func() { _ = f.Close() }

	var r io.Reader = f
	if format == "tar.gz" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			closeFn()
			return nil, func() {}, fmt.Errorf("open gzip stream: %w", err)
		}
		closeFn = func() {
			_ = gz.Close()
			_ = f.Close()
		}
		r = gz
	}
	return tar.NewReader(r), closeFn, nil
}

func safeArchiveDestination(destRoot, entryName string) (string, error) {
	if strings.TrimSpace(entryName) == "" {
		return "", fmt.Errorf("empty archive entry name")
	}
	cleanEntry := filepath.Clean(entryName)
	if filepath.IsAbs(cleanEntry) || cleanEntry == "." || cleanEntry == ".." || strings.HasPrefix(cleanEntry, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive entry path: %s", entryName)
	}
	target := filepath.Join(destRoot, cleanEntry)
	rel, err := filepath.Rel(destRoot, target)
	if err != nil {
		return "", fmt.Errorf("check archive destination: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("archive entry escapes destination: %s", entryName)
	}
	return target, nil
}

func archiveEntrySelected(name string, selected map[string]bool) bool {
	if len(selected) == 0 {
		return true
	}
	return selected[name]
}
