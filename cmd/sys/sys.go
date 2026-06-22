package sys

import (
	"github.com/spf13/cobra"
)

// Cmd represents the sys command
var Cmd = &cobra.Command{
	Use:   "sys",
	Short: "System utilities and log management",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(CleanLogCmd)
	Cmd.AddCommand(PushLogCmd)
	Cmd.AddCommand(SizeCmd)
	Cmd.AddCommand(CleanMacCmd)
}
