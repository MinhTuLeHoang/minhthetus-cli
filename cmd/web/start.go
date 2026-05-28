package web

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start [options] [-- [args]]",
	Short: "Starts the web project development server with automatic environment detection.",
	Long:  "Detects the environment and starts the development server using the appropriate package manager.",
	Example: `minhthetus-cli web start --port 3000`,
	Annotations: map[string]string{
		"title": "Web Project Starter",
	},
	Run: func(cmd *cobra.Command, args []string) {
		trackCurrentRepo()

		fmt.Printf("%s Detecting environment...\n", ui.InfoMessage(""))
		pkgManager := getWebInfo()

		if pkgManager == "" {
			fmt.Printf("\n%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Failed to detect package manager."))
			os.Exit(1)
		}

		fmt.Printf("\n%s Starting project using %s...\n", ui.BlueStyle().Render(ui.RocketIcon), ui.BoldStyle.Render(pkgManager))

		var execCmd *exec.Cmd
		switch pkgManager {
		case "pnpm":
			execCmd = exec.Command("pnpm", append([]string{"start"}, args...)...)
		case "npm":
			execCmd = exec.Command("npm", append([]string{"start"}, args...)...)
		case "yarn":
			execCmd = exec.Command("yarn", append([]string{"start"}, args...)...)
		}

		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin
		
		if err := execCmd.Run(); err != nil {
			fmt.Printf("\n%s %s\n", ui.ErrorIcon, ui.RedStyle().Render(fmt.Sprintf("Start failed: %v", err)))
			os.Exit(1)
		}
	},
}
