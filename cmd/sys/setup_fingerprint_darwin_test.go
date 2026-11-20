//go:build darwin

package sys

import (
	"os"
	"strings"
	"testing"
)

type stubDarwinRunner struct {
	outputs map[string][]byte
}

func (r *stubDarwinRunner) LookPath(file string) (string, error) {
	return "/usr/bin/" + file, nil
}

func (r *stubDarwinRunner) Output(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	if out, ok := r.outputs[key]; ok {
		return out, nil
	}
	return nil, os.ErrNotExist
}

func (r *stubDarwinRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return r.Output(name, args...)
}

func (r *stubDarwinRunner) RunInteractive(name string, args ...string) error {
	return nil
}

func TestDarwinPlanEnableWarnsOnLegacyPamTidInSudo(t *testing.T) {
	fs := &fakeFS{
		home: t.TempDir(),
		files: map[string][]byte{
			"/usr/lib/pam/pam_tid.so": []byte("macho"),
			DarwinSudoPath: []byte(
				"auth       include        sudo_local\n" +
					"auth       sufficient     pam_tid.so\n" +
					"auth       required       pam_opendirectory.so\n",
			),
			DarwinSudoLocalTemplatePath: []byte("#auth       sufficient     pam_tid.so\n"),
		},
	}
	runner := &stubDarwinRunner{
		outputs: map[string][]byte{
			"sw_vers -productVersion": []byte("14.6\n"),
			"bioutil -c":              []byte("1\n"),
		},
	}

	b := &darwinBackend{runner: runner, fs: fs}
	plan, err := b.PlanEnable()
	if err != nil {
		t.Fatalf("PlanEnable failed: %v", err)
	}
	foundWarn := false
	for _, w := range plan.Warnings {
		if strings.Contains(w, DarwinSudoPath) && strings.Contains(w, DarwinPAMTidModule) {
			foundWarn = true
		}
	}
	if !foundWarn {
		t.Fatalf("expected warning about legacy pam_tid.so in /etc/pam.d/sudo, got %v", plan.Warnings)
	}
}
