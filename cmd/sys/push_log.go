package sys

import (
	"fmt"

	"github.com/spf13/cobra"
)

var PushLogCmd = &cobra.Command{
	Use:   "push-log",
	Short: "Pushes system-level log files to a remote storage.",
	Annotations: map[string]string{
		"title": "Push Logs",
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Push logs...")
	},
}
