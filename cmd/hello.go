package cmd

import (
	"fmt"
	"os"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "A simple greeting script that welcomes you by name.",
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Hello",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-n", "--name", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = "User"
		}
		fmt.Printf("%s Hello %s!\n", ui.SuccessMessage(""), name)
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
	helloCmd.Flags().StringP("name", "n", "World", "Name to greet")
	
	helloCmd.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		currentUser := os.Getenv("USER")
		suggestions := []string{"World", "Developer", "Admin"}
		if currentUser != "" {
			suggestions = append([]string{currentUser}, suggestions...)
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	})
}
