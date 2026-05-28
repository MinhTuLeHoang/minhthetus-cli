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
		fmt.Println("👉 To completely uninstall the Go version of minhthetus-cli, run:")
		fmt.Println("   sudo rm /usr/local/bin/minhthetus-cli")
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
