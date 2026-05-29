package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	// Default/Fallback version info injected at build time (e.g. for Homebrew or manual build)
	Version   = "1.0.0"
	BuildDate = "2026-05-28"
)

// getVersionString resolves the installation version dynamically from Go runtime metadata, falling back to build variables
func getVersionString() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return Version
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of minhthetus-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("minhthetus-cli %s (%s)\n", getVersionString(), BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
