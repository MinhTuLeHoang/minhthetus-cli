package git

import (
	"fmt"
	"os"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var RenameBranchCmd = &cobra.Command{
	Use:   "rename-branch",
	Short: "Safely rename local and remote Git branch",
	Long: `Safely rename the current Git branch both locally and on the remote origin after verifying branch protection rules and uncommitted changes.`,
	Example: `minhthetus-cli git rename-branch`,
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Git Rename Branch",
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Get current branch
		currentBranch, err := git.Run("rev-parse", "--abbrev-ref", "HEAD")
		if err != nil || currentBranch == "" {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Not a git repository or unable to determine current branch.")
			os.Exit(1)
		}

		// 2. Check if current branch is protected
		if git.IsProtectedBranch(currentBranch) {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Branch '%s' is a protected branch and cannot be renamed.", currentBranch)))
			os.Exit(1)
		}

		// 3. Check for uncommitted changes
		statusOut, err := git.Run("status", "--porcelain")
		if err != nil {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Failed to check git working tree status.")
			os.Exit(1)
		}
		if statusOut != "" {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render("You have uncommitted changes in your working directory. Please commit or stash them before renaming."))
			os.Exit(1)
		}

		// 4. Display current branch
		fmt.Printf("\n%s Current branch: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render(currentBranch))

		// 5. Ask user for new branch name
		newBranch, err := ui.Input("Enter new branch name", "")
		if err != nil {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Cancelled or failed to read branch name input.")
			os.Exit(1)
		}

		newBranch = strings.TrimSpace(newBranch)
		if newBranch == "" {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "New branch name cannot be empty.")
			os.Exit(1)
		}

		if newBranch == currentBranch {
			fmt.Printf("%s %s\n", ui.InfoMessage(""), "New branch name is identical to the current branch name. No changes made.")
			return
		}

		// 6. Validate target branch name (check if it exists locally or on remote)
		if git.BranchExistsLocally(newBranch) {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Target branch '%s' already exists locally.", newBranch)))
			os.Exit(1)
		}

		if git.BranchExistsRemotely(newBranch) {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Target branch '%s' already exists on remote origin.", newBranch)))
			os.Exit(1)
		}

		// 7. Perform local and remote branch rename
		remoteOldExists := git.BranchExistsRemotely(currentBranch)

		fmt.Printf("%s Renaming local branch from %s to %s...\n", ui.HourglassIcon, ui.CyanStyle().Render(currentBranch), ui.GreenStyle().Render(newBranch))

		_, err = git.Run("branch", "-m", newBranch)
		if err != nil {
			fmt.Printf("%s %s: %v\n", ui.ErrorMessage(""), ui.RedStyle().Render("Failed to rename local branch"), err)
			os.Exit(1)
		}

		if remoteOldExists {
			fmt.Printf("%s Remote branch origin/%s detected. Updating remote origin...\n", ui.InfoMessage(""), currentBranch)

			fmt.Printf("  %s Pushing new branch %s to origin...\n", ui.HourglassIcon, ui.GreenStyle().Render(newBranch))
			if _, err := git.Run("push", "origin", "-u", newBranch); err != nil {
				fmt.Printf("%s %s: %v\n", ui.ErrorMessage(""), ui.RedStyle().Render("Failed to push new branch to remote origin"), err)
				os.Exit(1)
			}

			fmt.Printf("  %s Deleting old branch %s on origin...\n", ui.HourglassIcon, ui.RedStyle().Render(currentBranch))
			if _, err := git.Run("push", "origin", "--delete", currentBranch); err != nil {
				fmt.Printf("%s Warning: Failed to delete old remote branch origin/%s: %v\n", ui.WarningIcon, currentBranch, err)
			}

			fmt.Printf("\n%s %s\n", ui.CheckIcon, ui.GreenStyle().Render(fmt.Sprintf("Successfully renamed branch '%s' -> '%s' (locally and on remote origin).", currentBranch, newBranch)))
		} else {
			fmt.Printf("\n%s %s\n", ui.CheckIcon, ui.GreenStyle().Render(fmt.Sprintf("Successfully renamed local branch '%s' -> '%s'.", currentBranch, newBranch)))
		}
	},
}

