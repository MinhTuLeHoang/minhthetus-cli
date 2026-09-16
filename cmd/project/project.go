package project

import (
	"github.com/spf13/cobra"
)

// Cmd represents the project command
var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Project-level utilities and configurations",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(NavigateLinkCmd)
}
