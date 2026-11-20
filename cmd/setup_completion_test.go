package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestEnv(goos, home string, vars map[string]string) completionEnv {
	return completionEnv{
		goos: goos,
		getenv: func(key string) string {
			if vars == nil {
				return ""
			}
			return vars[key]
		},
		userHomeDir: func() (string, error) {
			return home, nil
		},
		lookPath: func(string) (string, error) {
			return "", errors.New("not found")
		},
		parentProc: func() (ShellType, string) {
			return ShellUnknown, ""
		},
		winDocsDir: func(h string) string {
			return filepath.Join(h, "Documents")
		},
		runCommand: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("unexpected subprocess call")
		},
		statPath: os.Stat,
		isTTY:    func() bool { return false },
		out:      &bytes.Buffer{},
	}
}

func TestShouldAutoSetupCompletion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: []string{"minhthetus-cli"}, want: true},
		{name: "regular subcommand", args: []string{"minhthetus-cli", "version"}, want: true},
		{name: "setup-completion", args: []string{"minhthetus-cli", "setup-completion"}, want: false},
		{name: "uninstall", args: []string{"minhthetus-cli", "uninstall"}, want: false},
		{name: "complete", args: []string{"minhthetus-cli", "__complete", "git"}, want: false},
		{name: "completeNoDesc", args: []string{"minhthetus-cli", "__completeNoDesc", "sys"}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldAutoSetupCompletion(tc.args); got != tc.want {
				t.Fatalf("shouldAutoSetupCompletion(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

func TestParseShellFlag(t *testing.T) {
	valid := []struct {
		input string
		want  ShellType
	}{
		{"", ShellUnknown},
		{"powershell", ShellPowerShell},
		{"PowerShell", ShellPowerShell},
		{" pwsh ", ShellPowerShell},
		{"ps", ShellPowerShell},
		{"powershell.exe", ShellPowerShell},
		{"bash", ShellBash},
		{"git-bash", ShellBash},
		{"zsh", ShellZsh},
		{"fish", ShellFish},
	}
	for _, tc := range valid {
		got, err := parseShellFlag(tc.input)
		if err != nil {
			t.Fatalf("parseShellFlag(%q) unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("parseShellFlag(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}

	for _, bad := range []string{"cmd", "tcsh", "nushell", "invalid"} {
		if _, err := parseShellFlag(bad); err == nil {
			t.Fatalf("parseShellFlag(%q) expected error, got nil", bad)
		}
	}
}

func TestSetupCompletion_CobraFlagAndCompletion(t *testing.T) {
	flag := setupCompletionCmd.Flags().Lookup("shell")
	if flag == nil {
		t.Fatal("expected --shell flag to be registered on setupCompletionCmd")
	}
	if flag.Shorthand != "s" {
		t.Fatalf("expected --shell shorthand 's', got %q", flag.Shorthand)
	}

	validArgs, dir := setupCompletionCmd.ValidArgsFunction(setupCompletionCmd, nil, "")
	if dir != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive: %v", dir)
	}
	joinedArgs := strings.Join(validArgs, ",")
	for _, req := range []string{"--shell", "--silent", "-s"} {
		if !strings.Contains(joinedArgs, req) {
			t.Fatalf("ValidArgsFunction missing %q in %v", req, validArgs)
		}
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"__complete", "setup-completion", "--shell", ""})
	defer func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	}()

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("__complete execution failed: %v", err)
	}

	output := out.String()
	for _, shellName := range []string{"powershell", "bash", "zsh", "fish"} {
		if !strings.Contains(output, shellName) {
			t.Fatalf("expected __complete output to suggest %q, got:\n%s", shellName, output)
		}
	}
}

func TestDetectShell_WindowsAndUnix(t *testing.T) {
	tests := []struct {
		name       string
		goos       string
		vars       map[string]string
		parentProc func() (ShellType, string)
		wantShell  ShellType
		wantFlavor string
	}{
		{
			name:       "windows parent pwsh",
			goos:       "windows",
			vars:       map[string]string{"PSModulePath": `C:\Windows\System32\WindowsPowerShell\v1.0\Modules`},
			parentProc: func() (ShellType, string) { return ShellPowerShell, "pwsh" },
			wantShell:  ShellPowerShell,
			wantFlavor: "pwsh",
		},
		{
			name:       "windows parent powershell 5.1",
			goos:       "windows",
			vars:       map[string]string{},
			parentProc: func() (ShellType, string) { return ShellPowerShell, "winps" },
			wantShell:  ShellPowerShell,
			wantFlavor: "winps",
		},
		{
			name:       "windows parent cmd ignores global PSModulePath",
			goos:       "windows",
			vars:       map[string]string{"PSModulePath": `C:\Program Files\WindowsPowerShell\Modules`},
			parentProc: func() (ShellType, string) { return ShellCmd, "" },
			wantShell:  ShellCmd,
			wantFlavor: "",
		},
		{
			name:      "windows git bash MSYSTEM wins over inherited PSModulePath",
			goos:      "windows",
			vars:      map[string]string{"MSYSTEM": "MINGW64", "PSModulePath": `C:\Program Files\WindowsPowerShell\Modules`},
			wantShell: ShellBash,
		},
		{
			name:      "windows SHELL bash wins over PSModulePath",
			goos:      "windows",
			vars:      map[string]string{"SHELL": "/usr/bin/bash", "PSModulePath": `C:\Program Files\WindowsPowerShell\Modules`},
			wantShell: ShellBash,
		},
		{
			name:       "windows pwsh channel env",
			goos:       "windows",
			vars:       map[string]string{"POWERSHELL_DISTRIBUTION_CHANNEL": "MSI:Windows 10 Pro"},
			wantShell:  ShellPowerShell,
			wantFlavor: "pwsh",
		},
		{
			name:      "windows cmd PROMPT fallback over PSModulePath",
			goos:      "windows",
			vars:      map[string]string{"PROMPT": "$P$G", "PSModulePath": `C:\Program Files\WindowsPowerShell\Modules`},
			wantShell: ShellCmd,
		},
		{
			name:       "windows PSModulePath PS 7",
			goos:       "windows",
			vars:       map[string]string{"PSModulePath": `C:\Users\Admin\Documents\PowerShell\Modules;C:\Program Files\PowerShell\7\Modules`},
			wantShell:  ShellPowerShell,
			wantFlavor: "pwsh",
		},
		{
			name:       "windows PSModulePath PS 5.1",
			goos:       "windows",
			vars:       map[string]string{"PSModulePath": `C:\Users\Admin\Documents\WindowsPowerShell\Modules`},
			wantShell:  ShellPowerShell,
			wantFlavor: "winps",
		},
		{
			name:      "unix zsh",
			goos:      "darwin",
			vars:      map[string]string{"SHELL": "/bin/zsh"},
			wantShell: ShellZsh,
		},
		{
			name:      "unix bash",
			goos:      "linux",
			vars:      map[string]string{"SHELL": "/bin/bash"},
			wantShell: ShellBash,
		},
		{
			name:      "unix fish",
			goos:      "linux",
			vars:      map[string]string{"SHELL": "/usr/bin/fish"},
			wantShell: ShellFish,
		},
		{
			name:      "unix unknown",
			goos:      "linux",
			vars:      map[string]string{},
			wantShell: ShellUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(tc.goos, t.TempDir(), tc.vars)
			if tc.parentProc != nil {
				env.parentProc = tc.parentProc
			}
			gotShell, gotFlavor := detectShell(env)
			if gotShell != tc.wantShell || gotFlavor != tc.wantFlavor {
				t.Fatalf("detectShell() = (%q, %q), want (%q, %q)", gotShell, gotFlavor, tc.wantShell, tc.wantFlavor)
			}
		})
	}
}

func TestSetupCompletion_SilentNoSubprocess(t *testing.T) {
	home := t.TempDir()
	var buf bytes.Buffer
	execCalls := 0

	env := newTestEnv("windows", home, map[string]string{
		"PSModulePath": `C:\Users\Admin\Documents\WindowsPowerShell\Modules`,
	})
	env.out = &buf
	env.runCommand = func(name string, args ...string) ([]byte, error) {
		execCalls++
		return nil, errors.New("subprocess must not be called in silent mode")
	}

	if err := setupCompletionWithEnv(true, "", env); err != nil {
		t.Fatalf("silent setup failed: %v", err)
	}
	if execCalls != 0 {
		t.Fatalf("expected 0 subprocess calls in silent mode, got %d", execCalls)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected zero output in silent mode, got %q", buf.String())
	}

	ps5Profile := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	data, err := os.ReadFile(ps5Profile)
	if err != nil {
		t.Fatalf("expected PS 5.1 profile to be created at %s: %v", ps5Profile, err)
	}
	if !strings.Contains(string(data), "Get-Command minhthetus-cli") {
		t.Fatalf("unexpected PS 5.1 snippet: %s", string(data))
	}

	// PS 7+ profile should NOT be created in silent PS 5.1 mode
	ps7Profile := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if _, err := os.Stat(ps7Profile); !os.IsNotExist(err) {
		t.Fatalf("did not expect PS 7+ profile to be created during silent PS 5.1 startup")
	}

	// Silent cmd.exe should be a zero-output no-op
	envCmd := newTestEnv("windows", t.TempDir(), map[string]string{"PROMPT": "$P$G"})
	var cmdBuf bytes.Buffer
	envCmd.out = &cmdBuf
	envCmd.runCommand = env.runCommand
	if err := setupCompletionWithEnv(true, "", envCmd); err != nil {
		t.Fatalf("silent cmd.exe setup failed: %v", err)
	}
	if cmdBuf.Len() != 0 || execCalls != 0 {
		t.Fatalf("expected silent cmd.exe to produce no output and 0 subprocess calls")
	}
}

func TestSetupCompletion_PowerShellProfiles_OneDriveAndDualEditions(t *testing.T) {
	home := t.TempDir()
	oneDriveDocs := filepath.Join(home, "OneDrive", "Documents")
	var buf bytes.Buffer

	env := newTestEnv("windows", home, map[string]string{})
	env.out = &buf
	env.winDocsDir = func(string) string { return oneDriveDocs }
	env.lookPath = func(bin string) (string, error) {
		if bin == "pwsh" || bin == "powershell" {
			return `C:\Program Files\` + bin + ".exe", nil
		}
		return "", errors.New("not found")
	}

	ps5Path := filepath.Join(oneDriveDocs, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(filepath.Dir(ps5Path), 0755); err != nil {
		t.Fatal(err)
	}
	existingUserCode := "$env:EDITOR = 'code'\r\n"
	if err := os.WriteFile(ps5Path, []byte(existingUserCode), 0644); err != nil {
		t.Fatal(err)
	}

	// First explicit non-silent run should configure both PS 7+ and PS 5.1
	if err := setupCompletionWithEnv(false, "powershell", env); err != nil {
		t.Fatalf("setupCompletionWithEnv(powershell) failed: %v", err)
	}

	ps7Path := filepath.Join(oneDriveDocs, "PowerShell", "Microsoft.PowerShell_profile.ps1")
	ps7Bytes, err := os.ReadFile(ps7Path)
	if err != nil {
		t.Fatalf("expected PS 7+ profile at %s: %v", ps7Path, err)
	}
	if !strings.Contains(string(ps7Bytes), completionBlockStart) || !strings.Contains(string(ps7Bytes), completionBlockEnd) {
		t.Fatalf("PS 7+ profile missing marker block:\n%s", string(ps7Bytes))
	}

	ps5Bytes, err := os.ReadFile(ps5Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(ps5Bytes), existingUserCode) {
		t.Fatalf("existing PS 5.1 content was not preserved:\n%s", string(ps5Bytes))
	}
	if !strings.Contains(string(ps5Bytes), "minhthetus-cli completion powershell | Out-String | Invoke-Expression") {
		t.Fatalf("PS 5.1 profile missing completion command:\n%s", string(ps5Bytes))
	}

	// Second run must be idempotent
	beforePS7 := string(ps7Bytes)
	beforePS5 := string(ps5Bytes)
	buf.Reset()
	if err := setupCompletionWithEnv(false, "powershell", env); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	afterPS7, _ := os.ReadFile(ps7Path)
	afterPS5, _ := os.ReadFile(ps5Path)
	if string(afterPS7) != beforePS7 || string(afterPS5) != beforePS5 {
		t.Fatalf("second run was not idempotent")
	}
	if !strings.Contains(buf.String(), "Completion already exists") {
		t.Fatalf("expected idempotent 'Completion already exists' message, got:\n%s", buf.String())
	}
}

func TestSetupCompletion_GitBash_BashProfileHandling(t *testing.T) {
	t.Run("creates bash_profile when all login profiles absent", func(t *testing.T) {
		home := t.TempDir()
		var buf bytes.Buffer
		env := newTestEnv("windows", home, map[string]string{"MSYSTEM": "MINGW64"})
		env.out = &buf

		if err := setupCompletionWithEnv(false, "", env); err != nil {
			t.Fatalf("Git Bash setup failed: %v", err)
		}

		bashrc, err := os.ReadFile(filepath.Join(home, ".bashrc"))
		if err != nil {
			t.Fatalf("expected .bashrc: %v", err)
		}
		if strings.Contains(string(bashrc), "\r\n") {
			t.Fatalf(".bashrc must use LF line endings, not CRLF")
		}
		if !strings.Contains(string(bashrc), "source <(minhthetus-cli completion bash)") {
			t.Fatalf("unexpected .bashrc content:\n%s", string(bashrc))
		}

		bp, err := os.ReadFile(filepath.Join(home, ".bash_profile"))
		if err != nil {
			t.Fatalf("expected .bash_profile to be created: %v", err)
		}
		if !strings.Contains(string(bp), ".bashrc") {
			t.Fatalf("expected .bash_profile to source .bashrc, got:\n%s", string(bp))
		}

		// Idempotent on second run
		beforeBP := string(bp)
		beforeRC := string(bashrc)
		if err := setupCompletionWithEnv(false, "", env); err != nil {
			t.Fatal(err)
		}
		afterBP, _ := os.ReadFile(filepath.Join(home, ".bash_profile"))
		afterRC, _ := os.ReadFile(filepath.Join(home, ".bashrc"))
		if string(afterBP) != beforeBP || string(afterRC) != beforeRC {
			t.Fatalf("expected second Git Bash run to be idempotent")
		}
	})

	t.Run("updates existing bash_profile without bashrc reference", func(t *testing.T) {
		home := t.TempDir()
		bpPath := filepath.Join(home, ".bash_profile")
		if err := os.WriteFile(bpPath, []byte("export PATH=$PATH:/usr/local/bin\n"), 0644); err != nil {
			t.Fatal(err)
		}

		env := newTestEnv("windows", home, map[string]string{"MSYSTEM": "MINGW64"})
		if err := setupCompletionWithEnv(false, "bash", env); err != nil {
			t.Fatal(err)
		}
		bpBytes, _ := os.ReadFile(bpPath)
		if !strings.Contains(string(bpBytes), "minhthetus-cli completion bash") {
			t.Fatalf("expected completion snippet in existing .bash_profile without .bashrc, got:\n%s", string(bpBytes))
		}
	})

	t.Run("does not create bash_profile if .profile exists", func(t *testing.T) {
		home := t.TempDir()
		if err := os.WriteFile(filepath.Join(home, ".profile"), []byte("# existing profile\n"), 0644); err != nil {
			t.Fatal(err)
		}
		env := newTestEnv("windows", home, map[string]string{"MSYSTEM": "MINGW64"})
		if err := setupCompletionWithEnv(false, "bash", env); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(home, ".bash_profile")); !os.IsNotExist(err) {
			t.Fatalf("must not create .bash_profile when .profile already exists")
		}
	})

	t.Run("updates existing bash_profile when .bashrc is only mentioned in a comment", func(t *testing.T) {
		home := t.TempDir()
		bpPath := filepath.Join(home, ".bash_profile")
		if err := os.WriteFile(bpPath, []byte("# NOTE: does not source .bashrc\nexport PATH=$PATH:/usr/local/bin # check .bashrc\n"), 0644); err != nil {
			t.Fatal(err)
		}

		env := newTestEnv("windows", home, map[string]string{"MSYSTEM": "MINGW64"})
		if err := setupCompletionWithEnv(false, "bash", env); err != nil {
			t.Fatal(err)
		}
		bpBytes, _ := os.ReadFile(bpPath)
		if !strings.Contains(string(bpBytes), "minhthetus-cli completion bash") {
			t.Fatalf("expected completion snippet in .bash_profile when .bashrc only appears in comments, got:\n%s", string(bpBytes))
		}
	})

	t.Run("does not modify bash_profile when .bashrc is actively sourced before inline comment", func(t *testing.T) {
		home := t.TempDir()
		bpPath := filepath.Join(home, ".bash_profile")
		initial := "source ~/.bashrc # load interactive settings\n"
		if err := os.WriteFile(bpPath, []byte(initial), 0644); err != nil {
			t.Fatal(err)
		}

		env := newTestEnv("windows", home, map[string]string{"MSYSTEM": "MINGW64"})
		if err := setupCompletionWithEnv(false, "bash", env); err != nil {
			t.Fatal(err)
		}
		bpBytes, _ := os.ReadFile(bpPath)
		if string(bpBytes) != initial {
			t.Fatalf("expected .bash_profile to remain unchanged, got:\n%s", string(bpBytes))
		}
	})
}

func TestSetupCompletion_ZshAndFish(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(home, "custom-xdg")

	env := newTestEnv("linux", home, map[string]string{"XDG_CONFIG_HOME": xdg})
	if err := setupCompletionWithEnv(false, "zsh", env); err != nil {
		t.Fatalf("zsh setup failed: %v", err)
	}
	zshrc, err := os.ReadFile(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(zshrc), "source <(minhthetus-cli completion zsh)") {
		t.Fatalf("unexpected .zshrc:\n%s", string(zshrc))
	}

	if err := setupCompletionWithEnv(false, "fish", env); err != nil {
		t.Fatalf("fish setup failed: %v", err)
	}
	fishPath := filepath.Join(xdg, "fish", "completions", "minhthetus-cli.fish")
	fishData, err := os.ReadFile(fishPath)
	if err != nil {
		t.Fatalf("expected fish completion file at %s: %v", fishPath, err)
	}
	if !strings.Contains(string(fishData), "minhthetus-cli completion fish | source") {
		t.Fatalf("unexpected fish content:\n%s", string(fishData))
	}
}

func TestSetupCompletion_UnknownShellAndInvalidFlag(t *testing.T) {
	home := t.TempDir()
	var buf bytes.Buffer
	env := newTestEnv("windows", home, map[string]string{"PROMPT": "$P$G"})
	env.out = &buf

	if err := setupCompletionWithEnv(false, "invalid-shell", env); err == nil {
		t.Fatal("expected error for invalid --shell value")
	}

	buf.Reset()
	if err := setupCompletionWithEnv(false, "", env); err != nil {
		t.Fatalf("non-TTY cmd.exe should warn and return nil, got: %v", err)
	}
	if !strings.Contains(buf.String(), "--shell powershell") {
		t.Fatalf("expected --shell hint for cmd.exe non-TTY, got:\n%s", buf.String())
	}

	// Interactive TTY choice selects fish
	buf.Reset()
	env.isTTY = func() bool { return true }
	env.chooseShell = func(string, []string) (string, error) { return "fish", nil }
	if err := setupCompletionWithEnv(false, "", env); err != nil {
		t.Fatalf("interactive TTY choice failed: %v", err)
	}
	fishFile := filepath.Join(home, ".config", "fish", "completions", "minhthetus-cli.fish")
	if _, err := os.Stat(fishFile); err != nil {
		t.Fatalf("expected fish completion file from TTY selection: %v", err)
	}
}
