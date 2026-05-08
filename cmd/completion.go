package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for bash, zsh, or fish.

To use the completions in the current shell:

  source <(starclaw completion bash)     # bash
  source <(starclaw completion zsh)      # zsh
  starclaw completion fish | source      # fish

For permanent installation:

  starclaw completion install

This will auto-detect your shell and write the completion script
to the appropriate location.`,
	DisableFlagsInUseLine: true,
}

var completionBashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completion script",
	Long:  "Generate the bash completion script for StarClaw.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Root().GenBashCompletionV2(os.Stdout, true)
	},
}

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh completion script",
	Long:  "Generate the zsh completion script for StarClaw.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Root().GenZshCompletion(os.Stdout)
	},
}

var completionFishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generate fish completion script",
	Long:  "Generate the fish completion script for StarClaw.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install shell completion automatically",
	Long: `Detect the current shell and permanently install auto-completion
for StarClaw.

Supports bash, zsh, and fish shells. The completion script will be
written to the appropriate location for the detected shell.

After installation you may need to restart your shell or source
your shell's configuration file for the changes to take effect.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installCompletion(cmd)
	},
}

func installCompletion(cmd *cobra.Command) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return fmt.Errorf("could not detect shell ($SHELL is empty)")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	switch {
	case strings.Contains(shell, "bash"):
		return installBashCompletion(cmd, homeDir)
	case strings.Contains(shell, "zsh"):
		return installZshCompletion(cmd, homeDir)
	case strings.Contains(shell, "fish"):
		return installFishCompletion(cmd, homeDir)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
}

func installBashCompletion(cmd *cobra.Command, homeDir string) error {
	completionDir := filepath.Join(homeDir, ".bash_completion.d")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("could not create completion directory: %w", err)
	}

	scriptPath := filepath.Join(completionDir, "starclaw")
	f, err := os.Create(scriptPath)
	if err != nil {
		return fmt.Errorf("could not create completion file: %w", err)
	}
	defer f.Close()

	if err := cmd.Root().GenBashCompletionV2(f, true); err != nil {
		return fmt.Errorf("could not generate bash completion: %w", err)
	}

	fmt.Printf("Bash completion installed to: %s\n", scriptPath)
	fmt.Println()
	fmt.Println("Add the following line to your ~/.bashrc and restart your shell:")
	fmt.Printf("  source %s\n", scriptPath)
	return nil
}

func installZshCompletion(cmd *cobra.Command, homeDir string) error {
	completionDir := filepath.Join(homeDir, ".zsh", "completions")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("could not create completion directory: %w", err)
	}

	scriptPath := filepath.Join(completionDir, "_starclaw")
	f, err := os.Create(scriptPath)
	if err != nil {
		return fmt.Errorf("could not create completion file: %w", err)
	}
	defer f.Close()

	if err := cmd.Root().GenZshCompletion(f); err != nil {
		return fmt.Errorf("could not generate zsh completion: %w", err)
	}

	fmt.Printf("Zsh completion installed to: %s\n", scriptPath)
	fmt.Println()
	fmt.Println("Add the following to your ~/.zshrc before compinit:")
	fmt.Printf("  fpath=(%s $fpath)\n", completionDir)
	fmt.Println("  autoload -Uz compinit && compinit")
	return nil
}

func installFishCompletion(cmd *cobra.Command, homeDir string) error {
	completionDir := filepath.Join(homeDir, ".config", "fish", "completions")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("could not create completion directory: %w", err)
	}

	scriptPath := filepath.Join(completionDir, "starclaw.fish")
	f, err := os.Create(scriptPath)
	if err != nil {
		return fmt.Errorf("could not create completion file: %w", err)
	}
	defer f.Close()

	if err := cmd.Root().GenFishCompletion(f, true); err != nil {
		return fmt.Errorf("could not generate fish completion: %w", err)
	}

	fmt.Printf("Fish completion installed to: %s\n", scriptPath)
	fmt.Println("Fish completions are automatically loaded from this directory.")
	fmt.Println("Restart your shell or run: source " + scriptPath)
	return nil
}

func init() {
	completionCmd.AddCommand(completionBashCmd)
	completionCmd.AddCommand(completionZshCmd)
	completionCmd.AddCommand(completionFishCmd)
	completionCmd.AddCommand(completionInstallCmd)
}
