package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// fetchLatestVersion queries GitHub for the latest release tag
func fetchLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Method 1: Try fetching the raw version.go from the master branch on GitHub (reliable, no API rate limits)
	req, err := http.NewRequest("GET", "https://raw.githubusercontent.com/MinhTuLeHoang/minhthetus-cli/master/internal/config/version.go", nil)
	if err == nil {
		req.Header.Set("User-Agent", "minhthetus-cli")
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err == nil {
				content := string(bodyBytes)
				// Find Version = "X.Y.Z"
				startIdx := strings.Index(content, "Version = \"")
				if startIdx != -1 {
					content = content[startIdx+11:]
					endIdx := strings.Index(content, "\"")
					if endIdx != -1 {
						return content[:endIdx], nil
					}
				}
			}
		}
	}

	// Method 2: Fallback to GitHub Releases API
	req, err = http.NewRequest("GET", "https://api.github.com/repos/MinhTuLeHoang/minhthetus-cli/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "minhthetus-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.TagName, nil
}

// parseVersion converts a version string like "v1.3.1" or "1.3.1" into [1, 3, 1]
func parseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	ints := make([]int, 0, len(parts))
	for _, p := range parts {
		if idx := strings.Index(p, "-"); idx != -1 {
			p = p[:idx]
		}
		val, err := strconv.Atoi(p)
		if err != nil {
			val = 0
		}
		ints = append(ints, val)
	}
	return ints
}

// isUpToDate returns true if current >= latest
func isUpToDate(current, latest string) bool {
	currParts := parseVersion(current)
	lateParts := parseVersion(latest)

	maxLen := len(currParts)
	if len(lateParts) > maxLen {
		maxLen = len(lateParts)
	}

	for i := 0; i < maxLen; i++ {
		currVal := 0
		if i < len(currParts) {
			currVal = currParts[i]
		}
		lateVal := 0
		if i < len(lateParts) {
			lateVal = lateParts[i]
		}
		if currVal > lateVal {
			return true
		}
		if currVal < lateVal {
			return false
		}
	}
	return true
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the CLI to the latest version",
	Long:  `Detects the installation method (Homebrew, Go, or Manual) and updates the minhthetus-cli binary to the latest version while safely preserving all local configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ui.InfoMessage("All local configurations in ~/.minhthetus-cli/ will be fully preserved."))
		fmt.Printf("%s Checking current installation method...\n", ui.HourglassIcon)

		// 1. Detect executable path
		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Failed to detect executable path: %v", err)))
			return
		}

		method := detectInstallMethod(exePath)
		fmt.Printf("%s Detected %s installation.\n", ui.HourglassIcon, method)

		// 1b. Check if update is needed
		fmt.Printf("%s Checking for updates...\n", ui.HourglassIcon)
		currentVer := getVersionString()
		latestVer, err := fetchLatestVersion()
		if err == nil {
			if isUpToDate(currentVer, latestVer) {
				fmt.Printf("%s minhthetus-cli is already up-to-date (current version: %s, latest version: %s).\n", ui.SuccessMessage(""), currentVer, latestVer)
				return
			}
			fmt.Printf("✨ A new version is available: %s (current: %s)\n", ui.GreenStyle().Render(latestVer), ui.GrayStyle().Render(currentVer))
		} else {
			// If online check fails, and it's Homebrew, we can try 'brew outdated' as a local/fallback check
			if method == Homebrew {
				fmt.Printf("%s Checking local Homebrew packages...\n", ui.HourglassIcon)
				outdatedCmd := exec.Command("brew", "outdated", "minhthetus-cli")
				out, outdatedErr := outdatedCmd.Output()
				if outdatedErr == nil && len(strings.TrimSpace(string(out))) == 0 {
					fmt.Printf("%s minhthetus-cli is already up-to-date via Homebrew.\n", ui.SuccessMessage(""))
					return
				}
			}

			fmt.Printf("%s %s\n", ui.WarningMessage(""), ui.YellowStyle().Render(fmt.Sprintf("Could not retrieve latest version information online: %v", err)))
			proceed, confirmErr := ui.Confirm("Would you like to force the update process anyway?", 0, false)
			if confirmErr != nil || !proceed {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Update cancelled.")
				return
			}
		}

		// 2. Perform update based on method
		switch method {
		case Homebrew:
			confirmMsg := "Would you like to run 'brew upgrade minhthetus-cli' to update?"
			confirmed, err := ui.Confirm(confirmMsg, 0, true)
			if err != nil || !confirmed {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Update cancelled.")
				return
			}
			fmt.Printf("%s Running 'brew update' to refresh formulae...\n", ui.HourglassIcon)
			updateCmd := exec.Command("brew", "update")
			_ = updateCmd.Run() // Attempt to update, ignore errors to still try upgrade

			fmt.Printf("%s Running 'brew upgrade minhthetus-cli'...\n", ui.HourglassIcon)
			brewCmd := exec.Command("brew", "upgrade", "minhthetus-cli")
			brewCmd.Stdout = os.Stdout
			brewCmd.Stderr = os.Stderr
			brewCmd.Stdin = os.Stdin
			if err := brewCmd.Run(); err != nil {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Homebrew update failed: %v", err)))
				fmt.Printf("%s Please run the command manually: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render("brew upgrade minhthetus-cli"))
				return
			}
			fmt.Printf("%s %s\n", ui.SuccessMessage(""), ui.GreenStyle().Render("Successfully updated minhthetus-cli via Homebrew!"))

		case GoInstall:
			confirmMsg := "Would you like to run 'go install' to update to the latest version?"
			confirmed, err := ui.Confirm(confirmMsg, 0, true)
			if err != nil || !confirmed {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), "Update cancelled.")
				return
			}

			fmt.Printf("%s Running 'go install github.com/MinhTuLeHoang/minhthetus-cli@latest'...\n", ui.HourglassIcon)
			goCmd := exec.Command("go", "install", "github.com/MinhTuLeHoang/minhthetus-cli@latest")
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr
			goCmd.Stdin = os.Stdin
			if err := goCmd.Run(); err != nil {
				fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Go install update failed: %v", err)))
				fmt.Printf("%s Please run the command manually: %s\n", ui.InfoMessage(""), ui.CyanStyle().Render("go install github.com/MinhTuLeHoang/minhthetus-cli@latest"))
				return
			}
			fmt.Printf("%s %s\n", ui.SuccessMessage(""), ui.GreenStyle().Render("Successfully updated minhthetus-cli via Go!"))

		case ManualBuild:
			// Check if we are running in the cloned directory by checking for Makefile
			if _, err := os.Stat("Makefile"); err == nil {
				confirmMsg := "Manual/Makefile installation detected. Do you want to pull the latest changes and reinstall now?"
				confirmed, err := ui.Confirm(confirmMsg, 0, true)
				if err == nil && confirmed {
					fmt.Printf("%s Pulling latest changes (git pull)...\n", ui.HourglassIcon)
					gitCmd := exec.Command("git", "pull")
					gitCmd.Stdout = os.Stdout
					gitCmd.Stderr = os.Stderr
					gitCmd.Stdin = os.Stdin
					if err := gitCmd.Run(); err != nil {
						fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Failed to git pull: %v", err)))
						return
					}

					fmt.Printf("%s Recompiling and reinstalling (make install)...\n", ui.HourglassIcon)
					makeCmd := exec.Command("make", "install")
					makeCmd.Stdout = os.Stdout
					makeCmd.Stderr = os.Stderr
					makeCmd.Stdin = os.Stdin
					if err := makeCmd.Run(); err != nil {
						fmt.Printf("%s %s\n", ui.ErrorMessage(""), ui.RedStyle().Render(fmt.Sprintf("Failed to run make install: %v", err)))
						return
					}
					fmt.Printf("%s %s\n", ui.SuccessMessage(""), ui.GreenStyle().Render("Successfully updated minhthetus-cli manually!"))
					return
				}
			}

			fmt.Printf("%s %s\n", ui.InfoMessage(""), "Manual/Makefile installation detected. To update, please navigate to your cloned minhthetus-cli repository and run:")
			fmt.Println(ui.GrayStyle().Render("   git pull && make install"))
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
