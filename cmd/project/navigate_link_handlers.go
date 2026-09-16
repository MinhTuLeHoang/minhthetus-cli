package project

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
)

func handleAdd(groups []*LinkGroup, lines []string) {
	groupOptions := []string{"[New Group]"}
	for i, g := range groups {
		groupOptions = append(groupOptions, fmt.Sprintf("%d. %s", i+1, g.Name))
	}

	selectedGroup, err := ui.Choose("Select a group or create new:", groupOptions)
	if err != nil || selectedGroup == "" {
		return
	}

	var targetGroupName string
	var insertLineIndex int
	var isNewGroup bool

	if selectedGroup == "[New Group]" {
		name, err := ui.Input("Enter new group name:", "")
		if err != nil || name == "" {
			return
		}
		targetGroupName = name
		insertLineIndex = len(lines)
		isNewGroup = true
	} else {
		// Extract index and name
		parts := strings.SplitN(selectedGroup, ". ", 2)
		if len(parts) == 2 {
			targetGroupName = parts[1]
			// Find the last line of the selected group to append the link
			for _, g := range groups {
				if g.Name == targetGroupName {
					insertLineIndex = g.LineIndex + 1
					if len(g.Items) > 0 {
						insertLineIndex = g.Items[len(g.Items)-1].LineIndex + 1
					}
					break
				}
			}
		}
	}

	linkName, err := ui.Input("Enter link name:", "")
	if err != nil || linkName == "" {
		return
	}

	linkURL, err := ui.Input("Enter link URL (must start with https://):", "https://")
	if err != nil || linkURL == "" {
		return
	}
	if !strings.HasPrefix(linkURL, "https://") && !strings.HasPrefix(linkURL, "http://") {
		fmt.Println(ui.ErrorMessage("URL must start with http:// or https://"))
		return
	}

	// formatting
	newLine := fmt.Sprintf("- %s: %s", linkName, linkURL)
	if isNewGroup {
		newLine = fmt.Sprintf("\n## %s\n- [default] %s: %s", targetGroupName, linkName, linkURL)
	}

	// Insert into lines
	if insertLineIndex >= len(lines) {
		if isNewGroup {
			lines = append(lines, strings.Split(newLine, "\n")...)
		} else {
			lines = append(lines, newLine)
		}
	} else {
		lines = append(lines[:insertLineIndex], append([]string{newLine}, lines[insertLineIndex:]...)...)
	}

	err = os.WriteFile(linkMdFile, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Failed to save: " + err.Error()))
		return
	}
	fmt.Println(ui.SuccessMessage("Link added successfully!"))
}

func handleDelete(groups []*LinkGroup, lines []string) {
	if len(groups) == 0 {
		fmt.Println(ui.WarningMessage("No links found."))
		return
	}

	var options []string
	var optionToLineMap = make(map[string]int)

	for _, g := range groups {
		for _, item := range g.Items {
			opt := fmt.Sprintf("[%s] %s: %s", g.Name, item.Name, item.URL)
			options = append(options, opt)
			optionToLineMap[opt] = item.LineIndex
		}
	}

	if len(options) == 0 {
		fmt.Println(ui.WarningMessage("No links found."))
		return
	}

	selected, err := ui.Choose("Select a link to delete:", options)
	if err != nil || selected == "" {
		return
	}

	lineIdx := optionToLineMap[selected]
	lines = append(lines[:lineIdx], lines[lineIdx+1:]...)

	err = os.WriteFile(linkMdFile, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Failed to delete: " + err.Error()))
		return
	}
	fmt.Println(ui.SuccessMessage("Link deleted successfully!"))
}

func handleChangeDefault(groups []*LinkGroup, lines []string) {
	if len(groups) == 0 {
		fmt.Println(ui.WarningMessage("No groups found."))
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

	if len(selectedGroup.Items) == 0 {
		fmt.Println(ui.WarningMessage("No links in this group."))
		return
	}

	var itemNames []string
	var optionToLineMap = make(map[string]int)
	for _, item := range selectedGroup.Items {
		name := item.Name
		if item.IsDefault {
			name = "[default] " + name
		}
		itemNames = append(itemNames, name)
		optionToLineMap[name] = item.LineIndex
	}

	selectedLink, err := ui.Choose("Select the new default link:", itemNames)
	if err != nil || selectedLink == "" {
		return
	}

	// Remove default from all items in this group
	for _, item := range selectedGroup.Items {
		idx := item.LineIndex
		lines[idx] = strings.Replace(lines[idx], "[default] ", "", 1)
	}

	// Add default to selected
	selectedIdx := optionToLineMap[selectedLink]
	// Using regex to replace the "- " with "- [default] "
	re := regexp.MustCompile(`^(\s*-\s*)(.*)`)
	lines[selectedIdx] = re.ReplaceAllString(lines[selectedIdx], "${1}[default] ${2}")

	err = os.WriteFile(linkMdFile, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Println(ui.ErrorMessage("Failed to change default: " + err.Error()))
		return
	}
	fmt.Println(ui.SuccessMessage("Default link updated!"))
}

func handleHelp() {
	helpText := `
How to manually update RELATED_TOOL_LINK.md:

1. The file uses markdown headings (##) to define groups.
2. Inside each group, use bullet points (-) to define links.
3. The format is: - [default] name: https://url
4. If a group only has 1 item, it will automatically navigate to it.
5. If there are multiple items, the one with [default] will be pre-selected.
`
	fmt.Println(ui.InfoMessage(helpText))
}
