package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
)

// ImagingTool processes images: describe, resize, convert.
type ImagingTool struct{}

type imagingArgs struct {
	Action string `json:"action"`
	Path   string `json:"path"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Format string `json:"format,omitempty"`
}

func (t *ImagingTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "imaging",
		Description: "Process images: describe (metadata + OCR), resize (width, height), convert (format).",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{"type": "string", "description": "Action: 'describe', 'resize', or 'convert'"},
				"path":   map[string]any{"type": "string", "description": "Path to the image file"},
				"width":  map[string]any{"type": "integer", "description": "Target width in pixels (for resize)"},
				"height": map[string]any{"type": "integer", "description": "Target height in pixels (for resize)"},
				"format": map[string]any{"type": "string", "description": "Target format: 'jpeg', 'png', 'gif', 'tiff', 'pdf' (for convert)"},
			},
		},
		Required: []string{"action", "path"},
	}
}

func (t *ImagingTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args imagingArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	if args.Action == "" {
		return agent.ValidationError("action is required (describe, resize, convert)"), nil
	}
	if args.Path == "" {
		return agent.ValidationError("path is required"), nil
	}

	// Expand and validate path
	args.Path = ExpandHome(args.Path)
	if err := IsSafePath(args.Path); err != nil {
		return agent.ValidationError("unsafe path: " + err.Error()), nil
	}

	// Validate action-specific args before checking file existence
	// so that validation errors are returned even when the file doesn't exist.
	switch args.Action {
	case "describe":
		// No additional args needed
		case "resize":
			if args.Width <= 0 && args.Height <= 0 {
				return agent.ValidationError("at least one of width or height must be positive for resize"), nil
			}
	case "convert":
		if args.Format == "" {
			return agent.ValidationError("format is required for convert action"), nil
		}
	default:
		return agent.ToolResult{Content: fmt.Sprintf("unknown action: %q (use 'describe', 'resize', or 'convert')", args.Action), IsError: true}, nil
	}

	// Check if file exists
	if _, err := os.Stat(args.Path); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("file not found: %v", err), IsError: true}, nil
	}

	switch args.Action {
	case "describe":
		return t.describe(ctx, args)
	case "resize":
		return t.resize(ctx, args)
	case "convert":
		return t.convert(ctx, args)
	}

	return agent.ToolResult{Content: "unexpected error: no handler for action", IsError: true}, nil
}

func (t *ImagingTool) RequiresApproval() bool { return true }

// describe returns image metadata and runs OCR when possible.
func (t *ImagingTool) describe(ctx context.Context, args imagingArgs) (agent.ToolResult, error) {
	var infoParts []string

	// Get metadata using platform-specific tools
	switch runtime.GOOS {
	case "darwin":
		meta, err := sipsDescribe(ctx, args.Path)
		if err != nil {
			return agent.ToolResult{Content: fmt.Sprintf("describe failed: %v", err), IsError: true}, nil
		}
		infoParts = append(infoParts, meta)

		// Try OCR with tesseract if available
		ocrText, ocrErr := tesseractOCR(ctx, args.Path)
		if ocrErr == nil && ocrText != "" {
			infoParts = append(infoParts, "--- OCR Text ---", ocrText)
		} else if ocrErr != nil {
			infoParts = append(infoParts, fmt.Sprintf("(OCR not available: %v)", ocrErr))
		}

	default:
		// Cross-platform: try ImageMagick identify
		meta, err := imagemagickDescribe(ctx, args.Path)
		if err != nil {
			// Fallback: use Go's os.Stat for basic file info
			meta = basicFileInfo(args.Path)
		}
		infoParts = append(infoParts, meta)

		// Try OCR
		ocrText, ocrErr := tesseractOCR(ctx, args.Path)
		if ocrErr == nil && ocrText != "" {
			infoParts = append(infoParts, "--- OCR Text ---", ocrText)
		}
	}

	return agent.ToolResult{Content: strings.Join(infoParts, "\n")}, nil
}

// resize changes the image dimensions.
func (t *ImagingTool) resize(ctx context.Context, args imagingArgs) (agent.ToolResult, error) {
	if args.Width <= 0 && args.Height <= 0 {
		return agent.ValidationError("at least one of width or height must be positive"), nil
	}

	switch runtime.GOOS {
	case "darwin":
		return sipsResize(ctx, args)
	default:
		return imagemagickResize(ctx, args)
	}
}

// convert changes the image format.
func (t *ImagingTool) convert(ctx context.Context, args imagingArgs) (agent.ToolResult, error) {
	if args.Format == "" {
		return agent.ValidationError("format is required for convert action"), nil
	}

	ext := strings.ToLower(strings.TrimLeft(args.Format, "."))

	switch runtime.GOOS {
	case "darwin":
		return sipsConvert(ctx, args.Path, ext)
	default:
		return imagemagickConvert(ctx, args.Path, ext)
	}
}

// --- macOS-specific helpers using sips ---

func sipsDescribe(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "sips", "-g", "all", path)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("sips describe failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func sipsResize(ctx context.Context, args imagingArgs) (agent.ToolResult, error) {
	var sipsArgs []string

	if args.Width > 0 && args.Height > 0 {
		// Both dimensions specified: use -z (height first, width second)
		sipsArgs = []string{"-z", fmt.Sprintf("%d", args.Height), fmt.Sprintf("%d", args.Width), args.Path}
	} else if args.Width > 0 {
		sipsArgs = []string{"--resampleWidth", fmt.Sprintf("%d", args.Width), args.Path}
	} else {
		sipsArgs = []string{"--resampleHeight", fmt.Sprintf("%d", args.Height), args.Path}
	}

	cmd := exec.CommandContext(ctx, "sips", sipsArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("resize failed: %v\n%s", err, string(out)), IsError: true}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Resized %s to %dx%d", filepath.Base(args.Path), args.Width, args.Height)}, nil
}

func sipsConvert(ctx context.Context, path, format string) (agent.ToolResult, error) {
	outPath := changeExtension(path, format)

	// sips -s format <fmt> <path> --out <outpath>
	cmd := exec.CommandContext(ctx, "sips", "-s", "format", format, path, "--out", outPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("convert failed: %v\n%s", err, string(out)), IsError: true}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Converted %s to %s", filepath.Base(path), filepath.Base(outPath))}, nil
}

// --- Cross-platform helpers using ImageMagick ---

func imagemagickDescribe(ctx context.Context, path string) (string, error) {
	if _, err := exec.LookPath("identify"); err != nil {
		return "", fmt.Errorf("ImageMagick identify not found: %w", err)
	}
	cmd := exec.CommandContext(ctx, "identify", "-verbose", path)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("identify failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func imagemagickResize(ctx context.Context, args imagingArgs) (agent.ToolResult, error) {
	if _, err := exec.LookPath("convert"); err != nil {
		return agent.ToolResult{Content: "ImageMagick convert not found. Install ImageMagick or use on macOS.", IsError: true}, nil
	}

	geometry := fmt.Sprintf("%dx%d", args.Width, args.Height)
	if args.Width <= 0 {
		geometry = fmt.Sprintf("x%d", args.Height)
	}
	if args.Height <= 0 {
		geometry = fmt.Sprintf("%d", args.Width)
	}

	cmd := exec.CommandContext(ctx, "convert", args.Path, "-resize", geometry, args.Path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("resize failed: %v\n%s", err, string(out)), IsError: true}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Resized %s to %s", filepath.Base(args.Path), geometry)}, nil
}

func imagemagickConvert(ctx context.Context, path, format string) (agent.ToolResult, error) {
	if _, err := exec.LookPath("convert"); err != nil {
		return agent.ToolResult{Content: "ImageMagick convert not found. Install ImageMagick or use on macOS.", IsError: true}, nil
	}

	outPath := changeExtension(path, format)
	cmd := exec.CommandContext(ctx, "convert", path, outPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("convert failed: %v\n%s", err, string(out)), IsError: true}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Converted %s to %s", filepath.Base(path), filepath.Base(outPath))}, nil
}

// --- tesseract OCR ---

func tesseractOCR(ctx context.Context, path string) (string, error) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		return "", fmt.Errorf("tesseract not installed")
	}

	cmd := exec.CommandContext(ctx, "tesseract", path, "stdout", "-l", "eng", "--psm", "3")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract failed: %w\n%s", err, stderr.String())
	}

	text := strings.TrimSpace(stdout.String())
	return text, nil
}

// --- utility helpers ---

func basicFileInfo(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Sprintf("File: %s", filepath.Base(path))
	}
	return fmt.Sprintf("File: %s\nSize: %d bytes", filepath.Base(path), info.Size())
}

func changeExtension(path, newExt string) string {
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	return base + "." + strings.TrimLeft(newExt, ".")
}
