package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/starclaw/starclaw/internal/agent"
)

// grepVCSDirs lists version control directories to skip when using ripgrep.
var grepVCSDirs = []string{".git", ".svn", ".hg", ".bzr", ".jj", ".sl"}

// GrepTool searches file contents with regex
type GrepTool struct{}

type grepArgs struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
	Glob    string `json:"glob,omitempty"`
}

func (t *GrepTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "grep",
		Description: "Search file contents with regex pattern. Returns matching lines with file names and line numbers. Uses ripgrep (rg) when available for faster searching with glob filtering support.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{"type": "string", "description": "Regex pattern to search for"},
				"path":    map[string]any{"type": "string", "description": "Directory or file to search (default: current directory)"},
				"glob":    map[string]any{"type": "string", "description": "File glob filter (e.g. '*.csv', '*.txt', '*.go'). Only honored with rg; ignored on grep fallback."},
			},
		},
		Required: []string{"pattern"},
	}
}

func (t *GrepTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args grepArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	if args.Path == "" {
		args.Path = "."
	}

	args.Path = ExpandHome(args.Path)

	// Try ripgrep first for better performance and features
	if bin, err := exec.LookPath("rg"); err == nil {
		return t.searchWithRG(ctx, bin, args)
	}

	// Fall back to native Go grep when rg is not available
	return t.searchNative(args)
}

// searchWithRG runs ripgrep with enhanced flags.
func (t *GrepTool) searchWithRG(ctx context.Context, bin string, args grepArgs) (agent.ToolResult, error) {
	var cmdArgs []string

	// Include hidden files in search
	cmdArgs = append(cmdArgs, "--hidden")

	// Skip VCS directories to avoid noise and performance issues.
	// Add both the directory itself and anything inside it.
	for _, dir := range grepVCSDirs {
		cmdArgs = append(cmdArgs, "--glob", "!"+dir, "--glob", "!"+dir+"/**")
	}

	// Prevent long lines (minified/base64) from dominating output
	cmdArgs = append(cmdArgs, "--max-columns", "500")

	// Sort by modification time for better result ordering
	cmdArgs = append(cmdArgs, "--sort", "modified")

	// Show line numbers
	cmdArgs = append(cmdArgs, "-n")

	// Glob filter
	if args.Glob != "" {
		cmdArgs = append(cmdArgs, "--glob", args.Glob)
	}

	cmdArgs = append(cmdArgs, "--", args.Pattern)
	cmdArgs = append(cmdArgs, args.Path)

	cmd := exec.CommandContext(ctx, bin, cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// rg exits with code 1 when no matches found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return agent.ToolResult{Content: "No matches found."}, nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("search error: %v\n%s", err, strings.TrimSpace(string(output))),
			IsError: true,
		}, nil
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return agent.ToolResult{Content: "No matches found."}, nil
	}

	return agent.ToolResult{Content: result}, nil
}

// searchNative uses Go-native file walking and regexp matching when rg is not available.
func (t *GrepTool) searchNative(args grepArgs) (agent.ToolResult, error) {
	// Compile regex
	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid regex pattern: %v", err)), nil
	}

	// Determine if path is a file or directory
	info, err := os.Stat(args.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ValidationError(fmt.Sprintf("path not found: %s", args.Path)), nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("error accessing path: %v", err),
			IsError: true,
		}, nil
	}

	var matches []string

	if info.IsDir() {
		// Search directory
		matches, err = t.searchDirectory(args.Path, re)
	} else {
		// Search single file
		matches, err = t.searchFile(args.Path, re)
	}

	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("search error: %v", err),
			IsError: true,
		}, nil
	}

	if len(matches) == 0 {
		return agent.ToolResult{Content: "No matches found."}, nil
	}

	return agent.ToolResult{
		Content: strings.Join(matches, "\n"),
	}, nil
}

func (t *GrepTool) searchDirectory(dir string, re *regexp.Regexp) ([]string, error) {
	var matches []string

	// Use doublestar to find all files recursively.
	// Root the filesystem at dir so absolute paths work correctly.
	pattern := "**/*"
	files, err := doublestar.Glob(os.DirFS(dir), pattern)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			continue
		}

		// Skip binary files
		if t.isBinaryFile(fullPath) {
			continue
		}

		fileMatches, err := t.searchFile(fullPath, re)
		if err != nil {
			continue
		}
		matches = append(matches, fileMatches...)
	}

	return matches, nil
}

func (t *GrepTool) searchFile(file string, re *regexp.Regexp) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var matches []string
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, fmt.Sprintf("%s:%d: %s", file, lineNum, line))
		}
	}

	return matches, scanner.Err()
}

func (t *GrepTool) isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := []string{".exe", ".dll", ".so", ".dylib", ".bin", ".o", ".a", ".png", ".jpg", ".gif", ".mp3", ".mp4", ".zip", ".tar", ".gz"}
	for _, b := range binaryExts {
		if ext == b {
			return true
		}
	}
	return false
}

func (t *GrepTool) RequiresApproval() bool { return false }

func (t *GrepTool) IsSafeArgs(argsJSON string) bool {
	return true // Grep is read-only
}
