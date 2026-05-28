package cmd

import (
	"fmt"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "A simple greeting script that welcomes you by name.",
	Annotations: map[string]string{
		"title": "Hello",
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
}
