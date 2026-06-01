package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	jiraID string
)

// CheckoutCmd represents the git checkout command
var CheckoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Checkout existing branches or create new ones using JIRA ticket IDs",
	Long: `Checks out an existing branch matching a JIRA ID, or creates a new one 
following a standardized naming convention. Supports numeric shorthand 
(e.g. 9404 -> PCFBANK-9404).`,
	Example: `minhthetus-cli git checkout --jira-ticket PCFBANK-9404
minhthetus-cli git checkout`,
	Annotations: map[string]string{
		"title": "Git Checkout",
	},
	Run: func(cmd *cobra.Command, args []string) {
		if jiraID == "" {
			var err error
			jiraID, err = ui.Input("Enter JIRA ID (e.g. 9404 or PCFBANK-9404)", "")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
		}

		// Normalize JIRA ID
		jiraID = strings.TrimSpace(jiraID)
		if jiraID == "none" || jiraID == "" {
			jiraID = ""
		} else {
			// Auto-prefix if numeric only
			if match, _ := regexp.MatchString("^[0-9]+$", jiraID); match {
				jiraID = "PCFBANK-" + jiraID
			}
			jiraID = strings.ToUpper(jiraID)
		}

		// --- Discovery ---
		if jiraID != "" {
			fmt.Printf("🔍 Searching for branches matching: %s\n", jiraID)
			branches, err := git.ListBranches(jiraID)
			if err != nil {
				fmt.Printf("Error listing branches: %v\n", err)
				return
			}

			if len(branches) > 0 {
				var selectedBranch string
				if len(branches) == 1 {
					selectedBranch = branches[0]
					fmt.Printf("✅ Found matching branch: %s\n", selectedBranch)
				} else {
					selectedBranch, err = ui.Choose("Multiple branches found. Select one:", branches)
					if err != nil {
						fmt.Printf("Error selecting branch: %v\n", err)
						return
					}
				}

				if selectedBranch != "" {
					if err := git.RunInteractive("checkout", selectedBranch); err != nil {
						fmt.Printf("Error checking out branch: %v\n", err)
						return
					}
					fmt.Println("⏳ Fetching latest updates from origin...")
					git.Run("pull", "origin", selectedBranch)
					return
				}
			}
		}

		// --- Creation ---
		fmt.Println("\nℹ️  No existing branch found. Entering creation flow...")

		types := []string{"feature", "features", "hotfix", "test", "docs", "improve", "bugfix", "refactor"}
		branchType, err := ui.Choose("Select branch type:", types)
		if err != nil || branchType == "" {
			return
		}

		desc, err := ui.Input("Enter branch description (e.g. update user profile)", "")
		if err != nil || desc == "" {
			return
		}

		// Format description
		reg := regexp.MustCompile("[^a-z0-9-]+")
		formattedDesc := strings.ToLower(desc)
		formattedDesc = strings.ReplaceAll(formattedDesc, " ", "-")
		formattedDesc = reg.ReplaceAllString(formattedDesc, "")
		
		finalName := branchType + "/" + formattedDesc
		if jiraID != "" {
			finalName = branchType + "/" + jiraID + "-" + formattedDesc
		}

		fmt.Printf("⏳ Creating and checking out: %s...\n", finalName)
		if err := git.RunInteractive("checkout", "-b", finalName); err != nil {
			fmt.Printf("Error creating branch: %v\n", err)
			return
		}

		fmt.Printf("✅ Successfully created and checked out %s\n", finalName)
		fmt.Println("⏳ Pushing to origin...")
		if _, err := git.Run("push", "-u", "origin", finalName); err != nil {
			fmt.Printf("⚠️ Failed to push to origin: %v\n", err)
		} else {
			fmt.Println("✅ Successfully pushed to origin.")
		}
	},
}

func init() {
	CheckoutCmd.Flags().StringVarP(&jiraID, "jira-ticket", "j", "", "JIRA ticket ID")
	
	CheckoutCmd.RegisterFlagCompletionFunc("jira-ticket", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		out, err := exec.Command("git", "branch", "--format=%(refname:short)").Output()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		
		re := regexp.MustCompile(`([A-Z]+-[0-9]+|[0-9]+)`)
		var suggestions []string
		seen := make(map[string]bool)
		
		for _, branch := range strings.Split(string(out), "\n") {
			branch = strings.TrimSpace(branch)
			if branch == "" {
				continue
			}
			matches := re.FindAllString(branch, -1)
			for _, m := range matches {
				val := strings.ToUpper(m)
				if !seen[val] {
					seen[val] = true
					suggestions = append(suggestions, val)
				}
			}
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	})
}
