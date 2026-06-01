package git

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/project"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	major         bool
	minor         bool
	patch         bool
	noVersion     bool
	commitMessage string
)

// MergeRequestCmd represents the git merge-request command
var MergeRequestCmd = &cobra.Command{
	Use:   "merge-request",
	Short: "Automatically bumps version, commits changes, and prepares a Merge Request",
	Args:  cobra.NoArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-M", "--major", "-N", "--minor", "-P", "--patch", "--no-version", "-m", "--message", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		currentBranch, err := git.Run("rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			fmt.Printf("Error getting current branch: %v\n", err)
			return
		}

		// Detect bump type
		incrementType := ""
		if major {
			incrementType = "major"
		} else if minor {
			incrementType = "minor"
		} else if patch {
			incrementType = "patch"
		} else {
			// Auto-detect from branch prefix
			if match, _ := regexp.MatchString("^(fix/|hotfix/|docs/|test/|debug/)", currentBranch); match {
				incrementType = "patch"
			} else {
				incrementType = "minor"
			}
		}

		if commitMessage == "" {
			var err error
			commitMessage, err = ui.Input("Enter commit message (default: [bump version])", "")
			if err != nil {
				return
			}
			if commitMessage == "" {
				commitMessage = "[bump version]"
			}
		}

		// --- Sync with master ---
		fmt.Println("⏳ Syncing with master before workflow...")
		git.Run("fetch", "origin", "master")
		if _, err := git.Run("rebase", "origin/master"); err != nil {
			fmt.Println("❌ Rebase conflict detected. Please resolve manually.")
			git.Run("rebase", "--abort")
			return
		}

		// MR Details
		jiraTicket := ""
		reg := regexp.MustCompile("[A-Z]+-[0-9]+")
		jiraTicket = reg.FindString(currentBranch)

		prefix := ""
		if strings.Contains(currentBranch, "/") {
			prefix = strings.Split(currentBranch, "/")[0]
		}
		prettyPrefix := ""
		if prefix != "" {
			prettyPrefix = strings.ToUpper(prefix[:1]) + prefix[1:] + "/ "
		}

		mrTitle := commitMessage
		closesText := ""
		if jiraTicket != "" {
			// Extract rest of branch name
			rest := currentBranch
			rest = strings.TrimPrefix(rest, prefix+"/")
			rest = strings.TrimPrefix(rest, jiraTicket)
			rest = strings.TrimLeft(rest, "-_")
			rest = strings.ReplaceAll(rest, "-", " ")
			rest = strings.ReplaceAll(rest, "_", " ")
			mrTitle = fmt.Sprintf("Resolve %s \"%s%s\"", jiraTicket, prettyPrefix, strings.TrimSpace(rest))
			closesText = "Closes " + jiraTicket
		}

		commitList, _ := git.Run("log", "origin/master..HEAD", "--oneline", "--format=- %s")
		mrDescription := closesText + "\n\n" + commitList

		// --- Version Bump ---
		if !noVersion {
			oldVersion, err := project.GetVersion()
			if err == nil {
				newVersion, _ := project.BumpVersion(oldVersion, incrementType)
				fmt.Printf("🔨 Bumping %s version: %s -> %s\n", incrementType, oldVersion, newVersion)

				confirmed, err := ui.Confirm(fmt.Sprintf("Bump version to %s?", newVersion), 3*time.Second, true)
				if err != nil || !confirmed {
					fmt.Println("⚠️ Version bump cancelled.")
				} else {
					if err := project.UpdateVersion(newVersion); err != nil {
						fmt.Printf("Error updating version: %v\n", err)
					}
				}
			}
		}

		// --- Commit and Push ---
		fmt.Println("\n🚀 Committing changes...")
		git.Run("add", ".")
		diff, _ := git.Run("diff", "--staged", "--quiet")
		if diff != "" {
			git.Run("commit", "-m", commitMessage)
		}

		fmt.Println("🚀 Pushing current branch to origin...")
		git.RunInteractive("push", "origin", currentBranch)

		// --- PR Generation ---
		remoteURL, _ := git.Run("remote", "get-url", "origin")
		fmt.Println("\n⏳ Preparing Merge Request link...")

		if strings.Contains(remoteURL, "github.com") {
			repoPath := ""
			// git@github.com:user/repo.git or https://github.com/user/repo.git
			if strings.Contains(remoteURL, "git@") {
				repoPath = strings.Split(strings.Split(remoteURL, ":")[1], ".git")[0]
			} else {
				repoPath = strings.Split(strings.Split(remoteURL, "github.com/")[1], ".git")[0]
			}

			titleEnc := url.QueryEscape(mrTitle)
			bodyEnc := url.QueryEscape(mrDescription)
			link := fmt.Sprintf("https://github.com/%s/compare/master...%s?expand=1&quick_pull=1&title=%s&body=%s", repoPath, currentBranch, titleEnc, bodyEnc)
			fmt.Printf("✅ GitHub PR Link: %s\n", link)
		} else if strings.Contains(remoteURL, "gitlab") {
			// GitLab logic...
			fmt.Println("✅ GitLab detected. Please use the web interface or GitLab push options.")
		}

		fmt.Println("\n✅ Workflow complete!")
	},
}

func init() {
	MergeRequestCmd.Flags().BoolVarP(&major, "major", "M", false, "Force major version bump")
	MergeRequestCmd.Flags().BoolVarP(&minor, "minor", "N", false, "Force minor version bump")
	MergeRequestCmd.Flags().BoolVarP(&patch, "patch", "P", false, "Force patch version bump")
	MergeRequestCmd.Flags().BoolVar(&noVersion, "no-version", false, "Skip version bump step")
	MergeRequestCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message")
	
	MergeRequestCmd.RegisterFlagCompletionFunc("message", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		prefixes := []string{
			"feat: impl ",
			"fix: resolve ",
			"chore: update ",
			"docs: update doc for ",
			"refactor: reorganize ",
			"test: add tests for ",
		}
		return prefixes, cobra.ShellCompDirectiveNoFileComp
	})
}
