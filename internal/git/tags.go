package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Client executes git commands in a target directory and remote.
// Zero value uses the current working directory and "origin" remote.
type Client struct {
	Dir    string
	Remote string
}

// DefaultClient returns a Client operating on the current working directory and "origin" remote.
func DefaultClient() Client {
	return Client{Remote: "origin"}
}

func (c Client) remoteName() string {
	if strings.TrimSpace(c.Remote) == "" {
		return "origin"
	}
	return strings.TrimSpace(c.Remote)
}

// Run executes a git command in c.Dir (or current working directory if empty) and returns trimmed output.
func (c Client) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		return msg, err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// IsRepo reports whether c.Dir is inside a git work tree.
func (c Client) IsRepo() bool {
	out, err := c.Run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// HasCommits reports whether HEAD resolves to a valid commit.
func (c Client) HasCommits() bool {
	_, err := c.Run("rev-parse", "--verify", "HEAD")
	return err == nil
}

// CurrentBranch returns the current branch name, or "HEAD" when in detached HEAD state.
func (c Client) CurrentBranch() (string, error) {
	out, err := c.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to resolve current branch: %w (%s)", err, out)
	}
	return out, nil
}

// HasRemote reports whether the specified remote (defaulting to c.Remote / "origin") is configured.
func (c Client) HasRemote(remote string) bool {
	if strings.TrimSpace(remote) == "" {
		remote = c.remoteName()
	}
	_, err := c.Run("remote", "get-url", remote)
	return err == nil
}

// FetchTags fetches tags from the configured remote. If the remote does not exist, it returns nil.
func (c Client) FetchTags() error {
	remote := c.remoteName()
	if !c.HasRemote(remote) {
		return nil
	}
	out, err := c.Run("fetch", "--tags", remote)
	if err != nil {
		return fmt.Errorf("git fetch --tags %s failed: %w (%s)", remote, err, out)
	}
	return nil
}

// ListTags returns local tags (optionally filtered by pattern), normalizing CRLF line endings.
func (c Client) ListTags(pattern string) ([]string, error) {
	args := []string{"tag", "-l"}
	if pattern != "" {
		args = append(args, pattern)
	}
	out, err := c.Run(args...)
	if err != nil {
		return nil, fmt.Errorf("git tag -l failed: %w (%s)", err, out)
	}
	if out == "" {
		return []string{}, nil
	}
	normalized := strings.ReplaceAll(out, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	tags := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	return tags, nil
}

// TagExistsLocally reports whether a tag ref exists in the local repository.
func (c Client) TagExistsLocally(tag string) bool {
	if strings.TrimSpace(tag) == "" {
		return false
	}
	_, err := c.Run("show-ref", "--verify", "--quiet", "refs/tags/"+tag)
	return err == nil
}

// ListRemoteTags queries `git ls-remote --tags <remote>` and returns a set of remote tag names.
// Annotated tag peeled refs (`^{}`) are stripped and deduplicated.
// If the remote does not exist, it returns an empty map and nil error.
func (c Client) ListRemoteTags() (map[string]bool, error) {
	remote := c.remoteName()
	if !c.HasRemote(remote) {
		return map[string]bool{}, nil
	}

	out, err := c.Run("ls-remote", "--tags", remote)
	if err != nil {
		return nil, fmt.Errorf("git ls-remote --tags %s failed: %w (%s)", remote, err, out)
	}

	tags := make(map[string]bool)
	if out == "" {
		return tags, nil
	}

	normalized := strings.ReplaceAll(out, "\r\n", "\n")
	for _, line := range strings.Split(normalized, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ref := fields[1]
		if !strings.HasPrefix(ref, "refs/tags/") {
			continue
		}
		name := strings.TrimPrefix(ref, "refs/tags/")
		name = strings.TrimSuffix(name, "^{}")
		if name != "" {
			tags[name] = true
		}
	}
	return tags, nil
}

// TagExistsRemotely reports whether a specific tag exists on the remote.
func (c Client) TagExistsRemotely(tag string) (bool, error) {
	if strings.TrimSpace(tag) == "" {
		return false, nil
	}
	remote := c.remoteName()
	if !c.HasRemote(remote) {
		return false, nil
	}
	out, err := c.Run("ls-remote", "--tags", remote, "refs/tags/"+tag, "refs/tags/"+tag+"^{}")
	if err != nil {
		return false, fmt.Errorf("git ls-remote --tags %s failed: %w (%s)", remote, err, out)
	}
	return strings.TrimSpace(out) != "", nil
}

// CreateAnnotatedTag creates an annotated tag pointing to target (defaults to HEAD when empty).
func (c Client) CreateAnnotatedTag(tag, target, message string) error {
	trimmedTag := strings.TrimSpace(tag)
	if trimmedTag == "" {
		return fmt.Errorf("tag name cannot be empty")
	}
	if strings.HasPrefix(tag, "-") || strings.HasPrefix(trimmedTag, "-") {
		return fmt.Errorf("invalid tag name %q: tag name cannot start with '-'", trimmedTag)
	}

	args := []string{"tag", "-a", trimmedTag}
	if strings.TrimSpace(target) != "" {
		args = append(args, target)
	}
	args = append(args, "-m", message)

	out, err := c.Run(args...)
	if err != nil {
		return fmt.Errorf("git tag -a %s failed: %w (%s)", tag, err, out)
	}
	return nil
}

// PushTag pushes a single tag ref to the configured remote.
func (c Client) PushTag(tag string) error {
	remote := c.remoteName()
	out, err := c.Run("push", remote, "refs/tags/"+tag)
	if err != nil {
		return fmt.Errorf("git push %s %s failed: %w (%s)", remote, tag, err, out)
	}
	return nil
}

// Package-level convenience wrappers using DefaultClient().

func IsRepo() bool                              { return DefaultClient().IsRepo() }
func HasCommits() bool                          { return DefaultClient().HasCommits() }
func CurrentBranch() (string, error)            { return DefaultClient().CurrentBranch() }
func HasRemote(remote string) bool              { return DefaultClient().HasRemote(remote) }
func FetchTags() error                          { return DefaultClient().FetchTags() }
func ListTags(pattern string) ([]string, error) { return DefaultClient().ListTags(pattern) }
func TagExistsLocally(tag string) bool          { return DefaultClient().TagExistsLocally(tag) }
func ListRemoteTags() (map[string]bool, error)  { return DefaultClient().ListRemoteTags() }
func TagExistsRemotely(tag string) (bool, error) {
	return DefaultClient().TagExistsRemotely(tag)
}
func CreateAnnotatedTag(tag, target, message string) error {
	return DefaultClient().CreateAnnotatedTag(tag, target, message)
}
func PushTag(tag string) error { return DefaultClient().PushTag(tag) }
