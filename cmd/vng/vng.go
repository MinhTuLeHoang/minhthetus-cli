package vng

import (
	"github.com/spf13/cobra"
)

// Cmd represents the vng command
var Cmd = &cobra.Command{
	Use:   "vng",
	Short: "Project code download and publishing utilities",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	Cmd.AddCommand(PublishCodeCmd)
	Cmd.AddCommand(DownloadCodeCmd)
}
