package tools

import (
	"archive/tar"
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// ArchiveExtractTool safely extracts local zip, tar, and tar.gz archives.
type ArchiveExtractTool struct{}

type archiveExtractArgs struct {
	Path        string   `json:"path"`
	Destination string   `json:"destination"`
	Entries     []string `json:"entries,omitempty"`
	Overwrite   bool     `json:"overwrite,omitempty"`
}

type archiveExtractOutput struct {
	Path        string   `json:"path"`
	Destination string   `json:"destination"`
	Format      string   `json:"format"`
	Extracted   []string `json:"extracted"`
	Skipped     []string `json:"skipped,omitempty"`
}

func (t *ArchiveExtractTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "archive_extract",
		Description: "Safely extract zip, tar, or tar.gz archives. Blocks path traversal and does not overwrite existing files unless overwrite is true.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":        map[string]any{"type": "string", "description": "Absolute or relative archive path"},
				"destination": map[string]any{"type": "string", "description": "Directory to extract into"},
				"entries":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional exact archive entry names to extract"},
				"overwrite":   map[string]any{"type": "boolean", "description": "Overwrite existing files (default: false)"},
			},
		},
		Required: []string{"path", "destination"},
	}
}

func (t *ArchiveExtractTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args archiveExtractArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}
	if strings.TrimSpace(args.Path) == "" {
		return agent.ValidationError("path is required"), nil
	}
	if strings.TrimSpace(args.Destination) == "" {
		return agent.ValidationError("destination is required"), nil
	}

	args.Path = ExpandHome(args.Path)
	args.Destination = ExpandHome(args.Destination)
	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}
	if err := IsSafePath(args.Destination); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	destRoot, err := filepath.Abs(args.Destination)
	if err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid destination: %v", err)), nil
	}
	destRoot = filepath.Clean(destRoot)

	format := archiveFormat(args.Path)
	selected := make(map[string]bool, len(args.Entries))
	for _, entry := range args.Entries {
		selected[entry] = true
	}

	var result archiveExtractOutput
	result.Path = args.Path
	result.Destination = destRoot
	result.Format = format

	switch format {
	case "zip":
		result.Extracted, result.Skipped, err = extractZipArchive(args.Path, destRoot, selected, args.Overwrite)
	case "tar", "tar.gz":
		result.Extracted, result.Skipped, err = extractTarArchive(args.Path, format, destRoot, selected, args.Overwrite)
	default:
		return agent.ValidationError("unsupported archive format"), nil
	}
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("error extracting archive: %v", err), IsError: true}, nil
	}

	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("error formatting extraction result: %v", err), IsError: true}, nil
	}
	return agent.ToolResult{Content: string(payload)}, nil
}

func (t *ArchiveExtractTool) RequiresApproval() bool { return true }

func (t *ArchiveExtractTool) IsSafeArgs(argsJSON string) bool {
	var args archiveExtractArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return IsPathUnderCWD(args.Path) && IsPathUnderCWD(args.Destination)
}

func extractZipArchive(path, destRoot string, selected map[string]bool, overwrite bool) ([]string, []string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	var extracted []string
	var skipped []string
	for _, f := range zr.File {
		if !archiveEntrySelected(f.Name, selected) {
			continue
		}
		target, err := safeArchiveDestination(destRoot, f.Name)
		if err != nil {
			return nil, nil, err
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, nil, fmt.Errorf("create directory %s: %w", target, err)
			}
			continue
		}
		if !f.Mode().IsRegular() {
			skipped = append(skipped, f.Name)
			continue
		}
		if !overwrite {
			if _, err := os.Stat(target); err == nil {
				skipped = append(skipped, f.Name)
				continue
			} else if !os.IsNotExist(err) {
				return nil, nil, fmt.Errorf("check destination %s: %w", target, err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return nil, nil, fmt.Errorf("create parent directory %s: %w", filepath.Dir(target), err)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, nil, fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}
		if err := writeExtractedFile(target, rc, f.Mode()); err != nil {
			_ = rc.Close()
			return nil, nil, err
		}
		_ = rc.Close()
		extracted = append(extracted, f.Name)
	}
	return extracted, skipped, nil
}

func extractTarArchive(path, format, destRoot string, selected map[string]bool, overwrite bool) ([]string, []string, error) {
	r, closeFn, err := openTarReader(path, format)
	if err != nil {
		return nil, nil, err
	}
	defer closeFn()

	var extracted []string
	var skipped []string
	for {
		hdr, err := r.Next()
		if err == io.EOF {
			return extracted, skipped, nil
		}
		if err != nil {
			return nil, nil, fmt.Errorf("read tar entry: %w", err)
		}
		if !archiveEntrySelected(hdr.Name, selected) {
			continue
		}
		target, err := safeArchiveDestination(destRoot, hdr.Name)
		if err != nil {
			return nil, nil, err
		}
		mode := hdr.FileInfo().Mode()
		if hdr.FileInfo().IsDir() {
			if err := os.MkdirAll(target, mode.Perm()); err != nil {
				return nil, nil, fmt.Errorf("create directory %s: %w", target, err)
			}
			continue
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			skipped = append(skipped, hdr.Name)
			continue
		}
		if !overwrite {
			if _, err := os.Stat(target); err == nil {
				skipped = append(skipped, hdr.Name)
				continue
			} else if !os.IsNotExist(err) {
				return nil, nil, fmt.Errorf("check destination %s: %w", target, err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return nil, nil, fmt.Errorf("create parent directory %s: %w", filepath.Dir(target), err)
		}
		if err := writeExtractedFile(target, r, mode); err != nil {
			return nil, nil, err
		}
		extracted = append(extracted, hdr.Name)
	}
}

func writeExtractedFile(path string, r io.Reader, mode os.FileMode) error {
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to overwrite symlink: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check extracted file %s: %w", path, err)
	}

	perm := mode.Perm()
	if perm == 0 {
		perm = 0644
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("create extracted file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write extracted file %s: %w", path, err)
	}
	return nil
}
