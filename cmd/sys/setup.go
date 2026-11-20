package sys

import (
	"github.com/spf13/cobra"
)

// SetupCmd represents the sys setup parent command.
var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure system-level integrations and security settings",
	Long:  `Parent command for system setup utilities such as biometric authentication.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	SetupCmd.AddCommand(FingerPrintCmd)
}
