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
	Run: func(cmd *cobra.Command, args []string) {
		silent, _ := cmd.Flags().GetBool("silent")
		SetupCompletion(silent, true) // Explicit run forces regeneration
	},
}

func init() {
	setupCompletionCmd.Flags().Bool("silent", false, "Run setup without printing progress")
	rootCmd.AddCommand(setupCompletionCmd)
}

// SetupCompletion ensures that the shell completion is installed in the user's RC file.
func SetupCompletion(silent bool, force bool) {
	shell := os.Getenv("SHELL")
	home, _ := os.UserHomeDir()

	// 1. Determine shell-specific paths
	var shellName string
	var rcFile string
	var staticFile string
	var completionCmd string

	if strings.Contains(shell, "zsh") {
		shellName = "zsh"
		rcFile = filepath.Join(home, ".zshrc")
		staticFile = filepath.Join(home, ".minhthetus-cli", "completion.zsh")
		completionCmd = fmt.Sprintf("[[ -f %s ]] && source %s", staticFile, staticFile)
	} else if strings.Contains(shell, "bash") {
		shellName = "bash"
		rcFile = filepath.Join(home, ".bashrc")
		staticFile = filepath.Join(home, ".minhthetus-cli", "completion.bash")
		completionCmd = fmt.Sprintf("[[ -f %s ]] && source %s", staticFile, staticFile)
	} else {
		if !silent {
			fmt.Printf("Unsupported shell: %s. Please manually setup completion.\n", shell)
		}
		return
	}

	// 2. Check if the static completion file already exists and is non-empty.
	staticFileExists := false
	if info, err := os.Stat(staticFile); err == nil && info.Size() > 0 {
		staticFileExists = true
	}

	// 3. Write completion script to the static file if it doesn't exist or if forced
	if !staticFileExists || force {
		if !silent {
			fmt.Printf("⏳ Generating static completion file for %s...\n", shellName)
		}
		dir := filepath.Dir(staticFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			if !silent {
				fmt.Printf("❌ Error creating directory %s: %v\n", dir, err)
			}
			return
		}

		f, err := os.OpenFile(staticFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			if !silent {
				fmt.Printf("❌ Error creating file %s: %v\n", staticFile, err)
			}
			return
		}

		// Generate completion directly from rootCmd
		if shellName == "zsh" {
			err = rootCmd.GenZshCompletion(f)
		} else {
			err = rootCmd.GenBashCompletion(f)
		}
		f.Close()

		if err != nil {
			if !silent {
				fmt.Printf("❌ Error generating completion script: %v\n", err)
			}
			return
		}
	}

	// 4. Read RC file to check if already integrated and clean old integrations
	content, err := os.ReadFile(rcFile)
	if err != nil {
		// If RC file doesn't exist, try to create it
		if !silent {
			fmt.Printf("⏳ Creating shell configuration file %s...\n", rcFile)
		}
		if err := os.WriteFile(rcFile, []byte("\n# minhthetus-cli completion\n"+completionCmd+"\n"), 0644); err != nil {
			if !silent {
				fmt.Printf("❌ Error creating %s: %v\n", rcFile, err)
			}
			return
		}
		if !silent {
			fmt.Printf("✅ Successfully set up autocomplete for %s!\n", shellName)
			fmt.Printf("👉 Please run 'source %s' to activate it in the current session.\n", rcFile)
		}
		return
	}

	rcContent := string(content)
	hasOldIntegration := strings.Contains(rcContent, "minhthetus-cli completion")

	// If old dynamic source integration exists, clean it up!
	if hasOldIntegration && (strings.Contains(rcContent, "completion zsh") || strings.Contains(rcContent, "completion bash")) {
		if !silent {
			fmt.Println("⏳ Updating completion integration in your shell configuration...")
		}
		lines := strings.Split(rcContent, "\n")
		var newLines []string
		skipMode := false
		for _, line := range lines {
			// Skip the comments and command block of the old autocomplete
			if strings.Contains(line, "# minhthetus-cli completion") {
				skipMode = true
				continue
			}
			if skipMode {
				if strings.Contains(line, "minhthetus-cli completion") || strings.Contains(line, "source <(") {
					continue
				}
				skipMode = false
			}
			newLines = append(newLines, line)
		}
		
		// Append the new clean static source line
		rcContent = strings.TrimSpace(strings.Join(newLines, "\n")) + "\n\n# minhthetus-cli completion\n" + completionCmd + "\n"
		_ = os.WriteFile(rcFile, []byte(rcContent), 0644)
		
		if !silent {
			fmt.Printf("✅ Successfully updated autocomplete integration in %s!\n", rcFile)
			fmt.Printf("👉 Please run 'source %s' to activate it in the current session.\n", rcFile)
		}
		return
	}

	// If no integration exists at all, append it
	if !strings.Contains(rcContent, "minhthetus-cli completion") {
		if !silent {
			fmt.Printf("⏳ Integrating completion in %s...\n", rcFile)
		}
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
			fmt.Printf("✅ Successfully set up autocomplete for %s!\n", shellName)
			fmt.Printf("👉 Please run 'source %s' to activate it in the current session.\n", rcFile)
		}
	}
}
