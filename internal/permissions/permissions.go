// Package permissions provides tool-level security for the agent loop.
// It evaluates tool calls against configurable rules and returns
// a decision: "allow" (proceed silently), "deny" (block), or "ask" (prompt user).
package permissions

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Decision represents a permission check result.
type Decision string

const (
	Allow Decision = "allow"
	Deny  Decision = "deny"
	Ask   Decision = "ask"
)

// Config defines user-configurable permission rules.
type Config struct {
	AllowedDirs       []string `yaml:"allowed_dirs"        json:"allowed_dirs"`
	AllowedCommands   []string `yaml:"allowed_commands"    json:"allowed_commands"`
	DeniedCommands    []string `yaml:"denied_commands"     json:"denied_commands"`
	SensitivePatterns []string `yaml:"sensitive_patterns"  json:"sensitive_patterns"`
	NetworkAllowlist  []string `yaml:"network_allowlist"   json:"network_allowlist"`
}

// hardBlockPatterns are always denied regardless of config.
var hardBlockPatterns = []string{
	"rm -rf /",
	"rm -rf ~",
	"rm -rf /System",
	"rm -rf /Users",
	"rm -rf /*",
	"> /dev/sd*",
	"> /dev/disk*",
	"mkfs.*",
	"dd if=* of=/dev/*",
	"curl * | sh",
	"curl * | bash",
	"wget * | sh",
	"wget * | bash",
}

// defaultSensitivePatterns are file patterns considered sensitive by default.
var defaultSensitivePatterns = []string{
	".env",
	".env.*",
	"*.pem",
	"*.key",
	"id_rsa*",
	"id_ed25519*",
	".ssh/config",
	"*.keychain*",
	"tokens.json",
	"credentials.json",
	"*.secrets",
}

// defaultSafeCommands are read-only commands allowed without explicit config.
var defaultSafeCommands = []safeEntry{
	// System info & file inspection
	{prefix: "ls"}, {exact: "pwd"}, {prefix: "which"}, {prefix: "whereis"},
	{prefix: "echo"}, {prefix: "cat"}, {prefix: "head"}, {prefix: "tail"},
	{prefix: "wc"}, {prefix: "file"}, {prefix: "stat"}, {prefix: "du"}, {prefix: "df"},
	{exact: "id"}, {exact: "whoami"}, {exact: "hostname"}, {prefix: "uname"},
	{exact: "uptime"}, {prefix: "date"}, {prefix: "cal"},
	{exact: "env"}, {exact: "printenv"},
	{prefix: "basename"}, {prefix: "dirname"}, {prefix: "realpath"}, {prefix: "readlink"},
	{exact: "true"}, {exact: "false"}, {prefix: "seq"}, {exact: "nproc"}, {exact: "arch"}, {exact: "tty"},

	// Checksums
	{prefix: "md5"}, {prefix: "shasum"}, {prefix: "sha256sum"}, {prefix: "cksum"},

	// Text processing
	{prefix: "grep"}, {prefix: "rg"}, {prefix: "sort"}, {prefix: "uniq"},
	{prefix: "tr"}, {prefix: "cut"}, {prefix: "diff"}, {prefix: "cmp"},
	{prefix: "strings"}, {prefix: "od"}, {prefix: "hexdump"}, {prefix: "base64"},
	{prefix: "jq"}, {prefix: "yq"},

	// File finding (read-only)
	{prefix: "fd"}, {prefix: "locate"}, {prefix: "mdfind"}, {prefix: "tree"},

	// Process info
	{prefix: "ps"}, {prefix: "top -l"}, {prefix: "pgrep"}, {prefix: "lsof"},

	// Network diagnostics
	{prefix: "ping"}, {prefix: "traceroute"}, {prefix: "dig"}, {prefix: "nslookup"},
	{prefix: "host"}, {prefix: "ifconfig"}, {prefix: "netstat"}, {prefix: "ss"},

	// Help / version
	{prefix: "man"}, {prefix: "info"}, {prefix: "help"},
	{prefix: "sw_vers"}, {prefix: "sysctl"},

	// Go
	{prefix: "go version"}, {prefix: "go env"}, {prefix: "go doc"}, {prefix: "go list"},
	{prefix: "go build"}, {prefix: "go test"}, {prefix: "go vet"},
	{prefix: "go mod download"}, {prefix: "go mod verify"}, {prefix: "go mod why"},

	// Git (read-only)
	{prefix: "git status"}, {prefix: "git diff"}, {prefix: "git log"}, {prefix: "git show"},
	{prefix: "git branch"}, {prefix: "git tag"}, {prefix: "git remote"},
	{prefix: "git ls-files"}, {prefix: "git rev-parse"}, {prefix: "git describe"},
	{prefix: "git blame"},

	// Linters
	{prefix: "golangci-lint"}, {prefix: "staticcheck"}, {prefix: "shellcheck"},
	{prefix: "mypy"}, {prefix: "ruff check"},

	// Build tools
	{prefix: "make"}, {prefix: "npm test"}, {prefix: "npm run"},
	{prefix: "cargo build"}, {prefix: "cargo test"}, {prefix: "cargo check"},
	{prefix: "python -m pytest"}, {prefix: "python3 -m pytest"}, {prefix: "pytest"},

	// CLI tools (read-only)
	{prefix: "gh pr list"}, {prefix: "gh pr view"}, {prefix: "gh pr status"},
	{prefix: "gh issue list"}, {prefix: "gh issue view"},
	{prefix: "gh repo view"}, {prefix: "gh run list"}, {prefix: "gh run view"},
	{prefix: "gh auth status"}, {prefix: "gh status"},

	// Docker (read-only)
	{prefix: "docker ps"}, {prefix: "docker images"}, {prefix: "docker inspect"},
	{prefix: "docker logs"}, {prefix: "docker stats"}, {prefix: "docker version"},
	{prefix: "docker info"}, {prefix: "docker compose ps"}, {prefix: "docker compose logs"},
}

type safeEntry struct {
	exact  string
	prefix string
}

// DefaultConfig returns a Config with safe defaults and no restrictions.
func DefaultConfig() *Config {
	return &Config{
		AllowedDirs: []string{"~", "."},
	}
}

// ──────────────────────────────────────────────
// Public API
// ──────────────────────────────────────────────

// CheckToolCall evaluates a tool call against permission rules.
// Returns the decision and a human-readable reason.
// An empty decision means the tool type is not handled by this engine.
func CheckToolCall(toolName, argsJSON string, cfg *Config) (Decision, string) {
	switch toolName {
	case "bash":
		cmd := extractField(argsJSON, "command")
		return checkCommand(cmd, cfg)
	case "file_read":
		return checkFilePath(extractField(argsJSON, "path"), "read", cfg)
	case "document_text", "archive_inspect":
		return checkFilePath(extractField(argsJSON, "path"), "read", cfg)
	case "archive_extract":
		pathDecision, pathReason := checkFilePath(extractField(argsJSON, "path"), "read", cfg)
		if pathDecision == Deny {
			return pathDecision, pathReason
		}
		destDecision, destReason := checkFilePath(extractField(argsJSON, "destination"), "write", cfg)
		if destDecision == Deny {
			return destDecision, destReason
		}
		return Ask, "archive extraction writes files and requires approval"
	case "file_write", "file_edit":
		return checkFilePath(extractField(argsJSON, "path"), "write", cfg)
	case "glob", "grep":
		path := extractField(argsJSON, "path")
		if path == "" {
			path = extractField(argsJSON, "pattern")
		}
		return checkFilePath(path, "read", cfg)
	case "directory_list":
		path := extractField(argsJSON, "path")
		if path == "" {
			path = "."
		}
		return checkFilePath(path, "read", cfg)
	case "http":
		return checkNetwork(extractField(argsJSON, "url"), cfg)
	}
	return "", ""
}

// IsHardBlocked returns true if a bash command matches an always-blocked pattern.
func IsHardBlocked(cmd string) bool {
	for _, p := range hardBlockPatterns {
		if matchGlob(strings.TrimSpace(cmd), p) {
			return true
		}
	}
	return false
}

// IsSensitiveFile returns true if a filename matches sensitive file patterns.
func IsSensitiveFile(filename string) bool {
	for _, p := range defaultSensitivePatterns {
		if matchGlob(filename, p) {
			return true
		}
	}
	return false
}

// ──────────────────────────────────────────────
// Command checking
// ──────────────────────────────────────────────

func checkCommand(cmd string, cfg *Config) (Decision, string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return Deny, "empty command"
	}

	// 1. Hard-block patterns
	for _, p := range hardBlockPatterns {
		if matchGlob(cmd, p) {
			return Deny, "matches hard-block pattern: " + p
		}
	}

	if cfg == nil {
		return Ask, "no permission config; requires approval"
	}

	// 2. User-denied patterns
	for _, p := range cfg.DeniedCommands {
		if matchGlob(cmd, p) {
			return Deny, "matches denied command: " + p
		}
	}

	// 3. Split compound commands (&&, ||, ;, |)
	subs := splitCompound(cmd)
	if len(subs) > 1 {
		for _, sub := range subs {
			d, r := checkSingleCommand(sub, cfg)
			if d == Deny {
				return Deny, "sub-command denied: " + r
			}
		}
		allAllowed := true
		for _, sub := range subs {
			d, _ := checkSingleCommand(sub, cfg)
			if d != Allow {
				allAllowed = false
				break
			}
		}
		if allAllowed {
			return Allow, "all sub-commands allowed"
		}
		return Ask, "compound command requires approval"
	}

	return checkSingleCommand(cmd, cfg)
}

func checkSingleCommand(cmd string, cfg *Config) (Decision, string) {
	cmd = strings.TrimSpace(cmd)

	for _, p := range hardBlockPatterns {
		if matchGlob(cmd, p) {
			return Deny, "matches hard-block pattern: " + p
		}
	}
	for _, p := range cfg.DeniedCommands {
		if matchGlob(cmd, p) {
			return Deny, "matches denied command: " + p
		}
	}

	// User allowed patterns have highest allow priority
	for _, p := range cfg.AllowedCommands {
		if matchGlob(cmd, p) {
			return Allow, "matches allowed command: " + p
		}
	}

	// Built-in safe defaults
	if isDefaultSafe(cmd) {
		return Allow, "built-in safe command"
	}

	return Ask, "requires approval"
}

func isDefaultSafe(cmd string) bool {
	for _, e := range defaultSafeCommands {
		if e.exact != "" && cmd == e.exact {
			return true
		}
		if e.prefix != "" && (cmd == e.prefix || strings.HasPrefix(cmd, e.prefix+" ")) {
			return true
		}
	}
	return false
}

func splitCompound(cmd string) []string {
	const sep = "\x00"
	for _, op := range []string{"&&", "||", ";", "|"} {
		cmd = strings.ReplaceAll(cmd, op, sep)
	}
	var out []string
	for _, s := range strings.Split(cmd, sep) {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ──────────────────────────────────────────────
// File path checking
// ──────────────────────────────────────────────

func checkFilePath(path, action string, cfg *Config) (Decision, string) {
	if path == "" {
		return Deny, "empty path"
	}

	expanded := expandHome(path)
	realPath, err := filepath.EvalSymlinks(expanded)
	if err != nil {
		realPath = filepath.Clean(expanded)
	}

	if IsSensitiveFile(filepath.Base(realPath)) {
		return Ask, "sensitive file: " + filepath.Base(realPath)
	}

	if cfg == nil {
		return Ask, "no permission config"
	}

	// Check if path is within allowed directories
	inAllowed := false
	for _, dir := range cfg.AllowedDirs {
		absDir, err := filepath.Abs(expandHome(dir))
		if err != nil {
			continue
		}
		absPath, err := filepath.Abs(realPath)
		if err != nil {
			continue
		}
		if isSubPath(absPath, absDir) {
			inAllowed = true
			break
		}
	}

	if action == "write" {
		return Ask, "write operations require approval"
	}
	if inAllowed {
		return Allow, "path within allowed directory"
	}
	return Ask, "path not in allowed directories"
}

// ──────────────────────────────────────────────
// Network checking
// ──────────────────────────────────────────────

func checkNetwork(rawURL string, cfg *Config) (Decision, string) {
	if rawURL == "" {
		return Deny, "empty URL"
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return Deny, "malformed URL: " + err.Error()
	}

	host := parsed.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return Allow, "localhost always allowed"
	}

	if cfg == nil {
		return Ask, "no permission config"
	}

	for _, a := range cfg.NetworkAllowlist {
		if host == a {
			return Allow, "host in network allowlist"
		}
		if strings.HasPrefix(a, "*.") && strings.HasSuffix(host, a[1:]) {
			return Allow, "host matches wildcard: " + a
		}
	}
	return Ask, "host not in network allowlist"
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func extractField(argsJSON, field string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &m); err != nil {
		return ""
	}
	if v, ok := m[field]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		return filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return path
}

func isSubPath(path, dir string) bool {
	path = filepath.Clean(path)
	dir = filepath.Clean(dir)
	if path == dir {
		return true
	}
	return strings.HasPrefix(path, dir+string(filepath.Separator))
}

func matchGlob(s, pattern string) bool {
	si, pi := 0, 0
	starIdx, matchIdx := -1, 0

	for si < len(s) {
		if pi < len(pattern) && (pattern[pi] == '?' || pattern[pi] == s[si]) {
			si++
			pi++
		} else if pi < len(pattern) && pattern[pi] == '*' {
			starIdx = pi
			matchIdx = si
			pi++
		} else if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			si = matchIdx
		} else {
			return false
		}
	}
	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}
	return pi == len(pattern)
}
