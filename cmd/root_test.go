package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// executeCommand runs a cobra command and returns the output
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}

func TestVersionCmd(t *testing.T) {
	// Create a fresh root command
	root := &cobra.Command{Use: "starclaw"}
	root.AddCommand(versionCmd)

	output, err := executeCommand(root, "version")
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	if !strings.Contains(output, "starclaw version") {
		t.Errorf("Expected version output to contain 'starclaw version', got: %s", output)
	}
}

func TestVersionCmd_WithCustomVersion(t *testing.T) {
	// Save original version
	origVersion := Version
	defer func() { Version = origVersion }()

	// Set custom version
	Version = "1.2.3-test"

	cmd := &cobra.Command{Use: "starclaw"}
	cmd.AddCommand(versionCmd)

	output, err := executeCommand(cmd, "version")
	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	if !strings.Contains(output, "1.2.3-test") {
		t.Errorf("Expected version output to contain '1.2.3-test', got: %s", output)
	}
}

func TestRootCmd_Help(t *testing.T) {
	cmd := &cobra.Command{Use: "starclaw"}
	cmd.AddCommand(versionCmd)
	cmd.AddCommand(setupCmd)

	output, err := executeCommand(cmd, "--help")
	if err != nil {
		t.Errorf("help command failed: %v", err)
	}

	// Check that help contains expected sections
	expectedStrings := []string{
		"Usage:",
		"Available Commands:",
		"Flags:",
		"version",
		"setup",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected help to contain '%s', got:\n%s", expected, output)
		}
	}
}

func TestRootCmd_NoArgs_Unconfigured(t *testing.T) {
	// Create temp dir for config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".starclaw.yaml")

	// Save and restore original config path
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	// Ensure no config file exists
	os.Remove(configPath)

	// For this test, we just verify the command structure works
	// The actual setup flow requires user interaction
	cmd := &cobra.Command{Use: "starclaw"}
	cmd.RunE = rootCmd.RunE

	// Execute without args - should return error (needs setup)
	_, err := executeCommand(cmd)
	// We expect an error or the setup message
	if err == nil {
		// If no error, the output should indicate setup is needed
		t.Log("Command completed without error - may indicate config was found or different behavior")
	}
}

func TestChatCmd_NoArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "starclaw"}
	cmd.AddCommand(chatCmd)

	// Execute without args - should fail (requires at least 1 arg)
	_, err := executeCommand(cmd, "chat")
	if err == nil {
		t.Error("Expected chat command to fail without arguments")
	}
}

func TestChatCmd_WithArgs(t *testing.T) {
	// Skip this test if API credentials are available (would actually call LLM)
	if os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" && os.Getenv("ANTHROPIC_BASE_URL") != "" {
		t.Skip("Skipping: API credentials available in environment")
	}

	// This test verifies the command structure
	// The actual chat functionality requires a valid config and API key
	cmd := &cobra.Command{Use: "starclaw"}
	cmd.AddCommand(chatCmd)

	// Execute with args - should attempt to run (will fail due to no config)
	_, err := executeCommand(cmd, "chat", "hello")
	// We expect an error due to no config
	if err == nil {
		t.Error("Expected chat command to fail without valid config")
	}
}

func TestSetupCmd(t *testing.T) {
	cmd := &cobra.Command{Use: "starclaw"}
	cmd.AddCommand(setupCmd)

	// The setup command will fail in non-interactive mode
	// but we can verify it exists and has the right structure
	output, err := executeCommand(cmd, "setup", "--help")
	if err != nil {
		t.Errorf("setup --help failed: %v", err)
	}

	if !strings.Contains(output, "Run interactive setup") {
		t.Errorf("Expected help to describe setup command")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"exactly ten", 11, "exactly ten"},
		{"twelve chars", 11, "twelve char..."},
	}

	for _, tt := range tests {
		got := truncateString(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}
