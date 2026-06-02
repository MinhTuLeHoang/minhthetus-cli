package git

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	forceCLI bool
)

// TreeCmd represents the git tree command
var TreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display the Git commit tree or open the project in SourceTree",
	Long: `Visualizes the repository's commit graph. 

If Atlassian SourceTree is installed on macOS, it will automatically open the current repository in SourceTree. 

If SourceTree is not installed, or if the --cli flag is specified, it will display 
a beautiful, colorized ASCII commit graph directly in your terminal using an interactive pager.`,
	Example: `  minhthetus-cli git tree
  minhthetus-cli git tree --cli`,
	Annotations: map[string]string{
		"title": "Git Tree",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-c", "--cli"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if !forceCLI && isSourceTreeInstalled() {
			fmt.Printf("%s Opening the current repository in SourceTree...\n", ui.HourglassIcon)

			// Open current repository in SourceTree
			openCmd := exec.Command("open", "-a", "SourceTree", ".")
			if err := openCmd.Run(); err != nil {
				fmt.Printf("%s Error: Failed to open project in SourceTree: %v\n", ui.ErrorIcon, err)
				fmt.Println("Falling back to terminal tree visualization...")
				displayTerminalTree()
				return
			}

			fmt.Printf("%s Successfully opened current repository in SourceTree.\n", ui.CheckIcon)
			return
		}

		// Fallback/CLI option
		displayTerminalTree()
	},
}

func init() {
	TreeCmd.Flags().BoolVarP(&forceCLI, "cli", "c", false, "Force displaying the colored Git commit tree in terminal")
}

// isSourceTreeInstalled checks if SourceTree application exists or is registered in macOS.
func isSourceTreeInstalled() bool {
	// Standard path check
	if _, err := os.Stat("/Applications/SourceTree.app"); err == nil {
		return true
	}
	// osascript check
	cmd := exec.Command("osascript", "-e", "id of application \"Sourcetree\"")
	if err := cmd.Run(); err == nil {
		return true
	}
	// Case variation check
	cmd2 := exec.Command("osascript", "-e", "id of application \"SourceTree\"")
	if err := cmd2.Run(); err == nil {
		return true
	}
	return false
}

// displayTerminalTree prints the interactive and colorized ASCII git log graph.
func displayTerminalTree() {
	fmt.Printf("%s Generating beautiful ASCII Git commit tree...\n", ui.HourglassIcon)

	// Invoke colorized git log graph using standard interactive runner
	err := git.RunInteractive(
		"log",
		"--graph",
		"--abbrev-commit",
		"--decorate",
		"--color=always",
		"--format=format:%C(bold blue)%h%C(reset) - %C(bold green)(%ar)%C(reset) %C(white)%s%C(reset) %C(dim white)- %an%C(reset)%C(bold yellow)%d%C(reset)",
		"--all",
	)
	if err != nil {
		fmt.Printf("%s Error visualizing git tree: %v\n", ui.ErrorIcon, err)
	}
}
