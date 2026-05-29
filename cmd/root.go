package cmd

import (
	"fmt"
	"os"

	"github.com/MinhTuLeHoang/minhthetus-cli/cmd/demo"
	"github.com/MinhTuLeHoang/minhthetus-cli/cmd/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/cmd/sys"
	"github.com/MinhTuLeHoang/minhthetus-cli/cmd/vng"
	"github.com/MinhTuLeHoang/minhthetus-cli/cmd/web"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/loader"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "minhthetus-cli",
	Short: "A professional, high-performance CLI for developer productivity",
	Long: `minhthetus-cli is a native Go-compiled tool designed to automate common 
developer tasks with near-zero latency. It provides a rich interactive TUI 
and supports various modules like git, sys, web, and more.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Auto-setup completion (silent)
	SetupCompletion(true)

	// Register commands from sub-packages
	rootCmd.AddCommand(git.Cmd)
	rootCmd.AddCommand(sys.Cmd)
	rootCmd.AddCommand(web.Cmd)
	rootCmd.AddCommand(vng.Cmd)
	rootCmd.AddCommand(demo.Cmd)

	// Auto-discovery: Load scripts from src/scripts (Phase 3 & 5)
	loader.DiscoverCommands(rootCmd, "src/scripts")

	// Custom Help (Phase 5)
	ui.SetCustomHelp(rootCmd)

	// Custom Error Handling
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(ui.ErrorMessage(err.Error()))
		rootCmd.Help()
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = getVersionString()
	// Global flags can be defined here.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
