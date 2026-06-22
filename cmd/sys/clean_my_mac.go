package sys

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var uninstallFlag bool

// CleanMyMacCmd represents the sys clean-my-mac command
var CleanMyMacCmd = &cobra.Command{
	Use:   "clean-my-mac",
	Short: "Cleans up macOS caches, logs, and temp files using mac-cleaner-cli.",
	Long: `Cleans up macOS system cache, logs, and developer temporary directories.
Ensures a Node.js version of 20 or higher is active (using nvm to switch/install if needed),
then runs the interactive npx mac-cleaner-cli tool.`,
	Annotations: map[string]string{
		"title": "Clean My Mac",
	},
	Run: func(cmd *cobra.Command, args []string) {
		runCleanMyMac()
	},
}

func init() {
	CleanMyMacCmd.Flags().BoolVarP(&uninstallFlag, "uninstall", "u", false, "Uninstall mac-cleaner-cli configurations")
}

func runCleanMyMac() {
	fmt.Printf("%s Checking and configuring environment...\n", ui.HourglassIcon)

	script := `
if [ -z "$NVM_DIR" ]; then
	export NVM_DIR="$HOME/.nvm"
fi

if [ -s "$NVM_DIR/nvm.sh" ]; then
	. "$NVM_DIR/nvm.sh"
fi

get_node_major() {
	if ! type node >/dev/null 2>&1; then
		echo 0
		return
	fi
	local version
	version=$(node -v 2>/dev/null)
	echo "${version#v}" | cut -d. -f1
}

# If nvm is available, try using 'node' alias (the latest installed Node version)
if type nvm >/dev/null 2>&1; then
	nvm use node >/dev/null 2>&1
fi

MAJOR=$(get_node_major)

if [ "$MAJOR" -lt 20 ]; then
	if type nvm >/dev/null 2>&1; then
		echo "Current active Node version ($MAJOR) is less than 20."
		echo "Attempting to install and use the latest Node version via nvm..."
		nvm install node
		nvm use node
	else
		echo "Node version ($MAJOR) is less than 20, and nvm is not found."
		echo "Please install Node.js version 20 or higher."
		exit 1
	fi
fi

MAJOR=$(get_node_major)
if [ "$MAJOR" -lt 20 ]; then
	echo "Failed to activate Node.js >= 20. Current version is $MAJOR."
	exit 1
fi

printf "Active Node version: \033[32m%s\033[0m\n" "$(node -v)"
echo "Executing mac-cleaner-cli..."

if [ "$1" = "uninstall" ]; then
	npx mac-cleaner-cli uninstall
else
	npx mac-cleaner-cli
fi
`

	var arg string
	if uninstallFlag {
		arg = "uninstall"
	}

	execCmd := exec.Command("bash", "-c", script, "--", arg)
	execCmd.Env = os.Environ()
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	if err := execCmd.Run(); err != nil {
		fmt.Printf("\n%s %s\n", ui.ErrorIcon, ui.RedStyle().Render(fmt.Sprintf("Failed to run clean-mac command: %v", err)))
		os.Exit(1)
	}

	fmt.Printf("\n%s %s\n", ui.CheckIcon, ui.GreenStyle().Render("System cleanup task completed successfully."))
}
