package web

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
)

func getWebInfo() string {
	// Step 1: Node version
	var nodeVer string
	if _, err := os.Stat(".nvmrc"); err == nil {
		data, _ := os.ReadFile(".nvmrc")
		nodeVer = strings.TrimSpace(string(data))
		fmt.Printf("%s Found .nvmrc: %s\n", ui.GreenStyle().Render(ui.CheckIcon), nodeVer)
		// We can't easily do 'nvm use' in Go for the parent shell without a pipe
	} else {
		fmt.Printf("%s No .nvmrc found.\n", ui.YellowStyle().Render(ui.InfoIcon))
		nvmDir := filepath.Join(os.Getenv("HOME"), ".nvm", "versions", "node")
		if info, err := os.Stat(nvmDir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(nvmDir)
			var versions []string
			for _, entry := range entries {
				if entry.IsDir() {
					versions = append(versions, entry.Name())
				}
			}
			sort.Slice(versions, func(i, j int) bool {
				return versions[i] > versions[j] // Descending
			})

			if len(versions) > 0 {
				selected := ui.GumChoose(versions...)
				if selected != "" {
					nodeVer = selected
					major := strings.Split(nodeVer, ".")[0]
					os.WriteFile(".nvmrc", []byte(major), 0644)
					fmt.Printf("%s Created .nvmrc with %s\n", ui.GreenStyle().Render(ui.CheckIcon), major)
				}
			}
		}
	}

	// Step 2: Package Manager
	var pkgManager string
	var locks []string
	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		locks = append(locks, "pnpm")
	}
	if _, err := os.Stat("package-lock.json"); err == nil {
		locks = append(locks, "npm")
	}
	if _, err := os.Stat("yarn.lock"); err == nil {
		locks = append(locks, "yarn")
	}

	if len(locks) == 0 {
		fmt.Printf("\n%s No lock files found. Please choose a package manager.\n", ui.YellowStyle().Render(ui.WarningIcon))
		pkgManager = ui.GumChoose("pnpm", "npm", "yarn")
	} else if len(locks) == 1 {
		pkgManager = locks[0]
		fmt.Printf("%s Detected package manager: %s\n", ui.GreenStyle().Render(ui.CheckIcon), pkgManager)
	} else {
		fmt.Printf("%s Multiple lock files detected.\n", ui.YellowStyle().Render(ui.InfoIcon))
		pkgManager = ui.GumChoose(locks...)
		for _, l := range locks {
			if l != pkgManager {
				file := ""
				switch l {
				case "pnpm":
					file = "pnpm-lock.yaml"
				case "npm":
					file = "package-lock.json"
				case "yarn":
					file = "yarn.lock"
				}
				if ui.GumConfirm("Delete redundant " + file + "?") {
					os.Remove(file)
					fmt.Printf("%s Deleted %s\n", ui.GreenStyle().Render(ui.CheckIcon), file)
				}
			}
		}
	}

	return pkgManager
}

func trackCurrentRepo() {
	cwd, _ := os.Getwd()
	exec.Command("minhthetus-cli", "repo-track", cwd, "--silent").Start()
}

func showSpinner(title string, command *exec.Cmd) error {
	// Simple spinner implementation using gum if available
	fmt.Printf("%s %s\n", ui.HourglassIcon, title)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func getDuration(start time.Time) string {
	elapsed := time.Since(start)
	return fmt.Sprintf("%.1fs", elapsed.Seconds())
}

// I need to import time in utils.go
