//go:build dev

package debug

import (
	"github.com/spf13/cobra"
)

// Cmd represents the debug command
var Cmd = &cobra.Command{
	Use:   "debug",
	Short: "Developer debug utilities",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(SwitchNodeCmd)
	Cmd.AddCommand(GumDemoCmd)
	Cmd.AddCommand(PushRcCmd)
}
