package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.Color("212")).
			MarginLeft(1)
	
	usageStyle = lipgloss.NewStyle().Bold(true)
	cmdStyle   = lipgloss.NewStyle().Foreground(Cyan)
	dirStyle   = lipgloss.NewStyle().Foreground(Blue).Bold(true)
	coreStyle  = lipgloss.NewStyle().Foreground(Green)
	grayStyle  = lipgloss.NewStyle().Foreground(Gray)
)

// SetCustomHelp configures a custom tree-based help for the given root command.
func SetCustomHelp(root *cobra.Command) {
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd.HasSubCommands() {
			fmt.Print(Splash())
		}

		// Usage / Header
		if cmd.HasSubCommands() {
			fmt.Println("\n" + headerStyle.Render("USAGE"))
			fmt.Printf(" %s %s [args]\n", cmd.CommandPath(), cmdStyle.Render("<command>"))
		} else {
			// Leaf command - Match print-help.sh exactly
			title := cmd.Name()
			if t, ok := cmd.Annotations["title"]; ok {
				title = t
			} else {
				// Capitalize first letter of each word
				words := strings.Split(title, "-")
				for i, w := range words {
					words[i] = strings.ToUpper(w[:1]) + w[1:]
				}
				title = strings.Join(words, " ")
			}
			
			// Match: ${BLUE}${INFO} ${BOLD}${title} Usage Guide:${NC}
			header := lipgloss.NewStyle().Foreground(Blue).Render(fmt.Sprintf("%s %s Usage Guide:", InfoIcon, BoldStyle.Render(title)))
			fmt.Printf("\n%s\n", header)
			fmt.Printf("   %s\n", cmd.CommandPath())
		}

		// Description
		if !cmd.HasSubCommands() {
			fmt.Println("\n" + BoldStyle.Foreground(Yellow).Render("Description:"))
			desc := cmd.Short
			if cmd.Long != "" {
				desc = cmd.Long
			}
			// Indent description by 3 spaces
			for _, line := range strings.Split(strings.TrimSpace(desc), "\n") {
				fmt.Printf("   %s\n", line)
			}
		} else if cmd != root {
			fmt.Printf("\n%s\n", cmd.Short)
		}

		// Flags / Options
		if cmd.HasAvailableLocalFlags() {
			title := "FLAGS"
			indent := "   "
			if !cmd.HasSubCommands() {
				title = "Options:"
				fmt.Println("\n" + BoldStyle.Foreground(Yellow).Render(title))
			} else {
				indent = "  "
				fmt.Println("\n" + usageStyle.Render(title))
			}
			
			// Custom flag formatting to match legacy
			cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if f.Hidden {
					return
				}
				flagName := ""
				if f.Shorthand != "" {
					flagName = fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
				} else {
					flagName = fmt.Sprintf("--%s", f.Name)
				}
				
				if f.Value.Type() != "bool" {
					typeName := f.Value.Type()
					if f.Name == "jira-ticket" {
						typeName = "id"
					}
					flagName += " <" + typeName + ">"
				}
				
				// Legacy uses 22 chars for flag column
				flagCol := cmdStyle.Render(fmt.Sprintf("%-22s", flagName))
				descCol := grayStyle.Render(f.Usage)
				
				fmt.Printf("%s%s %s\n", indent, flagCol, descCol)
			})
		}

		// Example
		if !cmd.HasSubCommands() {
			example := cmd.Example
			if example == "" {
				example = cmd.CommandPath()
			}
			fmt.Println("\n" + BoldStyle.Foreground(Yellow).Render("Example:"))
			for _, line := range strings.Split(strings.TrimSpace(example), "\n") {
				fmt.Printf("   %s\n", line)
			}
			fmt.Println("")
		}

		// Available Commands (Sub-tree)
		if cmd.HasAvailableSubCommands() {
			title := "AVAILABLE COMMANDS"
			if cmd != root {
				title = fmt.Sprintf("SUBCOMMANDS FOR %s", strings.ToUpper(cmd.Name()))
			}
			fmt.Println("\n " + lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Render(title))
			
			// Print tree starting from current command
			printTree(cmd, "", true)
		}

		if cmd == root {
			// Legend
			fmt.Printf("\n %sLegend: %sModules/%s, %sCore Commands%s\n", 
				grayStyle.Render(""), 
				dirStyle.Render(""), 
				grayStyle.Render(""), 
				coreStyle.Render(""), 
				grayStyle.Render(""))

			// Built-in
			fmt.Printf("\n %s\n", usageStyle.Render("Built-in commands:"))
			printBuiltIn(root)
		}
	})
}

func printTree(cmd *cobra.Command, indent string, isRoot bool) {
	commands := cmd.Commands()
	// Filter out built-in commands for the tree
	var filtered []*cobra.Command
	for _, c := range commands {
		if c.Name() == "help" || c.Name() == "completion" || c.Name() == "setup-completion" || c.Name() == "uninstall" {
			continue
		}
		filtered = append(filtered, c)
	}

	for i, c := range filtered {
		isLast := i == len(filtered)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		description := ""
		if c.Short != "" {
			description = grayStyle.Render(" # " + c.Short)
		}

		if c.HasSubCommands() {
			fmt.Printf(" %s%s%s\n", indent, connector, dirStyle.Render(c.Name()+"/"))
			newIndent := indent + "│   "
			if isLast {
				newIndent = indent + "    "
			}
			printTree(c, newIndent, false)
		} else {
			fmt.Printf(" %s%s%s%s\n", indent, connector, coreStyle.Render(c.Name()), description)
		}
	}
}

func printBuiltIn(root *cobra.Command) {
	builtIns := []string{"help", "setup-completion", "uninstall"}
	for _, name := range builtIns {
		for _, c := range root.Commands() {
			if c.Name() == name {
				fmt.Printf("  %s %s\n", lipgloss.NewStyle().Foreground(Cyan).Width(18).Render(c.Name()), c.Short)
			}
		}
	}
}
