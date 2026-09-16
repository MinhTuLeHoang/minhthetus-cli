package project

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	linkMdFile = "RELATED_TOOL_LINK.md"
	mdTemplate = `# RELATED TOOL LINK MANAGEMENT

<!-- this is version of this md file schema, please don't remove this -->
Version: 1

This file contains quick navigation links for the project (such as CI/CD pipelines, remote repositories, dashboards, and environments). You can use the ` + "`minhthetus-cli project navigate-link`" + ` to search, add, remove, and quickly launch these links directly from your terminal.

## Git Repo
- [default] gitlab: https://
`
)

type LinkItem struct {
	Name      string
	URL       string
	IsDefault bool
	LineIndex int // to update the file later
}

type LinkGroup struct {
	Name      string
	Items     []*LinkItem
	LineIndex int // to update the file later
}

var NavigateLinkCmd = &cobra.Command{
	Use:   "navigate-link",
	Short: "Navigate to project-related links defined in RELATED_TOOL_LINK.md",
	Run: func(cmd *cobra.Command, args []string) {
		runNavigateLink()
	},
}

func runNavigateLink() {
	// Check if in git repository
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		fmt.Println(ui.ErrorMessage("This command must be run inside a git repository."))
		return
	}

	if _, err := os.Stat(linkMdFile); os.IsNotExist(err) {
		fmt.Println(ui.WarningMessage(fmt.Sprintf("%s not found in the current directory.", linkMdFile)))
		fmt.Println(ui.InfoMessage("This file helps manage project-related links (e.g., CI pipelines, logs, dashboard, ...)."))

		confirm, err := ui.Confirm("Do you want to create a default template?", 30*time.Second, true)
		if err != nil || !confirm {
			fmt.Println(ui.InfoMessage("Canceled."))
			return
		}

		content := generateMdTemplate()
		if err := os.WriteFile(linkMdFile, []byte(content), 0644); err != nil {
			fmt.Println(ui.ErrorMessage("Failed to create file: " + err.Error()))
			return
		}
		fmt.Println(ui.SuccessMessage("Created " + linkMdFile))
	}

	for {
		action, err := ui.Choose("Select action:", []string{
			"navigate",
			"add new",
			"delete",
			"change default in group",
			"help",
			"exit",
		})
		if err != nil || action == "exit" || action == "" {
			break
		}

		groups, lines, err := parseLinkFile(linkMdFile)
		if err != nil {
			fmt.Println(ui.ErrorMessage("Failed to read file: " + err.Error()))
			return
		}

		switch action {
		case "navigate":
			handleSearch(groups)
		case "add new":
			handleAdd(groups, lines)
		case "delete":
			handleDelete(groups, lines)
		case "change default in group":
			handleChangeDefault(groups, lines)
		case "help":
			handleHelp()
		}
	}
}

func parseLinkFile(filename string) ([]*LinkGroup, []string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	lines := strings.Split(string(data), "\n")

	var groups []*LinkGroup
	var currentGroup *LinkGroup

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			currentGroup = &LinkGroup{
				Name:      strings.TrimSpace(trimmed[3:]),
				LineIndex: i,
			}
			groups = append(groups, currentGroup)
		} else if strings.HasPrefix(trimmed, "- ") && currentGroup != nil {
			isDefault := strings.Contains(trimmed, "[default]")
			cleanLine := strings.TrimPrefix(trimmed, "- ")
			cleanLine = strings.Replace(cleanLine, "[default]", "", 1)
			cleanLine = strings.TrimSpace(cleanLine)

			parts := strings.SplitN(cleanLine, ": http", 2)
			if len(parts) == 2 {
				currentGroup.Items = append(currentGroup.Items, &LinkItem{
					Name:      strings.TrimSpace(parts[0]),
					URL:       "http" + strings.TrimSpace(parts[1]),
					IsDefault: isDefault,
					LineIndex: i,
				})
			}
		}
	}
	return groups, lines, nil
}

func handleSearch(groups []*LinkGroup) {
	if len(groups) == 0 {
		fmt.Println(ui.WarningMessage("No groups found in the file."))
		return
	}

	var groupNames []string
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}

	groupName, err := ui.Choose("Select a group:", groupNames)
	if err != nil || groupName == "" {
		return
	}

	var selectedGroup *LinkGroup
	for _, g := range groups {
		if g.Name == groupName {
			selectedGroup = g
			break
		}
	}

	if selectedGroup == nil || len(selectedGroup.Items) == 0 {
		fmt.Println(ui.WarningMessage("No items in this group."))
		return
	}

	if len(selectedGroup.Items) == 1 {
		openURL(selectedGroup.Items[0].URL)
		return
	}

	// Auto-select default if there's one? Requirement says "select item -> navigate", let's list them.
	// But let's pre-select the default one in the UI.
	defaultIndex := 0
	var itemNames []string
	for i, item := range selectedGroup.Items {
		name := item.Name
		if item.IsDefault {
			name = "[default] " + name
			defaultIndex = i
		}
		itemNames = append(itemNames, name)
	}

	// Currently UI doesn't have ChooseWithDefault exposed if we just use Choose, wait, I can use ui.ChooseWithDefault
	itemName, err := ui.ChooseWithDefault("Select a link:", itemNames, defaultIndex)
	if err != nil || itemName == "" {
		return
	}

	for _, item := range selectedGroup.Items {
		name := item.Name
		if item.IsDefault {
			name = "[default] " + name
		}
		if name == itemName {
			openURL(item.URL)
			return
		}
	}
}

func openURL(url string) {
	fmt.Println(ui.InfoMessage("Opening: " + url))
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println(ui.ErrorMessage("Failed to open browser: " + err.Error()))
	}
}

func generateMdTemplate() string {
	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return mdTemplate
	}
	url := strings.TrimSpace(string(out))
	if url == "" {
		return mdTemplate
	}

	if strings.HasPrefix(url, "git@") {
		url = strings.Replace(url, ":", "/", 1)
		url = strings.Replace(url, "git@", "https://", 1)
	}
	url = strings.TrimSuffix(url, ".git")

	name := "gitlab"
	if strings.Contains(url, "github.com") {
		name = "github"
	}

	return strings.ReplaceAll(mdTemplate, "- [default] gitlab: https://", fmt.Sprintf("- [default] %s: %s", name, url))
}
