package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// setupCompletionCmd represents the setup-completion command
var setupCompletionCmd = &cobra.Command{
	Use:   "setup-completion",
	Short: "Install tab completion for your shell",
	Args:  cobra.NoArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"--silent", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		silent, _ := cmd.Flags().GetBool("silent")
		SetupCompletion(silent)
	},
}

func init() {
	setupCompletionCmd.Flags().Bool("silent", false, "Run setup without printing progress")
	rootCmd.AddCommand(setupCompletionCmd)
}

// SetupCompletion ensures that the shell completion is installed in the user's RC file.
func SetupCompletion(silent bool) {
	shell := os.Getenv("SHELL")
	home, _ := os.UserHomeDir()

	var rcFile string
	var completionCmd string

	if strings.Contains(shell, "zsh") {
		rcFile = filepath.Join(home, ".zshrc")
		completionCmd = "source <(minhthetus-cli completion zsh)"
	} else if strings.Contains(shell, "bash") {
		rcFile = filepath.Join(home, ".bashrc")
		completionCmd = "source <(minhthetus-cli completion bash)"
	} else {
		if !silent {
			fmt.Printf("Unsupported shell: %s. Please manually setup completion.\n", shell)
		}
		return
	}

	// Check if already installed
	content, _ := os.ReadFile(rcFile)
	if strings.Contains(string(content), "minhthetus-cli completion") {
		if !silent {
			fmt.Printf("✅ Completion already exists in %s\n", rcFile)
		}
		return
	}

	if !silent {
		fmt.Printf("⏳ Setting up completion for %s...\n", shell)
	}

	// Append to rc file
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		if !silent {
			fmt.Printf("❌ Error opening %s: %v\n", rcFile, err)
		}
		return
	}
	defer f.Close()

	if _, err := f.WriteString("\n# minhthetus-cli completion\n" + completionCmd + "\n"); err != nil {
		if !silent {
			fmt.Printf("❌ Error writing to %s: %v\n", rcFile, err)
		}
		return
	}

	if !silent {
		fmt.Printf("✅ Successfully added completion to %s\n", rcFile)
		fmt.Printf("👉 Please run 'source %s' to activate it in the current session.\n", rcFile)
	}
}
