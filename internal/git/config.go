package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GetLocalConfig reads a repository-local git config value (`git config --local --get <key>`).
// If the key is unset (git exit code 1), it returns ("", nil).
// Other errors (such as running outside a git repository) are returned.
func (c Client) GetLocalConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--local", "--get", key)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("git config --local --get %s failed: %w (%s)", key, err, msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// SetLocalConfig sets a repository-local git config value (`git config --local <key> <value>`).
// If value is empty, it unsets the key instead (ignoring exit code 5 when already unset).
func (c Client) SetLocalConfig(key, value string) error {
	if strings.TrimSpace(value) == "" {
		cmd := exec.Command("git", "config", "--local", "--unset", key)
		if c.Dir != "" {
			cmd.Dir = c.Dir
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 5 {
				return nil
			}
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				msg = strings.TrimSpace(stdout.String())
			}
			return fmt.Errorf("git config --local --unset %s failed: %w (%s)", key, err, msg)
		}
		return nil
	}

	out, err := c.Run("config", "--local", key, value)
	if err != nil {
		return fmt.Errorf("git config --local %s failed: %w (%s)", key, err, out)
	}
	return nil
}

// GetLocalConfig reads a repository-local git config value in the current working directory.
func GetLocalConfig(key string) (string, error) {
	return DefaultClient().GetLocalConfig(key)
}

// SetLocalConfig sets a repository-local git config value in the current working directory.
func SetLocalConfig(key, value string) error {
	return DefaultClient().SetLocalConfig(key, value)
}
