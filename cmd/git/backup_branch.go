package git

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var listBackups bool

var BackupBranchCmd = &cobra.Command{
	Use:   "backup-branch",
	Short: "Creates a backup branch named backup/<current-branch>-dd-mm-yyyy-HHh-MM. Maintains up to 3 versions.",
	Long: `Creates a backup of the current branch and maintains only 3 recent versions.
Maintains up to 3 versions of backups for the current branch and prompts for cleanup if exceeded.`,
	Example: `minhthetus-cli git backup-branch
minhthetus-cli git backup-branch --list`,
	Annotations: map[string]string{
		"title": "Git Backup Branch",
	},
	Run: func(cmd *cobra.Command, args []string) {
		currentBranch := getCurrentBranch()
		if currentBranch == "" {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Not a git repository or no branch found.")
			os.Exit(1)
		}

		if listBackups {
			listBackupsMode(currentBranch)
			return
		}

		createBackupMode(currentBranch)
	},
}

func init() {
	BackupBranchCmd.Flags().BoolVarP(&listBackups, "list", "l", false, "List all backup branches for the current branch without creating a new one.")
}

func getCurrentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func listBackupsMode(currentBranch string) {
	fmt.Printf("\n%s Listing all backups for branch: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render(currentBranch))

	backups := getMatchedBackups(currentBranch)
	if len(backups) == 0 {
		fmt.Printf("  %s No backups found for this branch.\n", ui.InfoIcon)
	} else {
		fmt.Println("")
		for _, b := range backups {
			fmt.Printf("  %s %s\n", ui.BulletIcon, ui.GreenStyle().Render(b.Name))
		}
	}
	fmt.Println("")
}

func createBackupMode(currentBranch string) {
	now := time.Now()
	dateSuffix := now.Format("02-01-2006")
	timeSuffix := now.Format("15h-04")
	backupName := fmt.Sprintf("backup/%s-%s-%s", currentBranch, dateSuffix, timeSuffix)

	fmt.Printf("\n%s Current branch: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render(currentBranch))

	// Check if exists
	err := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+backupName).Run()
	if err == nil {
		fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Backup branch '%s' already exists.", backupName)))
		fmt.Printf("%s %s\n", ui.InfoMessage(""), "Please wait at least 1 minute before creating another backup.")
		os.Exit(1)
	}

	fmt.Printf("%s Creating backup branch: %s...\n", ui.HourglassIcon, ui.GreenStyle().Render(backupName))

	// Create
	exec.Command("git", "branch", backupName).Run()

	// Push
	hasRemote := exec.Command("git", "remote").Run() == nil
	if hasRemote {
		out, _ := exec.Command("git", "remote").Output()
		if strings.Contains(string(out), "origin") {
			fmt.Printf("%s Pushing to origin...\n", ui.HourglassIcon)
			exec.Command("git", "push", "origin", backupName).Run()
			fmt.Printf("%s Backup created and pushed to origin: %s\n", ui.CheckIcon, ui.GreenStyle().Render(backupName))
		} else {
			fmt.Printf("%s Backup created locally: %s\n", ui.CheckIcon, ui.GreenStyle().Render(backupName))
		}
	} else {
		fmt.Printf("%s Backup created locally: %s\n", ui.CheckIcon, ui.GreenStyle().Render(backupName))
	}

	// Cleanup
	manageVersions(currentBranch)

	fmt.Printf("\n%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("Backup workflow completed!"))
}

type backupInfo struct {
	Name string
	Key  string
}

func getMatchedBackups(branch string) []backupInfo {
	out, _ := exec.Command("git", "branch", "--list", "backup/"+branch+"-*", "--format=%(refname:short)").Output()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Regex: backup/branch-name-DD-MM-YYYY-HHh-MM
	re := regexp.MustCompile(`backup/(.*)-([0-9]{2})-([0-9]{2})-([0-9]{4})-([0-9]{2})h-([0-9]{2})`)

	var matched []backupInfo
	for _, b := range lines {
		if b == "" {
			continue
		}
		res := re.FindStringSubmatch(b)
		if len(res) == 7 {
			if res[1] == branch {
				dd, mm, yyyy, hh, min := res[2], res[3], res[4], res[5], res[6]
				key := yyyy + mm + dd + hh + min
				matched = append(matched, backupInfo{Name: b, Key: key})
			}
		}
	}

	// Sort by Key desc
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Key > matched[j].Key
	})

	return matched
}

func manageVersions(currentBranch string) {
	backups := getMatchedBackups(currentBranch)
	count := len(backups)

	if count > 3 {
		fmt.Printf("\n%s Found %s backups for %s.\n", ui.WarningIcon, ui.YellowStyle().Render(fmt.Sprintf("%d", count)), ui.CyanStyle().Render(currentBranch))
		fmt.Printf("%s Keeping the 3 latest versions. The following old backups can be deleted:\n", ui.InfoMessage(""))

		oldBackups := backups[3:]
		fmt.Println("")
		for _, b := range oldBackups {
			fmt.Printf("  %s %s\n", ui.RedStyle().Render("-"), b.Name)
		}
		fmt.Println("")

		if ui.GumConfirm(fmt.Sprintf("Do you want to delete these %d old backups (locally and on origin)?", len(oldBackups))) {
			for _, b := range oldBackups {
				exec.Command("git", "branch", "-D", b.Name).Run()
				// Try to delete from origin
				out, _ := exec.Command("git", "remote").Output()
				if strings.Contains(string(out), "origin") {
					exec.Command("git", "push", "origin", "--delete", b.Name).Run()
				}
				fmt.Printf("%s Deleted: %s\n", ui.CheckIcon, ui.RedStyle().Render(b.Name))
			}
		} else {
			fmt.Printf("%s Deletion skipped.\n", ui.InfoMessage(""))
		}
	}
}
