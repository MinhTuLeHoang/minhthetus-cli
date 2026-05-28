package demo

import (
	"github.com/spf13/cobra"
)

// Cmd represents the demo command
var Cmd = &cobra.Command{
	Use:   "demo",
	Short: "Demonstrations of CLI capabilities",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(SwitchNodeCmd)
	Cmd.AddCommand(GumDemoCmd)
}
