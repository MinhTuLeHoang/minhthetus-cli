package git

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Run executes a git command and returns its output.
func Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return strings.TrimSpace(stderr.String()), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RunInteractive executes a git command and connects it to the terminal.
func RunInteractive(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// ListBranches returns a list of local branches matching the pattern.
func ListBranches(pattern string) ([]string, error) {
	output, err := Run("branch", "--list", "*"+pattern+"*", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	return strings.Split(output, "\n"), nil
}

// GetRemoteDefaultBranch retrieves the remote default branch name, if configured.
func GetRemoteDefaultBranch() string {
	ref, err := Run("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return ""
}

// IsProtectedBranch checks if a branch is protected by checking the remote default branch and git config patterns.
func IsProtectedBranch(branch string) bool {
	// 1. Check if it matches remote default branch
	defaultBranch := GetRemoteDefaultBranch()
	if defaultBranch != "" && branch == defaultBranch {
		return true
	}

	// 2. Check git config patterns
	cfg, err := Run("config", "--get", "minhthetus-cli.protected-branches")
	if err == nil && cfg != "" {
		var patterns []string
		for _, part := range strings.Split(cfg, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				patterns = append(patterns, trimmed)
			}
		}
		for _, pattern := range patterns {
			if branch == pattern {
				return true
			}
			matched, err := filepath.Match(pattern, branch)
			if err == nil && matched {
				return true
			}
		}
	}

	return false
}

// BranchExistsLocally checks if a branch exists locally.
func BranchExistsLocally(branch string) bool {
	err := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run()
	return err == nil
}

// BranchExistsRemotely checks if a branch exists on remote origin.
func BranchExistsRemotely(branch string) bool {
	out, err := Run("ls-remote", "--heads", "origin", branch)
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(out)) > 0
}

