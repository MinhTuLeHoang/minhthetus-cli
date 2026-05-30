package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

type Repository struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
}



var ListRepoCmd = &cobra.Command{
	Use:   "list-repo",
	Short: "Interactively view, add, or remove repositories from the tracking list.",
	Long:  "Interactively manage and list tracked repositories used for bulk scans.",
	Example: `minhthetus-cli git list-repo`,
	Annotations: map[string]string{
		"title": "Repository List Manager",
	},
	Run: func(cmd *cobra.Command, args []string) {
		for {
			ui.ClearScreen()

			repos := readRepos()
			var choice string

			if len(repos) == 0 {
				fmt.Printf("%s %s\n", ui.InfoMessage(""), ui.YellowStyle().Render("No repositories are currently being tracked."))
				action := ui.GumChoose("Add New", "Quit")
				if action == "Quit" || action == "" {
					return
				}
				choice = "➕ Add New"
			} else {
				var options []string
				for _, r := range repos {
					options = append(options, fmt.Sprintf("%s | %s", r.Name, r.Path))
				}
				addOpt := "➕ Add New"
				quitOpt := "🚪 Quit"
				options = append(options, addOpt, quitOpt)

				choice = ui.GumFilter(options, "Select repository to manage...")
				if choice == "" || choice == quitOpt {
					return
				}
			}

			if choice == "➕ Add New" {
				addNewRepo()
				continue
			}

			// Extract path (everything after ' | ')
			parts := strings.Split(choice, " | ")
			repoPath := parts[len(parts)-1]

			// Sub-menu
			ui.ClearScreen()
			fmt.Printf("%s %s\n", ui.TagIcon, ui.BoldStyle.Render("Repository Details:"))
			fmt.Println("")

			var selected Repository
			for _, r := range repos {
				if r.Path == repoPath {
					selected = r
					break
				}
			}

			fmt.Printf("  %s        %s\n", ui.BoldStyle.Render("Name:"), selected.Name)
			fmt.Printf("  %s %s\n", ui.BoldStyle.Render("Description:"), selected.Description)
			fmt.Printf("  %s        %s\n", ui.BoldStyle.Render("Path:"), selected.Path)
			fmt.Println("")

			subAction := ui.GumChoose("Delete", "Back")
			if subAction == "Delete" {
				if ui.GumConfirm("Are you sure you want to untrack this repository?") {
					exec.Command("minhthetus-cli", "repo-untrack", repoPath).Run()
					fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("Repository removed from tracking list."))
					ui.GumInput("Press Enter to continue...", "")
				}
			}
		}
	},
}

func readRepos() []Repository {
	var repos []Repository
	_ = config.ReadFile("list-repo.json", &repos)
	return repos
}

func addNewRepo() {
	fmt.Printf("\n%s %s\n", ui.TagIcon, ui.BoldStyle.Render("Registering New Repository"))
	path := ui.GumInput("Enter absolute path to repository (e.g. /Users/me/project)...", "")
	if path == "" {
		return
	}

	// Check directory
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		fmt.Printf("%s %s %s\n", ui.ErrorIcon, ui.RedStyle().Render("Directory does not exist:"), path)
		ui.GumInput("Press Enter to continue...", "")
		return
	}

	absPath, _ := filepath.Abs(path)
	exec.Command("minhthetus-cli", "repo-track", absPath).Run()
	fmt.Printf("%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("Repository added."))
	ui.GumInput("Press Enter to continue...", "")
}
