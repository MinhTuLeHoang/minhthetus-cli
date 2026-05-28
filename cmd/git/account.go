package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

type Account struct {
	Title string `json:"title"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	manageMode bool
	configDir  = filepath.Join(os.Getenv("HOME"), ".minhthetus-cli")
	configFile = filepath.Join(configDir, "git-accounts.json")
)

var AccountCmd = &cobra.Command{
	Use:   "account",
	Short: "Managed Git identities. Quickly switch between accounts or manage your saved list.",
	Long:  "Managed Git identities. Quickly switch between accounts or manage your saved list.",
	Example: `minhthetus-cli git account --manage
minhthetus-cli git account`,
	Annotations: map[string]string{
		"title": "Git Account Manager",
	},
	Run: func(cmd *cobra.Command, args []string) {
		ensureConfig()

		if manageMode {
			for {
				ui.ClearScreen()
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
				var options []string
				for _, acc := range accounts {
					options = append(options, fmt.Sprintf("%s ( %s <%s> )", acc.Title, acc.Name, acc.Email))
				}
				addOpt := "➕ Add New Account"
				quitOpt := "🚪 Quit"
				options = append(options, addOpt, quitOpt)

				choice := ui.GumFilter(options, "Select identity to apply...")
				if choice == "" || choice == quitOpt {
					return
				} else if choice == addOpt {
					createNewAccount()
				} else {
					title := strings.Split(choice, " (")[0]
					var selected Account
					for _, acc := range accounts {
						if acc.Title == title {
							selected = acc
							break
						}
					}

					if selected.Email != "" {
						exec.Command("git", "config", "user.email", selected.Email).Run()
						exec.Command("git", "config", "user.name", selected.Name).Run()
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

func ensureConfig() {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		os.WriteFile(configFile, []byte("[]"), 0644)
	}
}

func readAccounts() []Account {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil
	}
	var accounts []Account
	json.Unmarshal(data, &accounts)
	return accounts
}

func writeAccounts(accounts []Account) {
	data, _ := json.MarshalIndent(accounts, "", "  ")
	os.WriteFile(configFile, data, 0644)
}

func detectIdentity() {
	levels := []string{"local", "global", "system"}
	var foundEmail, foundName string
	var emailLevel, nameLevel string

	for _, level := range levels {
		if foundEmail == "" {
			out, _ := exec.Command("git", "config", "--"+level, "user.email").Output()
			val := strings.TrimSpace(string(out))
			if val != "" {
				foundEmail = val
				emailLevel = level
			}
		}
		if foundName == "" {
			out, _ := exec.Command("git", "config", "--"+level, "user.name").Output()
			val := strings.TrimSpace(string(out))
			if val != "" {
				foundName = val
				nameLevel = level
			}
		}
	}

	fmt.Printf("%s %s\n", ui.TagIcon, ui.BoldStyle.Render("Current Identity Detection:"))
	if foundEmail != "" {
		fmt.Printf("  Email: %s %-30s (from %s)\n", ui.CyanStyle().Render(""), fmt.Sprintf("%-30s", foundEmail), ui.YellowStyle().Render(emailLevel))
	} else {
		fmt.Printf("  Email: %s\n", ui.RedStyle().Render("Not set"))
	}

	if foundName != "" {
		fmt.Printf("  Name:  %s %-30s (from %s)\n", ui.CyanStyle().Render(""), fmt.Sprintf("%-30s", foundName), ui.YellowStyle().Render(nameLevel))
	} else {
		fmt.Printf("  Name:  %s\n", ui.RedStyle().Render("Not set"))
	}
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
		accounts = append(accounts, Account{Title: title, Name: name, Email: email})
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

	var options []string
	for _, acc := range accounts {
		options = append(options, fmt.Sprintf("%s ( %s <%s> )", acc.Title, acc.Name, acc.Email))
	}

	choice := ui.GumFilter(options, "Select account to DELETE...")
	if choice == "" {
		return
	}

	title := strings.Split(choice, " (")[0]
	if ui.GumConfirm(fmt.Sprintf("Are you sure you want to delete '%s'?", title)) {
		var newAccounts []Account
		for _, acc := range accounts {
			if acc.Title != title {
				newAccounts = append(newAccounts, acc)
			}
		}
		writeAccounts(newAccounts)
		fmt.Printf("%s %s\n", ui.SuccessMessage(""), "Account deleted.")
	}
}

