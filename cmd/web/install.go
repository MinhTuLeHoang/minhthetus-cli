package web

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	forceInstall bool
	ciMode       bool
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs project dependencies with automatic Node.js version switching and package manager detection.",
	Long:  "Installs project dependencies with automatic Node.js version switching and package manager detection.",
	Example: `minhthetus-cli web install --force
minhthetus-cli web install --ci`,
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Web Project Installer",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-f", "--force", "--ci", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		trackCurrentRepo()

		fmt.Printf("%s Detecting environment...\n\n", ui.InfoMessage(""))
		pkgManager, nodeBinDir := getWebInfo()

		if pkgManager == "" {
			fmt.Printf("\n%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Failed to detect package manager."))
			os.Exit(1)
		}

		if forceInstall {
			fmt.Printf("\n%s %s\n", ui.WarningIcon, ui.YellowStyle().Render("Force mode enabled. Cleaning up..."))
			if _, err := os.Stat("node_modules"); err == nil {
				fmt.Printf("  %s Removing node_modules...\n", ui.InfoIcon)
				os.RemoveAll("node_modules")
			}
			switch pkgManager {
			case "pnpm":
				os.Remove("pnpm-lock.yaml")
			case "npm":
				os.Remove("package-lock.json")
			case "yarn":
				os.Remove("yarn.lock")
			}
		}

		fmt.Println("")
		installMsg := fmt.Sprintf("%s Installing dependencies using %s", ui.BlueStyle().Render(ui.RocketIcon), ui.BoldStyle.Render(pkgManager))
		start := time.Now()

		var execCmd *exec.Cmd
		switch pkgManager {
		case "pnpm":
			if ciMode {
				execCmd = exec.Command("pnpm", "install", "--frozen-lockfile")
			} else {
				execCmd = exec.Command("pnpm", "i")
			}
		case "npm":
			if ciMode {
				execCmd = exec.Command("npm", "ci")
			} else {
				execCmd = exec.Command("npm", "i")
			}
		case "yarn":
			if ciMode {
				execCmd = exec.Command("yarn", "install", "--frozen-lockfile")
			} else {
				execCmd = exec.Command("yarn", "install")
			}
		}

		prepareCmdEnv(execCmd, nodeBinDir)

		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		err := execCmd.Run()

		duration := getDuration(start)
		fmt.Println("")
		if err == nil {
			fmt.Printf("%s %s %s (%s)\n", installMsg, ui.GreenStyle().Render(ui.CheckIcon), "", duration)
		} else {
			fmt.Printf("%s %s %s (%s)\n", ui.RedStyle().Render(ui.ErrorIcon), "Installation failed", "", duration)
			os.Exit(1)
		}
	},
}

func init() {
	InstallCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Force install: removes node_modules and existing lock files before installing.")
	InstallCmd.Flags().BoolVar(&ciMode, "ci", false, "CI mode: installs dependencies using the frozen lockfile.")
}
