//go:build dev

package debug

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// PushRcCmd represents the push-rc command
var PushRcCmd = &cobra.Command{
	Use:   "push-rc",
	Short: "Automatically calculates, tags, and pushes a release candidate version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Resolving current release candidate tag state...")

		// 1. Get the latest tag
		gitDescCmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
		latestTagBytes, err := gitDescCmd.Output()
		latestTag := "v1.0.2" // Default fallback matching current state
		if err == nil {
			latestTag = strings.TrimSpace(string(latestTagBytes))
		}
		fmt.Printf("🔍 Detected current highest tag: %s\n", latestTag)

		// 2. Parse tag stable vs non-stable
		// Stable format: e.g. v1.2.3
		// Non-stable format: e.g. v1.2.3-rc1
		rcRegex := regexp.MustCompile(`^v(\d+\.\d+\.\d+)-rc(\d+)$`)
		stableRegex := regexp.MustCompile(`^v(\d+\.\d+\.\d+)$`)

		var newRcTag string
		if rcRegex.MatchString(latestTag) {
			matches := rcRegex.FindStringSubmatch(latestTag)
			version := matches[1]
			rcNum, _ := strconv.Atoi(matches[2])
			newRcTag = fmt.Sprintf("v%s-rc%d", version, rcNum+1)
		} else if stableRegex.MatchString(latestTag) {
			matches := stableRegex.FindStringSubmatch(latestTag)
			version := matches[1]
			newRcTag = fmt.Sprintf("v%s-rc1", version)
		} else {
			fmt.Printf("❌ Error: Unsupported latest tag format: %s. Tag must follow vX.Y.Z semantic patterns.\n", latestTag)
			os.Exit(1)
		}

		fmt.Printf("✨ Calculated next release candidate tag: %s\n", newRcTag)

		// 3. Prompt for tag description using gum input
		fmt.Println("\n💬 Launching gum input for release candidate tag description...")
		gumCmd := exec.Command("gum", "input", "--placeholder", "Enter release candidate description (e.g. Test checkout improvements)...")
		gumCmd.Stdin = os.Stdin
		gumCmd.Stderr = os.Stderr
		descBytes, err := gumCmd.Output()
		if err != nil {
			fmt.Printf("❌ Error running gum input: %v. Please make sure gum is installed in your development environment.\n", err)
			os.Exit(1)
		}

		description := strings.TrimSpace(string(descBytes))
		if description == "" {
			description = fmt.Sprintf("Release Candidate %s", newRcTag)
		}

		// 4. Create annotated tag locally
		fmt.Printf("\n🏷 Tagging version %s locally...\n", newRcTag)
		tagMsg := fmt.Sprintf("%s - %s", newRcTag, description)
		gitTagCmd := exec.Command("git", "tag", "-a", newRcTag, "-m", tagMsg)
		tagOut, err := gitTagCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ Error creating Git tag: %s\n", string(tagOut))
			os.Exit(1)
		}

		fmt.Println("✅ Successfully created tag locally!")
		fmt.Println("==========================================================================")
		fmt.Printf("🎉 Release Candidate Tag: %s\n", newRcTag)
		fmt.Printf("💬 Description: %s\n\n", description)
		fmt.Println("To push this release candidate to GitHub, run:")
		fmt.Printf("👉 git push origin %s\n", newRcTag)
		fmt.Println("==========================================================================")
	},
}
