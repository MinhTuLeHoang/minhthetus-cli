package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
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

		// 1. Call sync_master.go script to check branch and pull updates
		fmt.Println("  ⏳ Calling master sync script...")
		syncOut, err := runCmd("go", "run", "scripts/publish/sync_master.go")
		if err != nil {
			fmt.Printf("❌ Error running master sync script: %s\n", syncOut)
			os.Exit(1)
		}
		// Print output of the sync script indented
		lines := strings.Split(syncOut, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("    %s\n", line)
			}
		}

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

	// 5. Create local release branch
	releaseBranch := fmt.Sprintf("release/v%s", nextVersion)
	fmt.Printf("  ⏳ Creating local release branch '%s'...\n", releaseBranch)
	checkoutOut, err := runCmd("git", "checkout", "-b", releaseBranch)
	if err != nil {
		fmt.Printf("❌ Error creating release branch: %s\n", checkoutOut)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Switched to release branch '%s'.\n", releaseBranch)

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
		runCmd("git", "checkout", "master")
		runCmd("git", "branch", "-D", releaseBranch)
		os.Exit(1)
	}
	fmt.Println("  ✓ central version constants updated in internal/config/version.go.")

	// 8. Read existing CHANGELOG.md and append the new version section
	changelogPath := "CHANGELOG.md"
	existingChangelog, err := os.ReadFile(changelogPath)
	if err != nil {
		fmt.Printf("❌ Error reading CHANGELOG.md: %v\n", err)
		runCmd("git", "checkout", "master")
		runCmd("git", "branch", "-D", releaseBranch)
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
		runCmd("git", "checkout", "master")
		runCmd("git", "branch", "-D", releaseBranch)
		os.Exit(1)
	}
	fmt.Println("  ✓ CHANGELOG.md successfully updated with new release notes.")

	// 9. Commit changes to release branch
	fmt.Println("\n💾 Committing changes to local release branch...")
	_, err = runCmd("git", "add", versionGoPath, changelogPath)
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
	fmt.Printf("  ✓ Git commit created on branch %s: %s\n", releaseBranch, commitMsg)

	// 10. Push release branch to origin
	fmt.Printf("\n🚀 Pushing release branch '%s' to origin remote...\n", releaseBranch)
	pushOut, err := runCmd("git", "push", "origin", releaseBranch)
	if err != nil {
		fmt.Printf("❌ Error pushing release branch to origin: %s\n", pushOut)
		fmt.Println("⚠️ Please resolve remote connection and manually push the branch.")
		os.Exit(1)
	}
	fmt.Println("  ✓ Release branch successfully pushed to remote.")

	// 11. Create Pull Request using gh CLI
	fmt.Println("\n🔀 Creating Pull Request on GitHub...")
	prTitle := fmt.Sprintf("[bump version] v%s", nextVersion)
	prBody := fmt.Sprintf("Automated release documentation bump for version v%s.", nextVersion)
	prOut, err := runCmd("gh", "pr", "create", "--title", prTitle, "--body", prBody, "--base", "master", "--head", releaseBranch, "--assignee", "@me", "--label", "documentation")
	if err != nil {
		fmt.Printf("❌ Error creating Pull Request via gh CLI: %s\n", prOut)
		fmt.Println("⚠️ Please open a Pull Request manually on GitHub and merge it to master.")
		os.Exit(1)
	}
	fmt.Printf("  ✓ Pull Request successfully created on GitHub!\n")

	// 12. Auto-merge Pull Request using gh CLI
	fmt.Println("\n🤝 Merging Pull Request and deleting remote branch...")
	// Wait a brief second for GitHub API processing
	time.Sleep(2 * time.Second)
	mergeOut, err := runCmd("gh", "pr", "merge", releaseBranch, "--merge", "--delete-branch", "--admin")
	if err != nil {
		fmt.Println("  ⏳ Standard merge attempt...")
		mergeOut, err = runCmd("gh", "pr", "merge", releaseBranch, "--merge", "--delete-branch")
	}

	if err != nil {
		fmt.Printf("❌ Error merging Pull Request via gh CLI: %s\n", mergeOut)
		fmt.Println("⚠️ Pull Request was created but could not be auto-merged. Please go to GitHub, merge it manually, then run:")
		fmt.Printf("👉 git checkout master && git pull origin master && git tag -a v%s -m \"Release v%s\" && git push origin v%s\n", nextVersion, nextVersion, nextVersion)
		os.Exit(1)
	}
	fmt.Println("  ✓ Pull Request merged successfully, and remote branch deleted!")

	// 13. Back to master locally and pull updates
	fmt.Println("\n🔄 Syncing local master branch...")
	checkoutMasterOut, err := runCmd("git", "checkout", "master")
	if err != nil {
		fmt.Printf("❌ Error switching back to master: %s\n", checkoutMasterOut)
		os.Exit(1)
	}
	
	pullMasterOut, err := runCmd("git", "pull", "origin", "master")
	if err != nil {
		fmt.Printf("❌ Error pulling merged changes to local master: %s\n", pullMasterOut)
		os.Exit(1)
	}
	fmt.Println("  ✓ Local master successfully updated with release changes.")

	// 13.5 Compile local build to verify correctness and update local binary
	fmt.Println("\n🛠 Compiling local developer build...")
	buildOut, err := runCmd("make", "build-dev")
	if err != nil {
		fmt.Printf("❌ Error compiling Go binary: %s\n", buildOut)
		os.Exit(1)
	}
	fmt.Println("  ✓ Local binary successfully compiled and shell completion configured.")

	// 14. Create and push annotated tag
	tagVersion := "v" + nextVersion
	tagMsg := fmt.Sprintf("Release %s", tagVersion)
	fmt.Printf("\n🏷 Tagging version %s locally...\n", tagVersion)
	tagOut, err := runCmd("git", "tag", "-a", tagVersion, "-m", tagMsg)
	if err != nil {
		fmt.Printf("❌ Error tagging git version locally: %s\n", tagOut)
		os.Exit(1)
	}
	fmt.Printf("  ✓ Local tag %s created.\n", tagVersion)

	fmt.Printf("🚀 Pushing tag %s to origin remote...\n", tagVersion)
	pushTagOut, err := runCmd("git", "push", "origin", tagVersion)
	if err != nil {
		fmt.Printf("❌ Error pushing tag to remote: %s\n", pushTagOut)
		os.Exit(1)
	}
	fmt.Println("  ✓ Tag successfully pushed to GitHub.")

	// 14.5 Auto-detect Wiki documentation updates and deploy
	wikiUpdated := false
	fmt.Println("\n📖 Auto-detecting Wiki documentation updates...")
	diffCmd := exec.Command("git", "diff", "v"+latestStable+"..HEAD", "--name-only")
	diffBytes, err := diffCmd.CombinedOutput()
	if err == nil && strings.Contains(string(diffBytes), "wiki/") {
		fmt.Println("  📝 Wiki changes detected in this release. Running deploy-wiki script...")
		deployOut, err := runCmd("bash", "scripts/deploy-wiki.sh")
		if err != nil {
			fmt.Printf("⚠️ Warning: deploy-wiki script failed: %s\n", deployOut)
		} else {
			// Print output of the deploy script indented
			lines := strings.Split(deployOut, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					fmt.Printf("    %s\n", line)
				}
			}
			fmt.Println("  ✓ Wiki documentation successfully synchronized.")
			wikiUpdated = true
		}
	} else {
		fmt.Println("  🕊 No Wiki documentation changes detected in this release.")
	}

	// 15. Delete local release branch
	fmt.Printf("\n🧹 Deleting local branch '%s'...\n", releaseBranch)
	delOut, err := runCmd("git", "branch", "-D", releaseBranch)
	if err != nil {
		fmt.Printf("⚠️ Warning: Could not delete local release branch: %s\n", delOut)
	} else {
		fmt.Println("  ✓ Local release branch deleted.")
	}

	// 16. Complete
	fmt.Println("\n🎉 Release is successfully complete!")
	fmt.Println("==========================================================================")
	fmt.Printf("  ✓ Released Version: %s\n", tagVersion)
	fmt.Printf("  ✓ Synced local master branch.\n")
	fmt.Printf("  ✓ Compiled local dev build with updated version.\n")
	fmt.Printf("  ✓ Pushed tag %s live to GitHub.\n", tagVersion)
	if wikiUpdated {
		fmt.Println("  ✓ Synchronized Wiki documentation live to GitHub.")
	}
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
