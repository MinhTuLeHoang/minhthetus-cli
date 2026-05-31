package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func main() {
	fmt.Println("🚀 Syncing local master branch with origin...")

	// 1. Verify master branch locally
	currentBranch, err := runCmd("git", "branch", "--show-current")
	if err != nil {
		fmt.Printf("❌ Git Error checking branch: %v\n", err)
		os.Exit(1)
	}
	if currentBranch != "master" {
		fmt.Printf("❌ Error: You are currently on branch '%s'. You must switch to 'master' locally before syncing.\n", currentBranch)
		os.Exit(1)
	}
	fmt.Println("  ✓ Confirmed on 'master' branch.")

	// 2. Fetch and pull master
	fmt.Println("  ⏳ Fetching origin and updating local master...")
	_, err = runCmd("git", "fetch", "origin")
	if err != nil {
		fmt.Printf("⚠️ Warning: Could not fetch origin: %v. Proceeding with caution...\n", err)
	} else {
		pullOut, err := runCmd("git", "pull", "origin", "master")
		if err != nil {
			fmt.Printf("❌ Error updating master from origin: %s\n", pullOut)
			os.Exit(1)
		}
		fmt.Println("  ✓ Local master branch is fully updated with origin/master.")
	}

	// Verify sync
	localHead, _ := runCmd("git", "rev-parse", "HEAD")
	originHead, _ := runCmd("git", "rev-parse", "origin/master")
	if localHead != originHead {
		behindCountStr, err := runCmd("git", "rev-list", "--count", "HEAD..origin/master")
		if err == nil {
			behindCount, _ := strconv.Atoi(behindCountStr)
			if behindCount > 0 {
				fmt.Printf("⚠️ Local master is behind origin/master by %d commits. Sync failed to complete cleanly.\n", behindCount)
				os.Exit(1)
			}
		}
	}
	fmt.Println("✅ Master branch synchronization successfully complete!")
}
