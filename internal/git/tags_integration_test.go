package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func runGitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s failed: %v\noutput: %s", args, dir, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func configureGitIdentity(t *testing.T, repoDir string) {
	t.Helper()
	runGitCmd(t, repoDir, "config", "user.name", "Test User")
	runGitCmd(t, repoDir, "config", "user.email", "test@example.com")
}

func setupBareAndClones(t *testing.T) (bareDir string, localA Client, localB Client) {
	t.Helper()
	root := t.TempDir()
	bareDir = filepath.Join(root, "remote.git")
	dirA := filepath.Join(root, "localA")
	dirB := filepath.Join(root, "localB")

	runGitCmd(t, root, "init", "--bare", bareDir)
	runGitCmd(t, root, "clone", bareDir, dirA)
	configureGitIdentity(t, dirA)

	// Create initial commit in localA and push to bareDir
	if err := os.WriteFile(filepath.Join(dirA, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write README failed: %v", err)
	}
	runGitCmd(t, dirA, "add", "README.md")
	runGitCmd(t, dirA, "commit", "-m", "initial commit")
	runGitCmd(t, dirA, "push", "-u", "origin", "HEAD")

	runGitCmd(t, root, "clone", bareDir, dirB)
	configureGitIdentity(t, dirB)

	return bareDir, Client{Dir: dirA, Remote: "origin"}, Client{Dir: dirB, Remote: "origin"}
}

func TestFetchAndListRemoteTags_AnnotatedAndLightweight(t *testing.T) {
	_, localA, localB := setupBareAndClones(t)

	// Create lightweight + annotated tags in localB and push to bare remote
	runGitCmd(t, localB.Dir, "tag", "qc-v1.0.0")
	if err := localB.CreateAnnotatedTag("stg-v1.0.0", "HEAD", "Release v1.0.0"); err != nil {
		t.Fatalf("CreateAnnotatedTag failed: %v", err)
	}
	runGitCmd(t, localB.Dir, "push", "origin", "qc-v1.0.0", "stg-v1.0.0")

	// ListRemoteTags on localA should strip refs/tags/ and ^{}
	remoteTags, err := localA.ListRemoteTags()
	if err != nil {
		t.Fatalf("ListRemoteTags failed: %v", err)
	}
	wantRemote := map[string]bool{
		"qc-v1.0.0":  true,
		"stg-v1.0.0": true,
	}
	if !reflect.DeepEqual(remoteTags, wantRemote) {
		t.Fatalf("ListRemoteTags = %v, want %v", remoteTags, wantRemote)
	}

	// FetchTags syncs remote tags into localA
	if err := localA.FetchTags(); err != nil {
		t.Fatalf("FetchTags failed: %v", err)
	}
	localTags, err := localA.ListTags("")
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	if len(localTags) != 2 || !localA.TagExistsLocally("qc-v1.0.0") || !localA.TagExistsLocally("stg-v1.0.0") {
		t.Fatalf("expected fetched tags locally, got %v", localTags)
	}
}

func TestLocalAndRemoteTagConflictDetection(t *testing.T) {
	_, localA, localB := setupBareAndClones(t)

	// localA has qc-v1.0.1 locally only; localB pushed stg-v1.0.1 to remote only
	if err := localA.CreateAnnotatedTag("qc-v1.0.1", "HEAD", "local only"); err != nil {
		t.Fatalf("CreateAnnotatedTag failed: %v", err)
	}
	if err := localB.CreateAnnotatedTag("stg-v1.0.1", "HEAD", "remote only"); err != nil {
		t.Fatalf("CreateAnnotatedTag failed: %v", err)
	}
	if err := localB.PushTag("stg-v1.0.1"); err != nil {
		t.Fatalf("PushTag failed: %v", err)
	}

	remoteTags, err := localA.ListRemoteTags()
	if err != nil {
		t.Fatalf("ListRemoteTags failed: %v", err)
	}

	if !localA.TagExistsLocally("qc-v1.0.1") || remoteTags["qc-v1.0.1"] {
		t.Fatalf("expected qc-v1.0.1 to exist locally only")
	}
	if localA.TagExistsLocally("stg-v1.0.1") || !remoteTags["stg-v1.0.1"] {
		t.Fatalf("expected stg-v1.0.1 to exist remotely only")
	}
}

func TestSequentialTagAndPush(t *testing.T) {
	bareDir, localA, _ := setupBareAndClones(t)

	for _, tag := range []string{"qc-v1.0.2", "stg-v1.0.2"} {
		if err := localA.CreateAnnotatedTag(tag, "HEAD", "Release v1.0.2"); err != nil {
			t.Fatalf("CreateAnnotatedTag(%s) failed: %v", tag, err)
		}
		if err := localA.PushTag(tag); err != nil {
			t.Fatalf("PushTag(%s) failed: %v", tag, err)
		}
	}

	// Verify annotated tag message & existence on bare remote
	msgOut := runGitCmd(t, localA.Dir, "tag", "-l", "--format=%(contents:subject)", "qc-v1.0.2")
	if msgOut != "Release v1.0.2" {
		t.Fatalf("annotated message = %q, want 'Release v1.0.2'", msgOut)
	}
	bareTags := runGitCmd(t, bareDir, "tag", "-l")
	if !strings.Contains(bareTags, "qc-v1.0.2") || !strings.Contains(bareTags, "stg-v1.0.2") {
		t.Fatalf("expected both tags on bare remote, got: %q", bareTags)
	}
}

func TestLocalConfigPersistence(t *testing.T) {
	_, localA, _ := setupBareAndClones(t)

	val, err := localA.GetLocalConfig("minhthetus.tag.sandbox")
	if err != nil {
		t.Fatalf("GetLocalConfig unset unexpected error: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string for unset key, got %q", val)
	}

	if err := localA.SetLocalConfig("minhthetus.tag.sandbox", "qc-v,stg-v"); err != nil {
		t.Fatalf("SetLocalConfig failed: %v", err)
	}
	got, err := localA.GetLocalConfig("minhthetus.tag.sandbox")
	if err != nil || got != "qc-v,stg-v" {
		t.Fatalf("GetLocalConfig got (%q, %v), want ('qc-v,stg-v', nil)", got, err)
	}

	// Empty value unsets cleanly
	if err := localA.SetLocalConfig("minhthetus.tag.sandbox", ""); err != nil {
		t.Fatalf("SetLocalConfig unset failed: %v", err)
	}
	gotAfterUnset, err := localA.GetLocalConfig("minhthetus.tag.sandbox")
	if err != nil || gotAfterUnset != "" {
		t.Fatalf("expected unset empty value, got (%q, %v)", gotAfterUnset, err)
	}

	// Outside a git repo must return an error (not swallow as unset)
	nonRepo := Client{Dir: t.TempDir()}
	if _, err := nonRepo.GetLocalConfig("minhthetus.tag.sandbox"); err == nil {
		t.Fatalf("expected error when running GetLocalConfig outside a git repository")
	}
}

func TestMissingOriginAndEmptyRepo(t *testing.T) {
	emptyDir := t.TempDir()
	runGitCmd(t, emptyDir, "init")
	configureGitIdentity(t, emptyDir)

	c := Client{Dir: emptyDir, Remote: "origin"}
	if !c.IsRepo() {
		t.Fatalf("expected IsRepo() == true")
	}
	if c.HasCommits() {
		t.Fatalf("expected HasCommits() == false on repo with 0 commits")
	}
	tags, err := c.ListTags("")
	if err != nil || len(tags) != 0 {
		t.Fatalf("expected empty tags on empty repo, got (%v, %v)", tags, err)
	}

	// Add 1 commit without adding an origin remote
	if err := os.WriteFile(filepath.Join(emptyDir, "a.txt"), []byte("a\n"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	runGitCmd(t, emptyDir, "add", "a.txt")
	runGitCmd(t, emptyDir, "commit", "-m", "first")

	if !c.HasCommits() {
		t.Fatalf("expected HasCommits() == true")
	}
	if c.HasRemote("origin") {
		t.Fatalf("expected HasRemote('origin') == false")
	}
	if err := c.FetchTags(); err != nil {
		t.Fatalf("expected FetchTags() graceful nil without origin, got %v", err)
	}
	remoteTags, err := c.ListRemoteTags()
	if err != nil || len(remoteTags) != 0 {
		t.Fatalf("expected empty map and nil error without origin, got (%v, %v)", remoteTags, err)
	}
}

func TestCreateAnnotatedTag_RejectsLeadingDashAndEmpty(t *testing.T) {
	c := Client{Dir: t.TempDir()}
	if err := c.CreateAnnotatedTag("-v1.0.0", "HEAD", "msg"); err == nil {
		t.Fatalf("expected error for tag starting with '-'")
	}
	if err := c.CreateAnnotatedTag("   ", "HEAD", "msg"); err == nil {
		t.Fatalf("expected error for empty tag name")
	}
}
