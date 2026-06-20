package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// InstallMethod represents the way the CLI was installed
type InstallMethod string

const (
	Homebrew    InstallMethod = "Homebrew"
	GoInstall   InstallMethod = "Go"
	ManualBuild InstallMethod = "Manual/Makefile"
)

// detectInstallMethod inspects the executable path to determine the installation type
func detectInstallMethod(path string) InstallMethod {
	if strings.Contains(path, "Cellar") || strings.Contains(path, "homebrew") || strings.Contains(path, "opt/homebrew") {
		return Homebrew
	}
	if strings.Contains(path, "go/bin") {
		return GoInstall
	}
	return ManualBuild
}

// runBrewUninstall triggers Homebrew's uninstallation database command
func runBrewUninstall() error {
	brewCmd := exec.Command("brew", "uninstall", "minhthetus-cli")
	brewCmd.Stdout = os.Stdout
	brewCmd.Stderr = os.Stderr
	brewCmd.Stdin = os.Stdin
	return brewCmd.Run()
}

// cleanupShellConfig automatically removes all autocomplete integration lines from zshrc and bashrc
func cleanupShellConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	rcFiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
	}

	for _, rcPath := range rcFiles {
		if _, err := os.Stat(rcPath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(rcPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		var newLines []string
		removed := false

		for _, line := range lines {
			// Skip any lines containing "minhthetus-cli" to remove autocomplete hooks
			if strings.Contains(line, "minhthetus-cli") {
				removed = true
				continue
			}
			newLines = append(newLines, line)
		}

		if removed {
			// Write the cleaned shell config back
			err = os.WriteFile(rcPath, []byte(strings.Join(newLines, "\n")), 0644)
			if err == nil {
				fmt.Printf("   - Cleaned up shell completion references in '%s'\n", filepath.Base(rcPath))
			}
		}
	}
}

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely remove the CLI and all integrations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⏳ Preparing uninstallation...")

		// 1. Detect executable path
		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("❌ Failed to detect executable path: %v\n", err)
			return
		}

		// 2. Ask user for confirmation
		confirmMsg := fmt.Sprintf("Are you sure you want to completely uninstall minhthetus-cli (running from '%s')?", exePath)
		confirmed, err := ui.Confirm(confirmMsg, 0, false)
		if err != nil || !confirmed {
			fmt.Println("❌ Uninstallation cancelled.")
			return
		}

		// 3. Perform custom cleanup (Removing logs, caches, autocomplete)
		fmt.Println("🧹 Performing cleanup...")
		fmt.Println("   - Removing temporary caches...")
		cleanupShellConfig()

		// Delete static completions directory
		home, _ := os.UserHomeDir()
		compDir := filepath.Join(home, ".minhthetus-cli")
		if _, err := os.Stat(compDir); err == nil {
			fmt.Println("   - Removing completion files...")
			_ = os.RemoveAll(compDir)
		}

		// 4. Determine installation method and proceed
		method := detectInstallMethod(exePath)
		fmt.Printf("⏳ Detected %s installation.\n", method)

		switch method {
		case Homebrew:
			fmt.Println("⏳ Triggering 'brew uninstall'...")
			if err := runBrewUninstall(); err != nil {
				fmt.Printf("❌ Homebrew uninstallation failed: %v\n", err)
				fmt.Println("👉 Please run the command manually: brew uninstall minhthetus-cli")
				return
			}

		case GoInstall, ManualBuild:
			fmt.Printf("⏳ Deleting binary file at '%s'...\n", exePath)
			if err := os.Remove(exePath); err != nil {
				fmt.Printf("❌ Permission denied or failed to delete binary.\n")
				fmt.Println("👉 Please run the command manually:")
				fmt.Printf("   sudo rm %s\n", exePath)
				return
			}
		}

		fmt.Println("✅ Successfully uninstalled minhthetus-cli! Goodbye!")
		fmt.Println("\n👉 To apply all changes and refresh your active shell, please run:")
		fmt.Println("   source ~/.zshrc")
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
