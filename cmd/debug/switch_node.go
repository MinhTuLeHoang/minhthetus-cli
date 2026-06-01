//go:build dev

package debug

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var nodeVersion string

var SwitchNodeCmd = &cobra.Command{
	Use:   "switch-node",
	Short: "Tests the shell integration pipe by requesting a Node.js version switch via nvm.",
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Switch Node",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"--node", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if nodeVersion == "" {
			fmt.Printf("%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Missing --node <version> argument."))
			os.Exit(1)
		}

		shellPipe := os.Getenv("MINHTHETUS_SHELL_PIPE")
		if shellPipe != "" {
			fmt.Printf("✦ Sending request to switch Node to %s...\n", nodeVersion)

			f, err := os.OpenFile(shellPipe, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("%s Error opening shell pipe: %v\n", ui.ErrorIcon, err)
				os.Exit(1)
			}
			defer f.Close()

			fmt.Fprintf(f, "nvm use %s\n", nodeVersion)
			fmt.Printf("%s Instruction sent. Node will switch once the CLI session ends.\n", ui.CheckIcon)
		} else {
			fmt.Printf("%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Shell integration pipe not detected."))
			fmt.Println("Please ensure you have run 'minhthetus-cli setup-completion' and sourced your shell config.")
			os.Exit(1)
		}
	},
}

func init() {
	SwitchNodeCmd.Flags().StringVarP(&nodeVersion, "node", "", "", "The Node version to switch to (e.g., 20)")
	
	SwitchNodeCmd.RegisterFlagCompletionFunc("node", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		home, _ := os.UserHomeDir()
		nvmDir := filepath.Join(home, ".nvm", "versions", "node")
		var versions []string
		
		if entries, err := os.ReadDir(nvmDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					versions = append(versions, entry.Name())
				}
			}
		}
		
		if len(versions) == 0 {
			versions = []string{"18", "20", "22", "lts/*", "node"}
		}
		return versions, cobra.ShellCompDirectiveNoFileComp
	})
}
