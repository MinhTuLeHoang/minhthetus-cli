package web

import (
	"github.com/spf13/cobra"
)

// Cmd represents the web command
var Cmd = &cobra.Command{
	Use:   "web",
	Short: "Web project building and dependency management",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(BuildCmd)
	Cmd.AddCommand(InstallCmd)
	Cmd.AddCommand(StartCmd)
	Cmd.AddCommand(CheckMalwareCmd)
}
