package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
		if !ui.GumConfirm(confirmMsg) {
			fmt.Println("❌ Uninstallation cancelled.")
			return
		}

		// 3. Perform custom cleanup (Placeholders for logs, caches, autocomplete)
		fmt.Println("🧹 Performing cleanup...")
		fmt.Println("   - Removing temporary caches...")

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
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
