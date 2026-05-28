package web

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build [options] [-- [args]]",
	Short: "Builds the web project with automatic environment detection.",
	Long:  "Detects the environment and runs the build script using the appropriate package manager.",
	Example: `minhthetus-cli web build`,
	Annotations: map[string]string{
		"title": "Web Project Builder",
	},
	Run: func(cmd *cobra.Command, args []string) {
		trackCurrentRepo()

		fmt.Printf("%s Detecting environment...\n", ui.InfoMessage(""))
		pkgManager := getWebInfo()

		if pkgManager == "" {
			fmt.Printf("\n%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Failed to detect package manager."))
			os.Exit(1)
		}

		fmt.Println("")
		buildMsg := fmt.Sprintf("%s Building project using %s", ui.BlueStyle().Render(ui.RocketIcon), ui.BoldStyle.Render(pkgManager))
		
		start := time.Now()
		var execCmd *exec.Cmd
		switch pkgManager {
		case "pnpm":
			execCmd = exec.Command("pnpm", append([]string{"run", "build"}, args...)...)
		case "npm":
			execCmd = exec.Command("npm", append([]string{"run", "build"}, args...)...)
		case "yarn":
			execCmd = exec.Command("yarn", append([]string{"run", "build"}, args...)...)
		}

		err := showSpinner(buildMsg, execCmd)
		duration := getDuration(start)

		if err == nil {
			fmt.Printf("\r  %s %s (%s)\n", buildMsg, ui.GreenStyle().Render(ui.CheckIcon), duration)
		} else {
			fmt.Printf("\r  %s %s (%s)\n", buildMsg, ui.RedStyle().Render(ui.ErrorIcon), duration)
			fmt.Printf("\n%s %s\n", ui.ErrorIcon, ui.RedStyle().Render(fmt.Sprintf("Build failed with error: %v", err)))
			os.Exit(1)
		}
	},
}
