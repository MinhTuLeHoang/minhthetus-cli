package git

import (
	"fmt"
	"os"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var DeleteBranchCmd = &cobra.Command{
	Use:   "delete-branch",
	Short: "Safely delete the current Git branch locally and/or on remote origin",
	Long:  `Safely delete the current Git branch locally and/or on the remote origin after checking protections, uncommitted changes, and prompting for targets.`,
	Example: `minhthetus-cli git delete-branch`,
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Git Delete Branch",
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
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Branch '%s' is a protected branch and cannot be deleted.", currentBranch)))
			os.Exit(1)
		}

		// 3. Check for uncommitted changes
		statusOut, err := git.Run("status", "--porcelain")
		if err != nil {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Failed to check git working tree status.")
			os.Exit(1)
		}

		// 4. Display current branch
		fmt.Printf("\n%s Current branch: %s\n\n", ui.InfoMessage(""), ui.CyanStyle().Render(currentBranch))

		// 5. Ask user if they want to delete local or remote (default select local)
		options := []string{
			"Local branch",
			fmt.Sprintf("Remote branch (origin/%s)", currentBranch),
		}
		preselected := []bool{true, false} // default select local, remote unselected
		selections, err := ui.MultiChoose("Select targets to delete:", options, preselected)
		if err != nil {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Cancelled or failed to read selection.")
			os.Exit(1)
		}

		deleteLocal := selections[0]
		deleteRemote := selections[1]

		if !deleteLocal && !deleteRemote {
			fmt.Printf("%s %s\n", ui.InfoMessage(""), "No targets selected for deletion. Exiting.")
			return
		}

		// If local delete is selected, verify uncommitted changes first to prevent losing work
		if deleteLocal && statusOut != "" {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render("You have uncommitted changes in your working directory. Please commit or stash them before deleting local branch."))
			os.Exit(1)
		}

		// 6. If remote is selected -> ask user to confirm using ui.Confirm
		if deleteRemote {
			remoteExists := git.BranchExistsRemotely(currentBranch)
			if !remoteExists {
				fmt.Printf("%s %s\n", ui.WarningMessage(""), fmt.Sprintf("Branch '%s' does not exist on remote origin.", currentBranch))
				if !deleteLocal {
					os.Exit(1)
				}
				deleteRemote = false
			} else {
				confirmed, err := ui.Confirm(fmt.Sprintf("Are you sure you want to delete remote branch origin/%s?", currentBranch), 0, false)
				if err != nil {
					fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Cancelled or failed to get confirmation.")
					os.Exit(1)
				}
				if !confirmed {
					fmt.Printf("%s Deletion of remote branch origin/%s cancelled.\n", ui.InfoMessage(""), currentBranch)
					deleteRemote = false
				}
			}
		}

		// Ensure we still have something to delete
		if !deleteLocal && !deleteRemote {
			fmt.Printf("%s %s\n", ui.InfoMessage(""), "No targets remaining for deletion. Exiting.")
			return
		}

		// 7. Execute delete branch
		var mainBranch string
		if deleteLocal {
			mainBranch, err = detectMainBranch()
			if err != nil {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(err.Error()))
				os.Exit(1)
			}

			// Checkout main branch
			fmt.Printf("%s Checking out %s branch...\n", ui.HourglassIcon, ui.CyanStyle().Render(mainBranch))
			_, err = git.Run("checkout", mainBranch)
			if err != nil {
				fmt.Printf("%s %s: %v\n", ui.ErrorMessage(""), ui.RedStyle().Render("Failed to checkout default branch "+mainBranch), err)
				os.Exit(1)
			}
			fmt.Printf("%s Switched to branch %s\n", ui.SwitchIcon, ui.CyanStyle().Render(mainBranch))
		}

		// Execute remote deletion
		if deleteRemote {
			fmt.Printf("%s Deleting remote branch origin/%s...\n", ui.HourglassIcon, ui.RedStyle().Render(currentBranch))
			_, err = git.Run("push", "origin", "--delete", currentBranch)
			if err != nil {
				fmt.Printf("%s %s: %v\n", ui.ErrorMessage(""), ui.RedStyle().Render("Failed to delete remote branch"), err)
			} else {
				fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render(fmt.Sprintf("Successfully deleted remote branch origin/%s.", currentBranch)))
			}
		}

		// Execute local deletion
		if deleteLocal {
			fmt.Printf("%s Deleting local branch %s...\n", ui.HourglassIcon, ui.RedStyle().Render(currentBranch))
			_, err = git.Run("branch", "-d", currentBranch)
			if err != nil {
				// If standard delete fails (not fully merged), ask user to confirm force delete
				if strings.Contains(err.Error(), "not fully merged") {
					fmt.Printf("%s %s\n", ui.WarningMessage(""), fmt.Sprintf("Local branch '%s' is not fully merged.", currentBranch))
					forceConfirmed, confirmErr := ui.Confirm("Do you want to force delete it (lose unmerged changes)?", 0, false)
					if confirmErr == nil && forceConfirmed {
						fmt.Printf("%s Force deleting local branch %s...\n", ui.HourglassIcon, ui.RedStyle().Render(currentBranch))
						_, err = git.Run("branch", "-D", currentBranch)
						if err != nil {
							fmt.Printf("%s %s: %v\n", ui.ErrorMessage(""), ui.RedStyle().Render("Failed to force delete local branch"), err)
						} else {
							fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render(fmt.Sprintf("Successfully force deleted local branch %s.", currentBranch)))
						}
					} else {
						fmt.Printf("%s Deletion of local branch %s skipped.\n", ui.InfoMessage(""), currentBranch)
					}
				} else {
					fmt.Printf("%s %s: %v\n", ui.ErrorMessage(""), ui.RedStyle().Render("Failed to delete local branch"), err)
				}
			} else {
				fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render(fmt.Sprintf("Successfully deleted local branch %s.", currentBranch)))
			}
		}
	},
}

func detectMainBranch() (string, error) {
	// Try resolving remote origin HEAD first
	defaultBranch := git.GetRemoteDefaultBranch()
	if defaultBranch != "" && git.BranchExistsLocally(defaultBranch) {
		return defaultBranch, nil
	}

	// Fallback to common branch names
	for _, b := range []string{"main", "master", "dev"} {
		if git.BranchExistsLocally(b) {
			return b, nil
		}
	}

	return "", fmt.Errorf("could not detect local main/master branch to switch to")
}

