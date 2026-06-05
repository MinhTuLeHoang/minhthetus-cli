package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func main() {
	// 1. Define optional flags
	targetVersionFlag := flag.String("version", "", "Release version string (e.g. 1.0.3) to run in automated non-interactive mode")
	var addedFlags stringSliceFlag
	var changedFlags stringSliceFlag
	var fixedFlags stringSliceFlag
	var removedFlags stringSliceFlag

	flag.Var(&addedFlags, "added", "Added changelog entries (can be specified multiple times)")
	flag.Var(&changedFlags, "changed", "Changed changelog entries (can be specified multiple times)")
	flag.Var(&fixedFlags, "fixed", "Fixed changelog entries (can be specified multiple times)")
	flag.Var(&removedFlags, "removed", "Removed changelog entries (can be specified multiple times)")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	// Discover latest stable version first so it is available globally
	latestStable, err := runCmd("go", "run", "scripts/publish/get_version.go")
	if err != nil {
		fmt.Printf("❌ Error retrieving stable version: %s\n", latestStable)
		os.Exit(1)
	}

	// Variables to determine version and release notes
	var nextVersion string
	var changelogSections = make(map[string][]string)
	categories := []string{"Added", "Changed", "Fixed", "Removed"}
	isNonInteractive := *targetVersionFlag != ""

	if isNonInteractive {
		// Non-interactive automated Agent mode
		nextVersion = strings.TrimPrefix(*targetVersionFlag, "v")
		fmt.Printf("🤖 Running in automated non-interactive publishing mode for target version v%s...\n", nextVersion)
		changelogSections["Added"] = addedFlags
		changelogSections["Changed"] = changedFlags
		changelogSections["Fixed"] = fixedFlags
		changelogSections["Removed"] = removedFlags
	} else {
		// Fully interactive human developer mode
		fmt.Println("🚀 Starting Interactive custom publishing checklist (Protected Master Compliant)...")


		// 2. Report stable version discovered earlier
		fmt.Printf("🔍 Detected latest stable tag: v%s\n", latestStable)

		// Parse version components
		vParts := strings.Split(latestStable, ".")
		if len(vParts) != 3 {
			fmt.Printf("❌ Error: Invalid stable version format: %s\n", latestStable)
			os.Exit(1)
		}
		major, _ := strconv.Atoi(vParts[0])
		minor, _ := strconv.Atoi(vParts[1])
		patch, _ := strconv.Atoi(vParts[2])

		// Prompt for bump type
		fmt.Println("\nSelect version bump type:")
		fmt.Printf("  [1] Patch Version: v%d.%d.%d -> v%d.%d.%d\n", major, minor, patch, major, minor, patch+1)
		fmt.Printf("  [2] Minor Version: v%d.%d.%d -> v%d.%d.0\n", major, minor, patch, major, minor+1)
		fmt.Print("Choose option [1 or 2]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "2" {
			nextVersion = fmt.Sprintf("%d.%d.0", major, minor+1)
		} else {
			nextVersion = fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
		}
		fmt.Printf("\n✨ Target Publish Version: v%s\n", nextVersion)

		// Gather changelog changes interactively
		fmt.Println("\n📜 Lets gather the CHANGELOG updates. Enter entries for the following categories:")
		for _, cat := range categories {
			fmt.Printf("\n--- Enter [%s] changes (type an entry and press Enter. Enter empty line to finish) ---\n", cat)
			for {
				fmt.Print("> ")
				entry, _ := reader.ReadString('\n')
				entry = strings.TrimSpace(entry)
				if entry == "" {
					break
				}
				changelogSections[cat] = append(changelogSections[cat], entry)
			}
		}
	}

	// 5. Detect current branch
	currentBranch, err := runCmd("git", "branch", "--show-current")
	if err != nil || currentBranch == "" {
		fmt.Printf("❌ Error detecting current branch: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Using current branch '%s' for publishing...\n", currentBranch)

	// 6. Update Version Constants in internal/config/version.go
	buildDate := time.Now().Format("2006-01-02")
	versionGoPath := "internal/config/version.go"
	versionContent := fmt.Sprintf(`package config

const (
	// Version is the current stable version of minhthetus-cli
	Version = "%s"
	// BuildDate is the release date of the current stable version
	BuildDate = "%s"
)
`, nextVersion, buildDate)

	err = os.WriteFile(versionGoPath, []byte(versionContent), 0644)
	if err != nil {
		fmt.Printf("❌ Error updating version constant file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ central version constants updated in internal/config/version.go.")

	// Update wiki documentation version references
	if err := updateWikiFiles(nextVersion); err != nil {
		fmt.Printf("⚠️ Warning: %v\n", err)
	}

	// 8. Read existing CHANGELOG.md and append the new version section
	changelogPath := "CHANGELOG.md"
	existingChangelog, err := os.ReadFile(changelogPath)
	if err != nil {
		fmt.Printf("❌ Error reading CHANGELOG.md: %v\n", err)
		os.Exit(1)
	}

	var newChangelogEntries strings.Builder
	newChangelogEntries.WriteString(fmt.Sprintf("## [%s] - %s\n\n", nextVersion, buildDate))
	hasEntries := false
	for _, cat := range categories {
		entries := changelogSections[cat]
		if len(entries) > 0 {
			hasEntries = true
			newChangelogEntries.WriteString(fmt.Sprintf("### %s\n", cat))
			for _, entry := range entries {
				newChangelogEntries.WriteString(fmt.Sprintf("- %s\n", entry))
			}
			newChangelogEntries.WriteString("\n")
		}
	}

	if !hasEntries {
		newChangelogEntries.WriteString("- Routine maintenance and version bump.\n\n")
	}

	unreleasedHeader := "## [Unreleased]\n\n"
	changelogContentStr := string(existingChangelog)
	index := strings.Index(changelogContentStr, unreleasedHeader)
	if index == -1 {
		index = 0
	} else {
		index += len(unreleasedHeader)
	}

	updatedChangelog := changelogContentStr[:index] + newChangelogEntries.String() + changelogContentStr[index:]
	err = os.WriteFile(changelogPath, []byte(updatedChangelog), 0644)
	if err != nil {
		fmt.Printf("❌ Error updating CHANGELOG.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("  ✓ CHANGELOG.md successfully updated with new release notes.")

	// 9. Commit changes to current branch
	fmt.Println("\n💾 Committing changes to local branch...")
	_, err = runCmd("git", "add", versionGoPath, changelogPath, "wiki/Home.md", "wiki/_Footer.md")
	if err != nil {
		fmt.Printf("❌ Error staging files: %v\n", err)
		os.Exit(1)
	}

	commitMsg := fmt.Sprintf("[bump version] v%s", nextVersion)
	commitOut, err := runCmd("git", "commit", "-m", commitMsg)
	if err != nil {
		fmt.Printf("❌ Error committing files: %s\n", commitOut)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Git commit created on branch %s: %s\n", currentBranch, commitMsg)

	// 10. Push current branch to origin
	fmt.Printf("\n🚀 Pushing branch '%s' to origin remote...\n", currentBranch)
	pushOut, err := runCmd("git", "push", "origin", currentBranch)
	if err != nil {
		fmt.Printf("❌ Error pushing branch to origin: %s\n", pushOut)
		fmt.Println("⚠️ Please resolve remote connection and manually push the branch.")
		os.Exit(1)
	}
	fmt.Println("  ✓ Branch successfully pushed to remote.")

	// 11. Compile local build to verify correctness and update local binary
	fmt.Println("\n🛠 Compiling local developer build...")
	buildOut, err := runCmd("make", "build-dev")
	if err != nil {
		fmt.Printf("❌ Error compiling Go binary: %s\n", buildOut)
		os.Exit(1)
	}
	fmt.Println("  ✓ Local binary successfully compiled and shell completion configured.")

	// 12. Complete
	fmt.Println("\n🎉 Release preparation is successfully complete!")
	fmt.Println("==========================================================================")
	fmt.Printf("  ✓ Updated version to: v%s\n", nextVersion)
	fmt.Printf("  ✓ Committed and pushed version changes to %s.\n", currentBranch)
	fmt.Printf("  ✓ Compiled local dev build with updated version.\n")
	fmt.Println("==========================================================================")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}

func updateWikiFiles(nextVersion string) error {
	// 1. Update wiki/Home.md
	homePath := "wiki/Home.md"
	homeBytes, err := os.ReadFile(homePath)
	if err != nil {
		return fmt.Errorf("failed to read wiki/Home.md: %w", err)
	}
	
	// Replace version badge url in wiki/Home.md
	// e.g., https://img.shields.io/badge/version-v1.3.2-green
	reHome := regexp.MustCompile(`(https://img\.shields\.io/badge/version-v)\d+\.\d+\.\d+`)
	updatedHome := reHome.ReplaceAllString(string(homeBytes), "${1}"+nextVersion)
	
	err = os.WriteFile(homePath, []byte(updatedHome), 0644)
	if err != nil {
		return fmt.Errorf("failed to write wiki/Home.md: %w", err)
	}
	fmt.Println("  ✓ wiki/Home.md version badge updated.")

	// 2. Update wiki/_Footer.md
	footerPath := "wiki/_Footer.md"
	footerBytes, err := os.ReadFile(footerPath)
	if err != nil {
		return fmt.Errorf("failed to read wiki/_Footer.md: %w", err)
	}
	
	// Replace version badge url, release tag url, and alt text in wiki/_Footer.md
	reFooterTag := regexp.MustCompile(`(releases/tag/v)\d+\.\d+\.\d+`)
	reFooterBadge := regexp.MustCompile(`(badge/version-v)\d+\.\d+\.\d+`)
	reFooterAlt := regexp.MustCompile(`(alt="v)\d+\.\d+\.\d+(")`)
	
	updatedFooter := reFooterTag.ReplaceAllString(string(footerBytes), "${1}"+nextVersion)
	updatedFooter = reFooterBadge.ReplaceAllString(updatedFooter, "${1}"+nextVersion)
	updatedFooter = reFooterAlt.ReplaceAllString(updatedFooter, "${1}"+nextVersion+"${2}")
	
	err = os.WriteFile(footerPath, []byte(updatedFooter), 0644)
	if err != nil {
		return fmt.Errorf("failed to write wiki/_Footer.md: %w", err)
	}
	fmt.Println("  ✓ wiki/_Footer.md version badge/tag updated.")
	return nil
}
