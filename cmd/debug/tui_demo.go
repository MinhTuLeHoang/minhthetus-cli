//go:build dev

package debug

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var GumDemoCmd = &cobra.Command{
	Use:   "tui-demo",
	Short: "An interactive demonstration of native TUI elements.",
	Annotations: map[string]string{
		"title": "TUI Demo",
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Stylized Header
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("212")).
			Align(lipgloss.Center).
			Width(50).
			Margin(1, 2).
			Padding(1, 2).
			Render("NATIVE TUI INTERACTION DEMO")
		fmt.Println(header)

		// 2. Interactive Choice
		action := ui.GumChoose("Say Hello", "System Info", "Spin Demo", "Confirm Demo", "Exit")

		switch action {
		case "Say Hello":
			name := ui.GumInput("What is your name?", "")
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
			fmt.Println(style.Render(fmt.Sprintf("Hello, %s! Welcome to the enhanced CLI.", name)))

		case "System Info":
			style := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				Margin(1, 2).
				Padding(1, 2).
				BorderForeground(lipgloss.Color("99"))
			
			osName, _ := exec.Command("uname", "-s").Output()
			kernel, _ := exec.Command("uname", "-r").Output()
			shell := os.Getenv("SHELL")

			content := fmt.Sprintf("OS: %sKernel: %sShell: %s", osName, kernel, shell)
			fmt.Println(style.Render(content))

		case "Spin Demo":
			fmt.Printf("%s Performing background magic...\n", ui.HourglassIcon)
			time.Sleep(2 * time.Second)
			fmt.Println(ui.SuccessMessage("Magic complete!"))

		case "Confirm Demo":
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("(Auto-approving in 2 seconds...)"))
			
			if ui.GumConfirmTimeout("Proceed with the operation?", "2s") {
				fmt.Println(ui.BoldStyle.Foreground(ui.Green).Render("Status: CONFIRMED (or auto-approved)"))
			} else {
				fmt.Println(ui.BoldStyle.Foreground(ui.Red).Render("Status: CANCELLED"))
			}

		case "Exit":
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Goodbye!"))
		}
	},
}
