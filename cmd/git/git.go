package git

import (
	"github.com/spf13/cobra"
)

// Cmd represents the git command
var Cmd = &cobra.Command{
	Use:   "git",
	Short: "Advanced Git automation and account management",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Add subcommands here
	Cmd.AddCommand(CheckoutCmd)
	Cmd.AddCommand(MergeRequestCmd)
	Cmd.AddCommand(AccountCmd)
	Cmd.AddCommand(TagDevStgCmd)
	Cmd.AddCommand(BackupBranchCmd)
	Cmd.AddCommand(SyncBranchCmd)
	Cmd.AddCommand(ListRepoCmd)
	Cmd.AddCommand(TreeCmd)
}
