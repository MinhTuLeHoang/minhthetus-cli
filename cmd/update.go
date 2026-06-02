package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the CLI to the latest version",
	Long:  `Detects the installation method (Homebrew, Go, or Manual) and updates the minhthetus-cli binary to the latest version while safely preserving all local configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ℹ️  All local configurations in ~/.minhthetus-cli/ will be fully preserved.")
		fmt.Println("⏳ Checking current installation method...")

		// 1. Detect executable path
		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("❌ Failed to detect executable path: %v\n", err)
			return
		}

		method := detectInstallMethod(exePath)
		fmt.Printf("⏳ Detected %s installation.\n", method)

		// 2. Perform update based on method
		switch method {
		case Homebrew:
			confirmMsg := "Would you like to run 'brew upgrade minhthetus-cli' to update?"
			confirmed, err := ui.Confirm(confirmMsg, 0, true)
			if err != nil || !confirmed {
				fmt.Println("❌ Update cancelled.")
				return
			}

			fmt.Println("⏳ Running 'brew upgrade minhthetus-cli'...")
			brewCmd := exec.Command("brew", "upgrade", "minhthetus-cli")
			brewCmd.Stdout = os.Stdout
			brewCmd.Stderr = os.Stderr
			brewCmd.Stdin = os.Stdin
			if err := brewCmd.Run(); err != nil {
				fmt.Printf("❌ Homebrew update failed: %v\n", err)
				fmt.Println("👉 Please run the command manually: brew upgrade minhthetus-cli")
				return
			}
			fmt.Println("✅ Successfully updated minhthetus-cli via Homebrew!")

		case GoInstall:
			confirmMsg := "Would you like to run 'go install' to update to the latest version?"
			confirmed, err := ui.Confirm(confirmMsg, 0, true)
			if err != nil || !confirmed {
				fmt.Println("❌ Update cancelled.")
				return
			}

			fmt.Println("⏳ Running 'go install github.com/MinhTuLeHoang/minhthetus-cli@latest'...")
			goCmd := exec.Command("go", "install", "github.com/MinhTuLeHoang/minhthetus-cli@latest")
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr
			goCmd.Stdin = os.Stdin
			if err := goCmd.Run(); err != nil {
				fmt.Printf("❌ Go install update failed: %v\n", err)
				fmt.Println("👉 Please run the command manually: go install github.com/MinhTuLeHoang/minhthetus-cli@latest")
				return
			}
			fmt.Println("✅ Successfully updated minhthetus-cli via Go!")

		case ManualBuild:
			// Check if we are running in the cloned directory by checking for Makefile
			if _, err := os.Stat("Makefile"); err == nil {
				confirmMsg := "Manual/Makefile installation detected. Do you want to pull the latest changes and reinstall now?"
				confirmed, err := ui.Confirm(confirmMsg, 0, true)
				if err == nil && confirmed {
					fmt.Println("⏳ Pulling latest changes (git pull)...")
					gitCmd := exec.Command("git", "pull")
					gitCmd.Stdout = os.Stdout
					gitCmd.Stderr = os.Stderr
					gitCmd.Stdin = os.Stdin
					if err := gitCmd.Run(); err != nil {
						fmt.Printf("❌ Failed to git pull: %v\n", err)
						return
					}

					fmt.Println("⏳ Recompiling and reinstalling (make install)...")
					makeCmd := exec.Command("make", "install")
					makeCmd.Stdout = os.Stdout
					makeCmd.Stderr = os.Stderr
					makeCmd.Stdin = os.Stdin
					if err := makeCmd.Run(); err != nil {
						fmt.Printf("❌ Failed to run make install: %v\n", err)
						return
					}
					fmt.Println("✅ Successfully updated minhthetus-cli manually!")
					return
				}
			}

			fmt.Println("👉 Manual/Makefile installation detected. To update, please navigate to your cloned minhthetus-cli repository and run:")
			fmt.Println("   git pull && make install")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
