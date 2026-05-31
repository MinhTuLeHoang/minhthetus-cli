package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func main() {
	// Query git tags
	tagsOut, err := runCmd("git", "tag", "-l", "v*")
	if err != nil {
		fmt.Printf("❌ Error listing git tags: %v\n", err)
		os.Exit(1)
	}

	tags := strings.Split(tagsOut, "\n")
	var stableTags []string
	stableRegex := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if stableRegex.MatchString(tag) {
			stableTags = append(stableTags, tag)
		}
	}

	latestStable := "1.0.2" // default fallback matching internal/config/version.go
	if len(stableTags) > 0 {
		latestStable = strings.TrimPrefix(stableTags[len(stableTags)-1], "v")
	}
	
	// Output purely the version string for easy machine parsing
	fmt.Println(latestStable)
}
