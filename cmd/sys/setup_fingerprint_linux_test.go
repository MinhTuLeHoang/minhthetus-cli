//go:build linux

package sys

import (
	"errors"
	"os"
	"strings"
	"testing"
)

type stubLinuxRunner struct {
	lookPaths map[string]string
	outputs   map[string][]byte
	errs      map[string]error
}

func (r *stubLinuxRunner) LookPath(file string) (string, error) {
	if p, ok := r.lookPaths[file]; ok {
		return p, nil
	}
	return "", os.ErrNotExist
}

func (r *stubLinuxRunner) Output(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	return r.outputs[key], r.errs[key]
}

func (r *stubLinuxRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return r.Output(name, args...)
}

func (r *stubLinuxRunner) RunInteractive(name string, args ...string) error {
	return nil
}

func TestLinuxBackendTransitiveIncludesDisableErrorAndFprintdError(t *testing.T) {
	t.Setenv("USER", "alice")
	fs := &fakeFS{
		home: t.TempDir(),
		files: map[string][]byte{
			"/usr/lib64/security/pam_fprintd.so": []byte("ELF"),
			LinuxSudoPath:                        []byte("auth       include      system-auth\n"),
			"/etc/pam.d/system-auth":             []byte("auth       substack     fingerprint-auth\n"),
			"/etc/pam.d/fingerprint-auth":        []byte("auth       sufficient   pam_fprintd.so\nauth       required     pam_unix.so\n"),
		},
	}

	runner := &stubLinuxRunner{
		lookPaths: map[string]string{
			"fprintd-list": "/usr/bin/fprintd-list",
		},
		outputs: map[string][]byte{
			"fprintd-list alice": []byte("Fingerprints for user alice:\n - #0: right-index-finger\n"),
		},
		errs: map[string]error{},
	}

	b := &linuxBackend{runner: runner, fs: fs}

	// 1. Inspect detects transitive include
	st, err := b.Inspect()
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}
	if !st.Enabled || !strings.Contains(st.ConfigPath, "fingerprint-auth") {
		t.Fatalf("expected Enabled=true via fingerprint-auth, got Enabled=%v ConfigPath=%q", st.Enabled, st.ConfigPath)
	}

	// 2. PlanDisable returns explicit error naming /etc/pam.d/fingerprint-auth rather than AlreadySatisfied
	_, err = b.PlanDisable()
	if err == nil || !strings.Contains(err.Error(), "/etc/pam.d/fingerprint-auth") {
		t.Fatalf("expected explicit error naming /etc/pam.d/fingerprint-auth, got %v", err)
	}

	// 3. Empty fprintd-list output preserves exec error
	runner.outputs["fprintd-list alice"] = nil
	runner.errs["fprintd-list alice"] = errors.New("dbus failure")
	q := b.queryFprintdStatus("alice")
	if q.HasDevice || !strings.Contains(q.Reason, "dbus failure") {
		t.Fatalf("expected preserved error in Reason, got %+v", q)
	}
}
