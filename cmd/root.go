package cmd

import (
	"fmt"
	"os"

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

// Hook for conditionally registering the developer debug command package
var registerDebug func() = nil

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Register commands from sub-packages
	rootCmd.AddCommand(git.Cmd)
	rootCmd.AddCommand(sys.Cmd)
	rootCmd.AddCommand(web.Cmd)
	rootCmd.AddCommand(vng.Cmd)

	if registerDebug != nil {
		registerDebug()
	}

	// Auto-discovery: Load scripts from src/scripts (Phase 3 & 5)
	loader.DiscoverCommands(rootCmd, "src/scripts")

	// Custom Help (Phase 5)
	ui.SetCustomHelp(rootCmd)

	// Auto-setup completion (silent)
	// We run this after registering all subcommands to ensure they are included in the generated completions
	SetupCompletion(true, false)

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
	// Global flags can be defined here.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
