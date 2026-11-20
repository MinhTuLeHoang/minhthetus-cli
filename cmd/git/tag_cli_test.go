package git

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("MINHTHETUS_TEST_GIT_SHIM") == "1" {
		os.Exit(runTestGitShim())
	}
	os.Exit(m.Run())
}

func runTestGitShim() int {
	realGit := os.Getenv("MINHTHETUS_REAL_GIT")
	if realGit == "" {
		return 1
	}

	cmd := exec.Command(realGit, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 1
	}

	if len(os.Args) >= 2 && os.Args[1] == "fetch" {
		marker := os.Getenv("MINHTHETUS_SHIM_MARKER")
		cloneB := os.Getenv("MINHTHETUS_SHIM_CLONE_B")
		tag := os.Getenv("MINHTHETUS_SHIM_TAG")
		if marker != "" && cloneB != "" && tag != "" {
			if _, statErr := os.Stat(marker); os.IsNotExist(statErr) {
				_ = os.WriteFile(marker, []byte("done"), 0644)
				tagCmd := exec.Command(realGit, "-C", cloneB, "tag", "-a", tag, "-m", "concurrent remote tag")
				if err := tagCmd.Run(); err != nil {
					return 1
				}
				pushCmd := exec.Command(realGit, "-C", cloneB, "push", "origin", "refs/tags/"+tag)
				if err := pushCmd.Run(); err != nil {
					return 1
				}
			}
		}
	}

	return 0
}

func buildTestCLIBinary(t *testing.T) string {
	t.Helper()
	binName := "minhthetus-test-cli"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(t.TempDir(), binName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "../..")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build CLI binary: %v\n%s", err, string(out))
	}
	return binPath
}

func mergeEnv(base []string, overrides map[string]string) []string {
	result := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		k, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		skip := false
		for okKey := range overrides {
			if strings.EqualFold(k, okKey) {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, entry)
		}
	}
	for k, v := range overrides {
		result = append(result, k+"="+v)
	}
	return result
}

func runCLICmd(t *testing.T, binPath, workDir string, envOverrides map[string]string, args ...string) (int, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader("")
	cmd.Env = mergeEnv(os.Environ(), envOverrides)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("CLI command %v in %s timed out (hung in non-TTY)\noutput:\n%s", args, workDir, buf.String())
	}

	if err == nil {
		return 0, buf.String()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), buf.String()
	}
	t.Fatalf("unexpected error running CLI %v: %v\noutput:\n%s", args, err, buf.String())
	return -1, ""
}

func runGitTestHelper(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s failed: %v\n%s", args, dir, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func initCommittedGitRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGitTestHelper(t, repo, "init")
	runGitTestHelper(t, repo, "config", "user.name", "Test User")
	runGitTestHelper(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write README failed: %v", err)
	}
	runGitTestHelper(t, repo, "add", "README.md")
	runGitTestHelper(t, repo, "commit", "-m", "initial commit")
	return repo
}

func copyFile(t *testing.T, src, dst string, mode os.FileMode) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("open %s failed: %v", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		t.Fatalf("create %s failed: %v", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		t.Fatalf("copy %s -> %s failed: %v", src, dst, err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close %s failed: %v", dst, err)
	}
}

func TestTagDevStgCLI_NonTTYExitsOnePromptly(t *testing.T) {
	binPath := buildTestCLIBinary(t)
	homeDir := t.TempDir()
	baseEnv := map[string]string{
		"HOME":        homeDir,
		"USERPROFILE": homeDir,
	}

	t.Run("unconfigured_repo_fails_fast_with_exit_1", func(t *testing.T) {
		repo := initCommittedGitRepo(t)
		code, out := runCLICmd(t, binPath, repo, baseEnv, "git", "tag-dev-stg", "-P", "-m", "test")
		if code != 1 {
			t.Fatalf("expected exit code 1 for unconfigured repo in non-TTY, got %d\noutput:\n%s", code, out)
		}
		if !strings.Contains(out, "TTY") && !strings.Contains(out, "terminal") {
			t.Fatalf("expected non-TTY error guidance in output, got:\n%s", out)
		}
	})

	t.Run("invalid_git_config_rule_fails_fast_with_reconfig_hint", func(t *testing.T) {
		repo := initCommittedGitRepo(t)
		runGitTestHelper(t, repo, "config", "--local", LocalConfigSandboxKey, "-v,qc-v")

		code, out := runCLICmd(t, binPath, repo, baseEnv, "git", "tag-dev-stg", "-P", "-m", "test")
		if code != 1 {
			t.Fatalf("expected exit code 1 for invalid .git/config rule, got %d\noutput:\n%s", code, out)
		}
		if !strings.Contains(out, "--reconfig") {
			t.Fatalf("expected --reconfig guidance in output, got:\n%s", out)
		}
	})

	t.Run("missing_bump_flag_in_non_tty_exits_1", func(t *testing.T) {
		repo := initCommittedGitRepo(t)
		runGitTestHelper(t, repo, "config", "--local", LocalConfigSandboxKey, "qc-v")

		code, out := runCLICmd(t, binPath, repo, baseEnv, "git", "tag-dev-stg", "-m", "test")
		if code != 1 {
			t.Fatalf("expected exit code 1 when bump flag missing in non-TTY, got %d\noutput:\n%s", code, out)
		}
		if !strings.Contains(out, "-P/--patch") {
			t.Fatalf("expected bump flag guidance in output, got:\n%s", out)
		}
	})

	t.Run("missing_message_flag_in_non_tty_exits_1", func(t *testing.T) {
		repo := initCommittedGitRepo(t)
		runGitTestHelper(t, repo, "config", "--local", LocalConfigSandboxKey, "qc-v")

		code, out := runCLICmd(t, binPath, repo, baseEnv, "git", "tag-dev-stg", "-P")
		if code != 1 {
			t.Fatalf("expected exit code 1 when -m missing in non-TTY, got %d\noutput:\n%s", code, out)
		}
		if !strings.Contains(out, "-m/--message") {
			t.Fatalf("expected -m/--message guidance in output, got:\n%s", out)
		}
	})

	t.Run("preflight_conflict_in_non_tty_exits_1_with_zero_local_tags", func(t *testing.T) {
		root := t.TempDir()
		bareDir := filepath.Join(root, "remote.git")
		cloneA := filepath.Join(root, "cloneA")
		cloneB := filepath.Join(root, "cloneB")

		runGitTestHelper(t, root, "init", "--bare", bareDir)
		runGitTestHelper(t, root, "clone", bareDir, cloneA)
		runGitTestHelper(t, cloneA, "config", "user.name", "Test User")
		runGitTestHelper(t, cloneA, "config", "user.email", "test@example.com")
		if err := os.WriteFile(filepath.Join(cloneA, "README.md"), []byte("hello\n"), 0644); err != nil {
			t.Fatalf("write README failed: %v", err)
		}
		runGitTestHelper(t, cloneA, "add", "README.md")
		runGitTestHelper(t, cloneA, "commit", "-m", "init")
		runGitTestHelper(t, cloneA, "push", "-u", "origin", "HEAD")
		runGitTestHelper(t, cloneA, "config", "--local", LocalConfigSandboxKey, "qc-v,stg-v")

		runGitTestHelper(t, root, "clone", bareDir, cloneB)
		runGitTestHelper(t, cloneB, "config", "user.name", "Test User")
		runGitTestHelper(t, cloneB, "config", "user.email", "test@example.com")

		realGit, err := exec.LookPath("git")
		if err != nil {
			t.Fatalf("failed to locate git: %v", err)
		}
		selfExe, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable failed: %v", err)
		}
		shimDir := filepath.Join(root, "shim")
		if err := os.MkdirAll(shimDir, 0755); err != nil {
			t.Fatalf("mkdir shim failed: %v", err)
		}
		shimName := "git"
		if runtime.GOOS == "windows" {
			shimName += ".exe"
		}
		copyFile(t, selfExe, filepath.Join(shimDir, shimName), 0755)

		markerPath := filepath.Join(root, "pushed.marker")
		shimEnv := map[string]string{
			"HOME":                     homeDir,
			"USERPROFILE":              homeDir,
			"PATH":                     shimDir + string(os.PathListSeparator) + os.Getenv("PATH"),
			"MINHTHETUS_TEST_GIT_SHIM": "1",
			"MINHTHETUS_REAL_GIT":      realGit,
			"MINHTHETUS_SHIM_CLONE_B":  cloneB,
			"MINHTHETUS_SHIM_TAG":      "stg-v0.0.1",
			"MINHTHETUS_SHIM_MARKER":   markerPath,
		}

		code, out := runCLICmd(t, binPath, cloneA, shimEnv, "git", "tag-dev-stg", "-P", "-m", "test")
		if code != 1 {
			t.Fatalf("expected exit code 1 on pre-flight conflict in non-TTY, got %d\noutput:\n%s", code, out)
		}
		if !strings.Contains(out, "Pre-flight check blocked!") || !strings.Contains(out, "stg-v0.0.1") {
			t.Fatalf("expected pre-flight conflict details for stg-v0.0.1, got:\n%s", out)
		}
		if !strings.Contains(out, "TTY") && !strings.Contains(out, "terminal") {
			t.Fatalf("expected non-TTY conflict error message, got:\n%s", out)
		}

		// Verify 0 partial local tags were created in cloneA
		localTags := runGitTestHelper(t, cloneA, "tag", "-l")
		if localTags != "" {
			t.Fatalf("expected 0 local tags created after pre-flight block, got: %q", localTags)
		}
	})
}
