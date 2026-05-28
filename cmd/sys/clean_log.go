package sys

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CleanLogCmd represents the sys clean-log command
var CleanLogCmd = &cobra.Command{
	Use:   "clean-log",
	Short: "Cleans up system-level log files.",
	Annotations: map[string]string{
		"title": "Clean Logs",
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning logs...")
	},
}
