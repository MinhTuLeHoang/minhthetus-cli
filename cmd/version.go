package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "2.0.0"
	BuildDate = "2026-05-05"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of minhthetus-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("minhthetus-cli v%s (%s)\n", Version, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
