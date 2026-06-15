package git

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	patchInc bool
	minorInc bool
	majorInc bool
	message  string
)

var TagDevStgCmd = &cobra.Command{
	Use:   "tag-dev-stg",
	Short: "Automatically calculates the next version, creates stg-v* and qc-v* tags, and pushes to origin.",
	Long:  "Automatically calculates the next version based on existing stg/qc tags, creates new annotated tags on the CURRENT branch, and pushes to origin.",
	Example: `minhthetus-cli git tag-dev-stg -P -m "Hotfix for production"
minhthetus-cli git tag-dev-stg`,
	Args:  cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Git Tag Dev/Stg",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-P", "--patch", "-N", "--minor", "-M", "--major", "-m", "--message", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s Starting Super Tag process...\n", ui.InfoMessage(""))

		// 1. Fetch tags
		fmt.Printf("%s Fetching latest tags from git...\n", ui.CyanStyle().Render(ui.InfoIcon))
		stgTag := getLatestTag("stg-v*")
		qcTag := getLatestTag("qc-v*")

		stgVer := extractVersion(stgTag)
		qcVer := extractVersion(qcTag)

		fmt.Println("")
		fmt.Printf("%s Latest STG tag: %s (Version: %s)\n", ui.TagIcon, ui.YellowStyle().Render(orNone(stgTag)), orZero(stgVer))
		fmt.Printf("%s Latest QC tag:  %s (Version: %s)\n", ui.TagIcon, ui.YellowStyle().Render(orNone(qcTag)), orZero(qcVer))

		baseVer := maxVersion(stgVer, qcVer)
		fmt.Printf("%s Base version identified: %s\n", ui.InfoMessage(""), ui.GreenStyle().Render(baseVer))

		// 2. Increment logic
		incType := "minor"
		if patchInc {
			incType = "patch"
		} else if majorInc {
			incType = "major"
		}

		newVer := incrementVersion(baseVer, incType)
		fmt.Println("")
		fmt.Printf("%s Incrementing version (%s)...\n", ui.HammerIcon, incType)
		fmt.Printf("%s Final version to be used: %s\n", ui.RocketIcon, ui.PurpleStyle().Render(newVer))

		// 3. Message
		if message == "" {
			message = ui.GumInput("Enter tag message (e.g. Release v"+newVer+")", "")
		}
		if message == "" {
			message = "Release v" + newVer
		}

		qcTagName := "qc-v" + newVer
		stgTagName := "stg-v" + newVer

		// 4. Branch check
		out, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		currentBranch := strings.TrimSpace(string(out))
		fmt.Printf("\n%s Current branch: %s\n", ui.InfoMessage(""), ui.YellowStyle().Render(currentBranch))

		// 5. Create tags
		fmt.Printf("\n%s Creating tags on %s branch...\n", ui.CyanStyle().Render(ui.InfoIcon), ui.YellowStyle().Render(currentBranch))

		if err := createTag(qcTagName, currentBranch, message); err != nil {
			fmt.Printf("%s Failed to create tag %s\n", ui.ErrorMessage(""), qcTagName)
			os.Exit(1)
		}
		fmt.Printf("%s Created tag: %s\n", ui.SuccessMessage(""), ui.GreenStyle().Render(qcTagName))

		if err := createTag(stgTagName, currentBranch, message); err != nil {
			fmt.Printf("%s Failed to create tag %s\n", ui.ErrorMessage(""), stgTagName)
			os.Exit(1)
		}
		fmt.Printf("%s Created tag: %s\n", ui.SuccessMessage(""), ui.GreenStyle().Render(stgTagName))

		// 6. Push
		fmt.Printf("\n%s Pushing tags to origin...\n", ui.CyanStyle().Render(ui.InfoIcon))
		err := exec.Command("git", "push", "origin", qcTagName, stgTagName).Run()
		if err != nil {
			fmt.Printf("%s Failed to push tags to origin.\n", ui.ErrorMessage(""))
			os.Exit(1)
		}

		fmt.Printf("\n%s Successfully created and pushed both tags for version %s!\n", ui.SuccessMessage(""), newVer)
		fmt.Printf("%s %s %s\n", ui.BlueStyle().Render(ui.TagIcon), "QC Tag: ", ui.YellowStyle().Render(qcTagName))
		fmt.Printf("%s %s %s\n", ui.BlueStyle().Render(ui.TagIcon), "STG Tag:", ui.YellowStyle().Render(stgTagName))
	},
}

func init() {
	TagDevStgCmd.Flags().BoolVarP(&patchInc, "patch", "P", false, "Increment the patch version (e.g. 1.0.0 -> 1.0.1)")
	TagDevStgCmd.Flags().BoolVarP(&minorInc, "minor", "N", false, "Increment the minor version (e.g. 1.0.0 -> 1.1.0) [Default]")
	TagDevStgCmd.Flags().BoolVarP(&majorInc, "major", "M", false, "Increment the major version (e.g. 1.0.0 -> 2.0.0)")
	TagDevStgCmd.Flags().StringVarP(&message, "message", "m", "", "Provide a custom tag message")
	
	TagDevStgCmd.RegisterFlagCompletionFunc("message", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		templates := []string{
			"dev release: ",
			"staging release: ",
			"hotfix release: ",
			"v",
		}
		return templates, cobra.ShellCompDirectiveNoFileComp
	})
}

// compareVersions compares two semantic version strings numerically.
// Returns < 0 if v1 < v2, 0 if v1 == v2, > 0 if v1 > v2.
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) || i < len(parts2); i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}
		if n1 != n2 {
			return n1 - n2
		}
	}
	return 0
}

func getLatestTag(pattern string) string {
	out, _ := exec.Command("git", "tag", "-l", pattern).Output()
	tags := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(tags) == 0 || (len(tags) == 1 && tags[0] == "") {
		return ""
	}
	sort.Slice(tags, func(i, j int) bool {
		vi := extractVersion(tags[i])
		vj := extractVersion(tags[j])
		return compareVersions(vi, vj) < 0
	})
	return tags[len(tags)-1]
}


func extractVersion(tag string) string {
	re := regexp.MustCompile(`[0-9.]+`)
	return re.FindString(tag)
}

func orNone(s string) string {
	if s == "" {
		return "None"
	}
	return s
}

func orZero(s string) string {
	if s == "" {
		return "0.0.0"
	}
	return s
}

func maxVersion(v1, v2 string) string {
	if v1 == "" {
		if v2 == "" {
			return "0.0.0"
		}
		return v2
	}
	if v2 == "" {
		return v1
	}
	// Simple comparison
	if v1 >= v2 {
		return v1
	}
	return v2
}

func incrementVersion(v, incType string) string {
	parts := strings.Split(v, ".")
	major, minor, patch := 0, 0, 0
	if len(parts) > 0 {
		fmt.Sscanf(parts[0], "%d", &major)
	}
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &minor)
	}
	if len(parts) > 2 {
		fmt.Sscanf(parts[2], "%d", &patch)
	}

	switch incType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func createTag(name, branch, msg string) error {
	return exec.Command("git", "tag", "-a", name, branch, "-m", msg).Run()
}
