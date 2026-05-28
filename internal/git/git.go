package git

import (
	"bytes"
	"os"
	"os/exec"
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
