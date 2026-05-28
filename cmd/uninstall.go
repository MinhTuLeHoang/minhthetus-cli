package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely remove the CLI and all integrations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⏳ Preparing uninstallation...")
		fmt.Println("⚠️ This feature is still being ported to Go.")
		fmt.Println("👉 Please use 'npm uninstall -g minhthetus-cli' for the legacy version.")
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
