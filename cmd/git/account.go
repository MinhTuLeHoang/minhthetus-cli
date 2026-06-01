package git

import (
	"fmt"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/git"
	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type AccountModel struct {
	Title string `json:"title"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	manageMode bool
)

var AccountCmd = &cobra.Command{
	Use:   "account",
	Short: "Managed Git identities. Quickly switch between accounts or manage your saved list.",
	Long:  "Managed Git identities. Quickly switch between accounts or manage your saved list.",
	Example: `minhthetus-cli git account --manage
minhthetus-cli git account`,
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Git Account Manager",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-m", "--manage", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {

		if manageMode {
			for {
				// ui.ClearScreen()
				detectIdentity()
				listAccounts()

				action := ui.GumChoose("Create New", "Delete", "Quit")
				switch action {
				case "Create New":
					createNewAccount()
				case "Delete":
					deleteAccount()
				case "Quit", "":
					return
				}

				fmt.Println("")
				ui.GumInput("Press Enter to continue...", "")
			}
		} else {
			detectIdentity()
			fmt.Println("")

			if !ui.GumConfirm("Switch account for this repository?") {
				return
			}

			accounts := readAccounts()
			if len(accounts) == 0 {
				createNewAccount()
			} else {
				maxTitle := 0
				maxName := 0
				maxEmailCol := 0
				for _, acc := range accounts {
					if len(acc.Title) > maxTitle {
						maxTitle = len(acc.Title)
					}
					if len(acc.Name) > maxName {
						maxName = len(acc.Name)
					}
					emailColLen := len(acc.Email) + 2
					if emailColLen > maxEmailCol {
						maxEmailCol = emailColLen
					}
				}

				var options []string
				for _, acc := range accounts {
					emailCol := fmt.Sprintf("<%s>", acc.Email)
					options = append(options, fmt.Sprintf("%-*s ( %-*s  %-*s )", maxTitle, acc.Title, maxName, acc.Name, maxEmailCol, emailCol))
				}
				addOpt := "➕ Add New Account"
				quitOpt := "🚪 Quit"
				options = append(options, addOpt, quitOpt)

				choice := ui.GumFilter(options, "Select identity to apply...")
				switch choice {
				case "", quitOpt:
					return
				case addOpt:
					createNewAccount()
				default:
					title := strings.TrimSpace(strings.Split(choice, " (")[0])
					var selected AccountModel
					for _, acc := range accounts {
						if acc.Title == title {
							selected = acc
							break
						}
					}

					if selected.Email != "" {
						if _, err := git.Run("config", "user.email", selected.Email); err != nil {
							fmt.Printf("%s Error: %s. Are you inside a Git repository?\n", ui.ErrorMessage(""), err)
							return
						}
						if _, err := git.Run("config", "user.name", selected.Name); err != nil {
							fmt.Printf("%s Error: %s\n", ui.ErrorMessage(""), err)
							return
						}
						fmt.Printf("%s Applied locally: %s <%s>\n", ui.SuccessMessage(""), selected.Name, selected.Email)
					}
				}
			}
		}
	},
}

func init() {
	AccountCmd.Flags().BoolVarP(&manageMode, "manage", "m", false, "Enter management mode (list, create, delete accounts)")
}

func readAccounts() []AccountModel {
	var accounts []AccountModel
	_ = config.ReadFile("git-accounts.json", &accounts)
	return accounts
}

func writeAccounts(accounts []AccountModel) {
	_ = config.WriteFile("git-accounts.json", accounts)
}

func detectIdentity() {
	localName, _ := git.Run("config", "--local", "user.name")
	localEmail, _ := git.Run("config", "--local", "user.email")

	globalName, _ := git.Run("config", "--global", "user.name")
	globalEmail, _ := git.Run("config", "--global", "user.email")

	fmt.Printf("%s %s\n", ui.TagIcon, ui.BoldStyle.Render("Current Identity Detection:"))

	if localName == "" && localEmail == "" && globalName == "" && globalEmail == "" {
		fmt.Printf("  %s\n", ui.WarningMessage("No Git identity configured on this repository or globally."))
		return
	}

	// Define display values (fallback to "not set")
	displayLocalName := localName
	if displayLocalName == "" {
		displayLocalName = "not set"
	}
	displayLocalEmail := localEmail
	if displayLocalEmail == "" {
		displayLocalEmail = "not set"
	}
	displayGlobalName := globalName
	if displayGlobalName == "" {
		displayGlobalName = "not set"
	}
	displayGlobalEmail := globalEmail
	if displayGlobalEmail == "" {
		displayGlobalEmail = "not set"
	}

	// Styles (no bold)
	labelSt := lipgloss.NewStyle()
	valueSt := ui.CyanStyle()
	levelSt := ui.YellowStyle()

	// Compute max value length for aligned columns
	maxValLen := 30
	for _, val := range []string{displayLocalName, displayLocalEmail, displayGlobalName, displayGlobalEmail} {
		if len(val) > maxValLen {
			maxValLen = len(val)
		}
	}

	// 1. Print Local Config
	var localNameStr string
	if localName != "" {
		localNameStr = valueSt.Render(fmt.Sprintf("%-*s", maxValLen, localName))
	} else {
		localNameStr = ui.RedStyle().Render(fmt.Sprintf("%-*s", maxValLen, "not set"))
	}
	fmt.Printf("  %s  %s (from %s)\n", labelSt.Render("Name:"), localNameStr, levelSt.Render("local"))

	var localEmailStr string
	if localEmail != "" {
		localEmailStr = valueSt.Render(fmt.Sprintf("%-*s", maxValLen, localEmail))
	} else {
		localEmailStr = ui.RedStyle().Render(fmt.Sprintf("%-*s", maxValLen, "not set"))
	}
	fmt.Printf("  %s %s (from %s)\n", labelSt.Render("Email:"), localEmailStr, levelSt.Render("local"))

	// 2. Print Global Config
	var globalNameStr string
	if globalName != "" {
		globalNameStr = valueSt.Render(fmt.Sprintf("%-*s", maxValLen, globalName))
	} else {
		globalNameStr = ui.RedStyle().Render(fmt.Sprintf("%-*s", maxValLen, "not set"))
	}
	fmt.Printf("  %s  %s (from %s)\n", labelSt.Render("Name:"), globalNameStr, levelSt.Render("global"))

	var globalEmailStr string
	if globalEmail != "" {
		globalEmailStr = valueSt.Render(fmt.Sprintf("%-*s", maxValLen, globalEmail))
	} else {
		globalEmailStr = ui.RedStyle().Render(fmt.Sprintf("%-*s", maxValLen, "not set"))
	}
	fmt.Printf("  %s %s (from %s)\n", labelSt.Render("Email:"), globalEmailStr, levelSt.Render("global"))
}

func listAccounts() {
	fmt.Printf("\n%s %s\n", ui.TagIcon, ui.BoldStyle.Render("Saved Accounts:"))
	accounts := readAccounts()
	if len(accounts) == 0 {
		fmt.Printf("  %s\n", ui.YellowStyle().Render("No accounts saved."))
	} else {
		// Simple table layout
		fmt.Printf("  %-15s %-20s %-30s\n", "TITLE", "NAME", "EMAIL")
		fmt.Printf("  %-15s %-20s %-30s\n", "-----", "----", "-----")
		for _, acc := range accounts {
			fmt.Printf("  %-15s %-20s %-30s\n", acc.Title, acc.Name, acc.Email)
		}
	}
	fmt.Println("")
}

func createNewAccount() {
	fmt.Printf("\n%s %s\n", ui.TagIcon, ui.BoldStyle.Render("Registering New Account"))
	title := ui.GumInput("Title (e.g. Work)", "")
	if title == "" {
		return
	}

	name := ui.GumInput("User Name", "")
	email := ui.GumInput("User Email", "")

	if name != "" && email != "" {
		accounts := readAccounts()
		accounts = append(accounts, AccountModel{Title: title, Name: name, Email: email})
		writeAccounts(accounts)
		fmt.Printf("%s %s\n", ui.SuccessMessage(""), "Account saved.")
	} else {
		fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Invalid input. Required all fields.")
	}
}

func deleteAccount() {
	accounts := readAccounts()
	if len(accounts) == 0 {
		fmt.Printf("%s %s\n", ui.WarningMessage(""), "No accounts to delete.")
		return
	}

	maxTitle := 0
	maxName := 0
	maxEmailCol := 0
	for _, acc := range accounts {
		if len(acc.Title) > maxTitle {
			maxTitle = len(acc.Title)
		}
		if len(acc.Name) > maxName {
			maxName = len(acc.Name)
		}
		emailColLen := len(acc.Email) + 2
		if emailColLen > maxEmailCol {
			maxEmailCol = emailColLen
		}
	}

	var options []string
	for _, acc := range accounts {
		emailCol := fmt.Sprintf("<%s>", acc.Email)
		options = append(options, fmt.Sprintf("%-*s ( %-*s  %-*s )", maxTitle, acc.Title, maxName, acc.Name, maxEmailCol, emailCol))
	}

	choice := ui.GumFilter(options, "Select account to DELETE...")
	if choice == "" {
		return
	}

	title := strings.TrimSpace(strings.Split(choice, " (")[0])
	if ui.GumConfirm(fmt.Sprintf("Are you sure you want to delete '%s'?", title)) {
		var newAccounts []AccountModel
		for _, acc := range accounts {
			if acc.Title != title {
				newAccounts = append(newAccounts, acc)
			}
		}
		writeAccounts(newAccounts)
		fmt.Printf("%s %s\n", ui.SuccessMessage(""), "Account deleted.")
	}
}

