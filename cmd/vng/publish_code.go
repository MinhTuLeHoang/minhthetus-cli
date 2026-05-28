package vng

import (
	"fmt"
	"github.com/spf13/cobra"
)

var PublishCodeCmd = &cobra.Command{
	Use:   "publish-code",
	Short: "Publishes the project codebase to the remote VNG server.",
	Annotations: map[string]string{
		"title": "Publish Code",
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Publishing code...")
	},
}
