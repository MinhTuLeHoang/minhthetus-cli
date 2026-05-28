package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var SyncBranchCmd = &cobra.Command{
	Use:   "sync-branch",
	Short: "Synchronizes the current branch's latest commit to dev and staging branches.",
	Long:  "Synchronizes the current branch's latest commit to dev and staging branches. Automatically handles rebase if the branch is linear or cherry-picks otherwise.",
	Example: `minhthetus-cli git sync-branch`,
	Annotations: map[string]string{
		"title": "Git Branch Sync",
	},
	Run: func(cmd *cobra.Command, args []string) {
		originalBranch := getSyncCurrentBranch()
		latestCommit := getLatestCommit()
		commitMessage := getCommitMessage()

		fmt.Println("")
		fmt.Printf("%s Starting sync script...\n", ui.InfoMessage(""))
		fmt.Println("")
		fmt.Printf("%s Original Branch: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render(originalBranch))
		fmt.Printf("%s Latest Commit: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render(latestCommit))
		fmt.Printf("%s Commit Message: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render(commitMessage))
		fmt.Println("")

		// Fetch latest info
		fmt.Printf("%s Fetching origin/dev and origin/staging...\n", ui.HourglassIcon)
		exec.Command("git", "fetch", "origin", "dev").Run()
		exec.Command("git", "fetch", "origin", "staging").Run()

		devHash := getHash("origin/dev", "dev")
		stagingHash := getHash("origin/staging", "staging")

		fmt.Println("")
		fmt.Printf("%s dev hash: %s\n", ui.InfoMessage(""), ui.BlueStyle().Render(devHash))
		fmt.Printf("%s staging hash: %s\n", ui.InfoMessage(""), ui.BlueStyle().Render(stagingHash))
		fmt.Println("")

		if devHash == "" || stagingHash == "" {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Failed to identify dev or staging hash.")
			os.Exit(1)
		}

		devIsAncestor := isAncestor(devHash, latestCommit)
		stagingIsAncestor := isAncestor(stagingHash, latestCommit)

		devStatus := ""
		stagingStatus := ""

		if devIsAncestor && stagingIsAncestor {
			// Linear Case
			fmt.Printf("%s %s\n", ui.TagIcon, ui.GreenStyle().Render("Detected Linear Case: Both dev and staging are ancestors."))
			fmt.Printf("%s %s\n", ui.TagIcon, ui.GreenStyle().Render("Syncing with rebase/FF (no cherry-pick needed)."))
			fmt.Println("")

			// Sync dev
			fmt.Printf("%s Checking out dev...\n", ui.HourglassIcon)
			exec.Command("git", "checkout", "dev").Run()
			fmt.Printf("%s Pulling latest dev from origin (rebase)...\n", ui.HourglassIcon)
			exec.Command("git", "pull", "--rebase", "origin", "dev").Run()
			fmt.Printf("%s Rebasing dev onto %s...\n", ui.HammerIcon, originalBranch)
			if err := exec.Command("git", "rebase", originalBranch).Run(); err != nil {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render("Rebase onto "+originalBranch+" failed."))
				exec.Command("git", "rebase", "--abort").Run()
				exec.Command("git", "checkout", originalBranch).Run()
				printFinalStatus("Rebase failed", "Skipped")
				os.Exit(1)
			}
			exec.Command("git", "push", "origin", "dev", "--force").Run()
			fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("dev updated and pushed."))
			fmt.Println("")

			// Sync staging
			fmt.Printf("%s Checking out staging...\n", ui.HourglassIcon)
			exec.Command("git", "checkout", "staging").Run()
			fmt.Printf("%s Pulling dev into staging (rebase)...\n", ui.HourglassIcon)
			exec.Command("git", "pull", "--rebase", "origin", "dev").Run()
			exec.Command("git", "push", "origin", "staging", "--force").Run()
			fmt.Printf("%s %s\n", ui.HammerIcon, "Sync dev to staging successfully")

			devStatus = ui.SuccessMessage("Successfully rebased and pushed")
			stagingStatus = ui.SuccessMessage("Successfully pulled from dev (rebase) and pushed")

		} else if devHash == stagingHash {
			// Case A: same node
			fmt.Printf("%s %s\n", ui.TagIcon, ui.GreenStyle().Render("Detected Case A: dev and staging are at the same node."))
			fmt.Println("")

			// Checkout dev
			fmt.Printf("%s Checking out dev...\n", ui.HourglassIcon)
			if err := exec.Command("git", "checkout", "dev").Run(); err != nil {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Failed to checkout dev")
				os.Exit(1)
			}

			// Cherry pick
			fmt.Printf("%s Cherry picking %s into dev...\n", ui.HammerIcon, latestCommit)
			if err := exec.Command("git", "cherry-pick", latestCommit).Run(); err == nil {
				exec.Command("git", "push", "origin", "dev").Run()
				fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("Cherry pick into dev successful."))
				fmt.Println("")

				// Sync staging
				fmt.Printf("%s Checking out staging...\n", ui.HourglassIcon)
				exec.Command("git", "checkout", "staging").Run()
				fmt.Printf("%s Pulling dev into staging (rebase)...\n", ui.HourglassIcon)
				if err := exec.Command("git", "pull", "--rebase", "origin", "dev").Run(); err == nil {
					exec.Command("git", "push", "origin", "staging").Run()
					fmt.Printf("%s %s\n", ui.HammerIcon, "Sync dev to staging successfully")
					devStatus = ui.SuccessMessage("Successfully cherry-picked and pushed")
					stagingStatus = ui.SuccessMessage("Successfully pulled from dev and pushed")
				} else {
					fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Failed to pull dev into staging")
					devStatus = ui.SuccessMessage("Successfully cherry-picked and pushed")
					stagingStatus = ui.ErrorMessage("Pull from dev failed")
				}
			} else {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render("Conflict occurred while cherry-picking into dev."))
				exec.Command("git", "cherry-pick", "--abort").Run()
				devStatus = ui.ErrorMessage("Cherry-pick conflict")
				stagingStatus = ui.InfoMessage("Skipped")
				exec.Command("git", "checkout", originalBranch).Run()
				printFinalStatus(devStatus, stagingStatus)
				os.Exit(1)
			}
		} else {
			// Case B: different nodes
			fmt.Printf("%s %s\n", ui.TagIcon, ui.YellowStyle().Render("Detected Case B: dev and staging are at different nodes."))
			fmt.Println("")

			// Sync dev
			fmt.Printf("%s Checking out dev...\n", ui.HourglassIcon)
			exec.Command("git", "checkout", "dev").Run()
			fmt.Printf("%s Cherry picking %s into dev...\n", ui.HammerIcon, latestCommit)
			if err := exec.Command("git", "cherry-pick", latestCommit).Run(); err == nil {
				fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("Cherry pick into dev successful."))
				fmt.Printf("%s Pushing dev...\n", ui.RocketIcon)
				exec.Command("git", "push", "origin", "dev").Run()
				devStatus = ui.SuccessMessage("Successfully cherry-picked and pushed")
			} else {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render("Conflict occurred while cherry-picking into dev."))
				exec.Command("git", "cherry-pick", "--abort").Run()
				devStatus = ui.ErrorMessage("Cherry-pick conflict")
				stagingStatus = ui.InfoMessage("Skipped due to dev failure")
				exec.Command("git", "checkout", originalBranch).Run()
				printFinalStatus(devStatus, stagingStatus)
				os.Exit(1)
			}

			// Sync staging
			fmt.Printf("%s Checking out staging...\n", ui.HourglassIcon)
			exec.Command("git", "checkout", "staging").Run()
			fmt.Printf("%s Cherry picking %s into staging...\n", ui.HammerIcon, latestCommit)
			if err := exec.Command("git", "cherry-pick", latestCommit).Run(); err == nil {
				fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("Cherry pick into staging successful."))
				fmt.Printf("%s Pushing staging...\n", ui.RocketIcon)
				exec.Command("git", "push", "origin", "staging").Run()
				stagingStatus = ui.SuccessMessage("Successfully cherry-picked and pushed")
			} else {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render("Conflict occurred while cherry-picking into staging."))
				exec.Command("git", "cherry-pick", "--abort").Run()
				stagingStatus = ui.ErrorMessage("Cherry-pick conflict")
				exec.Command("git", "checkout", originalBranch).Run()
				printFinalStatus(devStatus, stagingStatus)
				os.Exit(1)
			}
		}

		// Return
		fmt.Printf("\n%s Returning to original branch: %s\n", ui.HourglassIcon, originalBranch)
		exec.Command("git", "checkout", originalBranch).Run()

		printFinalStatus(devStatus, stagingStatus)

		if strings.Contains(devStatus, ui.CheckIcon) && strings.Contains(stagingStatus, ui.CheckIcon) {
			fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("All branches updated successfully!"))
		} else {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render("Some updates failed or were skipped."))
			os.Exit(1)
		}
	},
}

func getSyncCurrentBranch() string {
	out, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	return strings.TrimSpace(string(out))
}

func getLatestCommit() string {
	out, _ := exec.Command("git", "rev-parse", "HEAD").Output()
	return strings.TrimSpace(string(out))
}

func getCommitMessage() string {
	out, _ := exec.Command("git", "log", "-1", "--pretty=%s").Output()
	return strings.TrimSpace(string(out))
}

func getHash(ref, fallback string) string {
	out, err := exec.Command("git", "rev-parse", ref).Output()
	if err != nil {
		out, err = exec.Command("git", "rev-parse", fallback).Output()
		if err != nil {
			return ""
		}
	}
	return strings.TrimSpace(string(out))
}

func isAncestor(ancestor, commit string) bool {
	err := exec.Command("git", "merge-base", "--is-ancestor", ancestor, commit).Run()
	return err == nil
}

func printFinalStatus(dev, staging string) {
	fmt.Println("")
	fmt.Printf("%s FINAL STATUS:\n", ui.InfoMessage(""))
	fmt.Printf("dev: %s\n", dev)
	fmt.Printf("staging: %s\n", staging)
	fmt.Println("")
}
