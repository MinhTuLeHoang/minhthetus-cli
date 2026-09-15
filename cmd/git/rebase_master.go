package git

import (
	"fmt"
	"os"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var RebaseMasterCmd = &cobra.Command{
	Use:   "rebase-master",
	Short: "Rebase current branch onto origin/HEAD (main branch) and force push if successful, or abort on conflict",
	Long:  "Attempts to rebase the current branch onto origin/HEAD (the default main branch of remote origin). If successful without conflicts, force pushes to remote origin. If conflicts occur, automatically aborts the rebase and restores the working state.",
	Example: `minhthetus-cli git rebase-master`,
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Git Rebase Master",
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Get current branch
		currentBranch, err := git.Run("rev-parse", "--abbrev-ref", "HEAD")
		if err != nil || currentBranch == "" {
			fmt.Printf("%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Not a git repository or unable to determine current branch."))
			os.Exit(1)
		}

		// 2. Check for uncommitted changes
		statusOut, err := git.Run("status", "--porcelain")
		if err != nil {
			fmt.Printf("%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Failed to check git working tree status."))
			os.Exit(1)
		}
		if statusOut != "" {
			fmt.Printf("%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("You have uncommitted changes in your working directory. Please commit or stash them before rebasing."))
			os.Exit(1)
		}

		// 3. Fetch latest origin
		fmt.Printf("%s Fetching latest origin...\n", ui.HourglassIcon)
		_, err = git.Run("fetch", "origin")
		if err != nil {
			fmt.Printf("%s %s: %v\n", ui.ErrorIcon, ui.RedStyle().Render("Failed to fetch from remote origin"), err)
			os.Exit(1)
		}

		// 4. Resolve target ref from origin/HEAD
		targetRef, mainBranch := resolveOriginHeadTarget()

		// 5. Prevent running directly on default main branch
		if currentBranch == mainBranch {
			fmt.Printf("%s %s\n", ui.ErrorIcon, ui.RedStyle().Render(fmt.Sprintf("Branch '%s' is the target default branch and cannot be rebased onto itself. Please run this command from a feature or working branch.", currentBranch)))
			os.Exit(1)
		}

		fmt.Println("")
		fmt.Printf("%s Current branch: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render(currentBranch))
		fmt.Printf("%s Target ref: %s (%s)\n\n", ui.InfoMessage(""), ui.CyanStyle().Render(targetRef), ui.CyanStyle().Render(mainBranch))

		// 5.1 Confirmation prompt if target branch is not master or main
		if mainBranch != "master" && mainBranch != "main" {
			confirmed, err := ui.Confirm(fmt.Sprintf("Target branch '%s' is not 'master' or 'main'. Are you sure you want to rebase onto %s (%s)?", mainBranch, targetRef, mainBranch), 0, false)
			if err != nil || !confirmed {
				fmt.Printf("%s Rebase cancelled by user.\n", ui.InfoMessage(""))
				os.Exit(0)
			}
		}

		// 6. Attempt rebase onto origin/HEAD
		fmt.Printf("%s Rebasing %s onto %s...\n", ui.HammerIcon, ui.CyanStyle().Render(currentBranch), ui.CyanStyle().Render(targetRef))
		_, rebaseErr := git.Run("rebase", targetRef)

		if rebaseErr != nil {
			// Conflict or rebase error -> abort rebase
			fmt.Printf("%s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Conflict detected during rebase onto "+targetRef+"."))
			fmt.Printf("%s Aborting rebase...\n", ui.HourglassIcon)
			git.Run("rebase", "--abort")
			fmt.Printf("%s %s\n", ui.WarningIcon, ui.YellowStyle().Render("Rebase aborted. Working branch restored to original state."))
			os.Exit(1)
		}

		// Success -> push force
		fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("Rebase succeeded cleanly (no conflicts)."))
		fmt.Printf("%s Force pushing %s to origin...\n", ui.RocketIcon, ui.CyanStyle().Render(currentBranch))

		_, pushErr := git.Run("push", "origin", currentBranch, "--force")
		if pushErr != nil {
			fmt.Printf("%s %s: %v\n", ui.ErrorIcon, ui.RedStyle().Render("Failed to force push to origin"), pushErr)
			os.Exit(1)
		}

		fmt.Printf("\n%s %s\n", ui.CheckIcon, ui.GreenStyle().Render(fmt.Sprintf("Successfully rebased %s onto %s (%s) and force pushed to remote origin!", currentBranch, targetRef, mainBranch)))
	},
}

func resolveOriginHeadTarget() (targetRef string, mainBranch string) {
	// Query symbolic ref for origin/HEAD
	ref, err := git.Run("symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil || ref == "" {
		// Auto detect origin HEAD if not configured locally
		git.Run("remote", "set-head", "origin", "--auto")
		ref, _ = git.Run("symbolic-ref", "refs/remotes/origin/HEAD")
	}

	if ref != "" {
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			branch := parts[len(parts)-1]
			return "origin/HEAD", branch
		}
	}

	return "origin/HEAD", "HEAD"
}
