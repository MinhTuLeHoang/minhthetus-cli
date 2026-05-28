package vng

import (
	"fmt"
	"github.com/spf13/cobra"
)

var DownloadCodeCmd = &cobra.Command{
	Use:   "download-code",
	Short: "Downloads the project codebase from the remote VNG server.",
	Annotations: map[string]string{
		"title": "Download Code",
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Download code...")
	},
}
