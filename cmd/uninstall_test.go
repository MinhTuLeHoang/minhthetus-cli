package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripCompletionHooks(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        string
		wantChanged bool
	}{
		{
			name:        "empty file",
			input:       "",
			want:        "",
			wantChanged: false,
		},
		{
			name:        "unrelated content untouched",
			input:       "if ($true) {\r\n    Write-Host 'hello'\r\n}\r\n",
			want:        "if ($true) {\r\n    Write-Host 'hello'\r\n}\r\n",
			wantChanged: false,
		},
		{
			name: "marker-delimited PowerShell block with CRLF and surrounding code",
			input: "Set-Alias k kubectl\r\n\r\n" +
				"# >>> minhthetus-cli completion >>>\r\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"# <<< minhthetus-cli completion <<<\r\n\r\n" +
				"if ($true) {\r\n    Write-Host \"keep\"\r\n}\r\n",
			want:        "Set-Alias k kubectl\r\n\r\nif ($true) {\r\n    Write-Host \"keep\"\r\n}\r\n",
			wantChanged: true,
		},
		{
			name: "legacy unmarked 4-line PowerShell block leaves zero orphan brace",
			input: "function prompt { \"PS> \" }\r\n\r\n" +
				"# minhthetus-cli completion\r\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n\r\n" +
				"if ($env:USER) {\r\n    Write-Host $env:USER\r\n}\r\n",
			want:        "function prompt { \"PS> \" }\r\n\r\nif ($env:USER) {\r\n    Write-Host $env:USER\r\n}\r\n",
			wantChanged: true,
		},
		{
			name: "legacy Allman style and nested braces inside PowerShell block",
			input: "Write-Host 'before'\n" +
				"# minhthetus-cli completion\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue)\n" +
				"{\n" +
				"    if ($true) {\n" +
				"        $x = \"keep } inside string\"\n" +
				"        minhthetus-cli completion powershell | Out-String | Invoke-Expression\n" +
				"    }\n" +
				"}\n" +
				"Write-Host 'after'\n",
			want:        "Write-Host 'before'\nWrite-Host 'after'\n",
			wantChanged: true,
		},
		{
			name: "unclosed malformed PowerShell block does not truncate user code",
			input: "if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\n" +
				"    Write-Host 'important user script'\n",
			want: "if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\n" +
				"    Write-Host 'important user script'\n",
			wantChanged: false,
		},
		{
			name: "UTF-8 BOM preserved when user code remains",
			input: "\xef\xbb\xbf$env:FOO = 'bar'\r\n\r\n" +
				"# >>> minhthetus-cli completion >>>\r\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"# <<< minhthetus-cli completion <<<\r\n",
			want:        "\xef\xbb\xbf$env:FOO = 'bar'\r\n",
			wantChanged: true,
		},
		{
			name: "profile containing only completion block cleans to empty",
			input: "# >>> minhthetus-cli completion >>>\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\n" +
				"}\n" +
				"# <<< minhthetus-cli completion <<<\n",
			want:        "",
			wantChanged: true,
		},
		{
			name: "legacy single-line zsh and bash if-fi blocks",
			input: "export PATH=$HOME/bin:$PATH\n\n" +
				"# minhthetus-cli completion\n" +
				"source <(minhthetus-cli completion zsh)\n\n" +
				"if command -v minhthetus-cli >/dev/null 2>&1; then\n" +
				"    source <(minhthetus-cli completion bash)\n" +
				"fi\n\n" +
				"alias ll='ls -la'\n",
			want:        "export PATH=$HOME/bin:$PATH\n\nalias ll='ls -la'\n",
			wantChanged: true,
		},
		{
			name: "legacy Fish if-end block",
			input: "set -gx EDITOR vim\n" +
				"if type -q minhthetus-cli\n" +
				"    minhthetus-cli completion fish | source\n" +
				"end\n" +
				"alias g git\n",
			want:        "set -gx EDITOR vim\nalias g git\n",
			wantChanged: true,
		},
		{
			name: "bash if with unquoted param expansion braces",
			input: "if [ -n ${BASH_VERSION} ] && command -v minhthetus-cli >/dev/null 2>&1; then\n" +
				"    source <(minhthetus-cli completion bash)\n" +
				"fi\n" +
				"export KEEP=1\n",
			want:        "export KEEP=1\n",
			wantChanged: true,
		},
		{
			name: "bash if with command group braces",
			input: "if { command -v minhthetus-cli; } >/dev/null 2>&1; then\n" +
				"    source <(minhthetus-cli completion bash)\n" +
				"fi\n" +
				"export KEEP=1\n",
			want:        "export KEEP=1\n",
			wantChanged: true,
		},
		{
			name: "bash single-line if followed by user if",
			input: "if command -v minhthetus-cli >/dev/null 2>&1; then source <(minhthetus-cli completion bash); fi\n" +
				"export KEEP=1\n" +
				"if [ -d /tmp ]; then\n" +
				"    echo ok\n" +
				"fi\n",
			want: "export KEEP=1\n" +
				"if [ -d /tmp ]; then\n" +
				"    echo ok\n" +
				"fi\n",
			wantChanged: true,
		},
		{
			name: "bash nested single-line if inside multi-line completion if",
			input: "if command -v minhthetus-cli >/dev/null 2>&1; then\n" +
				"    if [ -n \"$BASH_VERSION\" ]; then source <(minhthetus-cli completion bash); fi\n" +
				"fi\n" +
				"export KEEP=1\n",
			want:        "export KEEP=1\n",
			wantChanged: true,
		},
		{
			name: "fish single-line and nested single-line if-end",
			input: "if type -q minhthetus-cli; minhthetus-cli completion fish | source; end\n" +
				"if type -q minhthetus-cli\n" +
				"    if status is-interactive; minhthetus-cli completion fish | source; end\n" +
				"end\n" +
				"set -gx KEEP 1\n",
			want:        "set -gx KEEP 1\n",
			wantChanged: true,
		},
		{
			name: "powershell multi-line string containing closing brace",
			input: "Write-Host 'before'\r\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    $msg = \"line1\r\n" +
				"}\r\n" +
				"line3\"\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"Write-Host 'after'\r\n",
			want:        "Write-Host 'before'\r\nWrite-Host 'after'\r\n",
			wantChanged: true,
		},
		{
			name: "powershell double-quoted here-string containing closing brace",
			input: "Write-Host 'before'\r\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    $s = @\"\r\n" +
				"}\r\n" +
				"\"@\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"Write-Host 'after'\r\n",
			want:        "Write-Host 'before'\r\nWrite-Host 'after'\r\n",
			wantChanged: true,
		},
		{
			name: "powershell single-quoted here-string containing closing brace",
			input: "Write-Host 'before'\r\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    $s = @'\r\n" +
				"}\r\n" +
				"'@\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"Write-Host 'after'\r\n",
			want:        "Write-Host 'before'\r\nWrite-Host 'after'\r\n",
			wantChanged: true,
		},
		{
			name: "powershell block comment containing closing brace",
			input: "Write-Host 'before'\r\n" +
				"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    <#\r\n" +
				"    }\r\n" +
				"    #>\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"Write-Host 'after'\r\n",
			want:        "Write-Host 'before'\r\nWrite-Host 'after'\r\n",
			wantChanged: true,
		},
		{
			name: "powershell Allman brace with intervening comment line",
			input: "if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue)\r\n" +
				"# setup\r\n" +
				"{\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"Write-Host 'after'\r\n",
			want:        "Write-Host 'after'\r\n",
			wantChanged: true,
		},
		{
			name: "powershell completion if with next-line else block leaves no orphan else",
			input: "if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
				"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
				"}\r\n" +
				"else {\r\n" +
				"    Write-Host 'not installed'\r\n" +
				"}\r\n" +
				"Write-Host 'after'\r\n",
			want:        "Write-Host 'after'\r\n",
			wantChanged: true,
		},
		{
			name: "user powershell function mentioning minhthetus-cli in comment stays intact",
			input: "function Invoke-MyWrapper { # wraps minhthetus-cli\r\n" +
				"    Write-Host 'running'\r\n" +
				"}\r\n",
			want: "function Invoke-MyWrapper { # wraps minhthetus-cli\r\n" +
				"    Write-Host 'running'\r\n" +
				"}\r\n",
			wantChanged: false,
		},
		{
			name: "user powershell hashtable mentioning minhthetus-cli stays intact",
			input: "$cfg = @{ tool = 'minhthetus-cli'\r\n" +
				"    enabled = $true\r\n" +
				"}\r\n",
			want: "$cfg = @{ tool = 'minhthetus-cli'\r\n" +
				"    enabled = $true\r\n" +
				"}\r\n",
			wantChanged: false,
		},
		{
			name: "user powershell if-elseif mentioning minhthetus-cli stays intact",
			input: "if ($env:A -eq '1') {\r\n" +
				"    Write-Host 'A'\r\n" +
				"} elseif ($env:B -eq 'minhthetus-cli') {\r\n" +
				"    Write-Host 'B'\r\n" +
				"}\r\n",
			want: "if ($env:A -eq '1') {\r\n" +
				"    Write-Host 'A'\r\n" +
				"} elseif ($env:B -eq 'minhthetus-cli') {\r\n" +
				"    Write-Host 'B'\r\n" +
				"}\r\n",
			wantChanged: false,
		},
		{
			name: "user powershell if comparing minhthetus-cli string stays intact",
			input: "if ($env:MY_TOOL -eq 'minhthetus-cli') {\r\n" +
				"    Write-Host 'custom user logic'\r\n" +
				"}\r\n",
			want: "if ($env:MY_TOOL -eq 'minhthetus-cli') {\r\n" +
				"    Write-Host 'custom user logic'\r\n" +
				"}\r\n",
			wantChanged: false,
		},
		{
			name: "user bash function mentioning minhthetus-cli in comment stays intact",
			input: "my_fn() { # minhthetus-cli\n" +
				"    echo 'hi'\n" +
				"}\n",
			want: "my_fn() { # minhthetus-cli\n" +
				"    echo 'hi'\n" +
				"}\n",
			wantChanged: false,
		},
		{
			name: "user bash if echoing minhthetus-cli stays intact",
			input: "echo 'keep'\n" +
				"if [ -n \"$X\" ]; then\n" +
				"    echo 'minhthetus-cli'\n" +
				"fi\n",
			want: "echo 'keep'\n" +
				"if [ -n \"$X\" ]; then\n" +
				"    echo 'minhthetus-cli'\n" +
				"fi\n",
			wantChanged: false,
		},
		{
			name:        "standalone exe completion invocation is removed",
			input:       "Write-Host 'keep'\r\nminhthetus-cli.exe completion powershell | Out-String | Invoke-Expression\r\n",
			want:        "Write-Host 'keep'\r\n",
			wantChanged: true,
		},
		{
			name:        "crlf bash user if calling minhthetus-cli completion stays intact",
			input:       "if [ \"$1\" = \"setup\" ]; then\r\n    minhthetus-cli completion bash\r\nfi\r\n",
			want:        "if [ \"$1\" = \"setup\" ]; then\r\n    minhthetus-cli completion bash\r\nfi\r\n",
			wantChanged: false,
		},
		{
			name: "crlf bash escaped single quote before completion block",
			input: "echo \\'hello\r\n" +
				"# minhthetus-cli completion\r\n" +
				"if command -v minhthetus-cli >/dev/null 2>&1; then\r\n" +
				"    source <(minhthetus-cli completion bash)\r\n" +
				"fi\r\n" +
				"export KEEP=1\r\n",
			want:        "echo \\'hello\r\nexport KEEP=1\r\n",
			wantChanged: true,
		},
		{
			name: "powershell trailing backslash before closing quote in user if",
			input: "if ($dir -eq \"C:\\tools\\\" -and $mode -eq \"minhthetus-cli completion\") {\r\n" +
				"    Write-Host 'user code'\r\n" +
				"}\r\n",
			want: "if ($dir -eq \"C:\\tools\\\" -and $mode -eq \"minhthetus-cli completion\") {\r\n" +
				"    Write-Host 'user code'\r\n" +
				"}\r\n",
			wantChanged: false,
		},
		{
			name: "powershell trailing backslash before closing quote in user var",
			input: "$cmd = \"C:\\tools\\\" + \"minhthetus-cli completion\"\r\n" +
				"Write-Host $cmd\r\n",
			want: "$cmd = \"C:\\tools\\\" + \"minhthetus-cli completion\"\r\n" +
				"Write-Host $cmd\r\n",
			wantChanged: false,
		},
		{
			name:        "powershell trailing backslash before closing quote in standalone var without Write-Host",
			input:       "$cmd = \"C:\\tools\\\" + \"minhthetus-cli completion\"\r\n",
			want:        "$cmd = \"C:\\tools\\\" + \"minhthetus-cli completion\"\r\n",
			wantChanged: false,
		},
		{
			name: "bash multiline condition user if calling minhthetus-cli completion stays intact",
			input: "if [ -n \"$A\" ] && \\\n" +
				"   [ -n \"$B\" ]; then\n" +
				"    minhthetus-cli completion bash\n" +
				"fi\n",
			want: "if [ -n \"$A\" ] && \\\n" +
				"   [ -n \"$B\" ]; then\n" +
				"    minhthetus-cli completion bash\n" +
				"fi\n",
			wantChanged: false,
		},
		{
			name: "bash three-line condition with comment and newline then stays intact",
			input: "if [ -n \"$A\" ] && \\\n" +
				"   # check second condition\n" +
				"   [ -n \"$B\" ] || \\\n" +
				"   [ -n \"$C\" ]\n" +
				"then\n" +
				"    minhthetus-cli completion bash\n" +
				"fi\n",
			want: "if [ -n \"$A\" ] && \\\n" +
				"   # check second condition\n" +
				"   [ -n \"$B\" ] || \\\n" +
				"   [ -n \"$C\" ]\n" +
				"then\n" +
				"    minhthetus-cli completion bash\n" +
				"fi\n",
			wantChanged: false,
		},
		{
			name: "bash multiline condition completion if block is removed",
			input: "export KEEP=1\n" +
				"if command -v minhthetus-cli >/dev/null 2>&1 && \\\n" +
				"   [ -n \"$BASH_VERSION\" ]; then\n" +
				"    source <(minhthetus-cli completion bash)\n" +
				"fi\n" +
				"export KEEP2=2\n",
			want:        "export KEEP=1\nexport KEEP2=2\n",
			wantChanged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed := stripCompletionHooks(tc.input)
			if changed != tc.wantChanged {
				t.Fatalf("stripCompletionHooks() changed = %v, want %v", changed, tc.wantChanged)
			}
			if got != tc.want {
				t.Fatalf("stripCompletionHooks() mismatch:\ngot:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

func TestCollectCleanupTargets(t *testing.T) {
	home := t.TempDir()
	oneDrive := filepath.Join(home, "OneDrive")
	winDocs := filepath.Join(home, "CustomDocs")
	xdg := filepath.Join(home, ".xdg")
	customProfile := filepath.Join(home, "CustomPS", "profile.ps1")

	env := newTestEnv("windows", home, map[string]string{
		"OneDrive":        oneDrive,
		"XDG_CONFIG_HOME": xdg,
		"PROFILE":         customProfile,
	})
	env.winDocsDir = func(string) string { return winDocs }

	targets := collectCleanupTargets(env, home)
	hasPath := func(p string, wantDelete bool) bool {
		clean := filepath.Clean(p)
		for _, tgt := range targets {
			if strings.EqualFold(tgt.Path, clean) && tgt.DeleteFile == wantDelete {
				return true
			}
		}
		return false
	}

	required := []struct {
		path   string
		delete bool
	}{
		{filepath.Join(home, ".zshrc"), false},
		{filepath.Join(home, ".bashrc"), false},
		{filepath.Join(home, ".bash_profile"), false},
		{filepath.Join(xdg, "fish", "config.fish"), false},
		{filepath.Join(xdg, "fish", "completions", "minhthetus-cli.fish"), true},
		{filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), false},
		{filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"), false},
		{filepath.Join(winDocs, "PowerShell", "Microsoft.PowerShell_profile.ps1"), false},
		{filepath.Join(oneDrive, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"), false},
		{customProfile, false},
		{filepath.Join(xdg, "powershell", "Microsoft.PowerShell_profile.ps1"), false},
	}

	for _, req := range required {
		if !hasPath(req.path, req.delete) {
			t.Fatalf("expected cleanup target %q (delete=%v) in %+v", req.path, req.delete, targets)
		}
	}
}

func TestCleanupShellConfigWithEnv(t *testing.T) {
	home := t.TempDir()
	var buf bytes.Buffer
	env := newTestEnv("windows", home, map[string]string{})
	env.out = &buf

	ps5 := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	ps7 := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	bashrc := filepath.Join(home, ".bashrc")
	bashProfile := filepath.Join(home, ".bash_profile")
	fishComp := filepath.Join(home, ".config", "fish", "completions", "minhthetus-cli.fish")

	for _, p := range []string{ps5, ps7, fishComp} {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
	}

	_ = os.WriteFile(ps5, []byte("Write-Host 'ok'\r\n"+strings.ReplaceAll(powerShellSnippet(), "\n", "\r\n")+"\r\n"), 0644)
	_ = os.WriteFile(ps7, []byte("# minhthetus-cli completion\r\nif (Get-Command minhthetus-cli) {\r\n    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n}\r\n"), 0644)
	_ = os.WriteFile(bashrc, []byte("export FOO=1\n"+posixSnippet("bash")+"\n"), 0644)
	_ = os.WriteFile(bashProfile, []byte("# generated by minhthetus-cli setup-completion\nif [ -f ~/.bashrc ]; then . ~/.bashrc; fi\n"), 0644)
	_ = os.WriteFile(fishComp, []byte(fishSnippet()+"\n"), 0644)

	cleaned := cleanupShellConfigWithEnv(env)
	if len(cleaned) != 5 {
		t.Fatalf("expected 5 cleaned targets, got %d (%v)", len(cleaned), cleaned)
	}

	gotPS5, _ := os.ReadFile(ps5)
	if string(gotPS5) != "Write-Host 'ok'\r\n" {
		t.Fatalf("unexpected ps5 after cleanup: %q", string(gotPS5))
	}
	gotPS7, _ := os.ReadFile(ps7)
	if string(gotPS7) != "" {
		t.Fatalf("expected empty ps7 after cleanup, got %q", string(gotPS7))
	}
	gotBashProfile, _ := os.ReadFile(bashProfile)
	if string(gotBashProfile) != "if [ -f ~/.bashrc ]; then . ~/.bashrc; fi\n" {
		t.Fatalf("expected .bash_profile to preserve bashrc loader, got %q", string(gotBashProfile))
	}
	if _, err := os.Stat(fishComp); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be deleted", fishComp)
	}
}

func TestCleanupShellConfigWithEnv_CRLFAndBackslashPaths(t *testing.T) {
	home := t.TempDir()
	var buf bytes.Buffer
	env := newTestEnv("windows", home, map[string]string{})
	env.out = &buf

	ps7 := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	bashrc := filepath.Join(home, ".bashrc")
	if err := os.MkdirAll(filepath.Dir(ps7), 0755); err != nil {
		t.Fatal(err)
	}

	psInput := "$cmd = \"C:\\tools\\\" + \"minhthetus-cli completion\"\r\n" +
		"# minhthetus-cli completion\r\n" +
		"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\r\n" +
		"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\r\n" +
		"}\r\n"
	bashInput := "echo \\'hello\r\n" +
		"if [ \"$1\" = \"setup\" ]; then\r\n" +
		"    minhthetus-cli completion bash\r\n" +
		"fi\r\n" +
		"# minhthetus-cli completion\r\n" +
		"if command -v minhthetus-cli >/dev/null 2>&1; then\r\n" +
		"    source <(minhthetus-cli completion bash)\r\n" +
		"fi\r\n"

	_ = os.WriteFile(ps7, []byte(psInput), 0644)
	_ = os.WriteFile(bashrc, []byte(bashInput), 0644)

	cleaned := cleanupShellConfigWithEnv(env)
	if len(cleaned) != 2 {
		t.Fatalf("expected 2 cleaned targets, got %d (%v)", len(cleaned), cleaned)
	}

	gotPS7, _ := os.ReadFile(ps7)
	wantPS7 := "$cmd = \"C:\\tools\\\" + \"minhthetus-cli completion\"\r\n"
	if string(gotPS7) != wantPS7 {
		t.Fatalf("ps7 mismatch:\ngot:\n%q\nwant:\n%q", string(gotPS7), wantPS7)
	}

	gotBashrc, _ := os.ReadFile(bashrc)
	wantBashrc := "echo \\'hello\r\nif [ \"$1\" = \"setup\" ]; then\r\n    minhthetus-cli completion bash\r\nfi\r\n"
	if string(gotBashrc) != wantBashrc {
		t.Fatalf("bashrc mismatch:\ngot:\n%q\nwant:\n%q", string(gotBashrc), wantBashrc)
	}
}

func TestDetectInstallMethod(t *testing.T) {
	tests := []struct {
		path string
		want InstallMethod
	}{
		{"/opt/homebrew/Cellar/minhthetus-cli/1.11.4/bin/minhthetus-cli", Homebrew},
		{"/usr/local/Cellar/minhthetus-cli/1.11.4/bin/minhthetus-cli", Homebrew},
		{"/home/linuxbrew/.linuxbrew/Homebrew/bin/minhthetus-cli", Homebrew},
		{`C:\Users\Admin\AppData\Local\Microsoft\WinGet\Packages\MinhTuLeHoang.minhthetus-cli\minhthetus-cli.exe`, WinGet},
		{"/Users/dev/go/bin/minhthetus-cli", GoInstall},
		{`C:\Users\Admin\go\bin\minhthetus-cli.exe`, GoInstall},
		{"/usr/local/bin/minhthetus-cli", ManualBuild},
	}

	for _, tc := range tests {
		if got := detectInstallMethod(tc.path); got != tc.want {
			t.Fatalf("detectInstallMethod(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}

	// Symlink-aware check (e.g., /usr/local/bin/minhthetus-cli -> .../Cellar/.../minhthetus-cli)
	tmp := t.TempDir()
	cellarTarget := filepath.Join(tmp, "Cellar", "minhthetus-cli", "1.12.0", "bin", "minhthetus-cli")
	if err := os.MkdirAll(filepath.Dir(cellarTarget), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cellarTarget, []byte("bin"), 0755); err != nil {
		t.Fatal(err)
	}
	symPath := filepath.Join(tmp, "minhthetus-cli")
	if err := os.Symlink(cellarTarget, symPath); err == nil {
		if got := detectInstallMethod(symPath); got != Homebrew {
			t.Fatalf("detectInstallMethod(symlink %q) = %q, want %q", symPath, got, Homebrew)
		}
	}
}

func TestRefreshHintForShell(t *testing.T) {
	tests := []struct {
		name       string
		goos       string
		vars       map[string]string
		wantSubstr string
	}{
		{"windows powershell", "windows", map[string]string{"PSModulePath": `C:\WindowsPowerShell\Modules`}, "$PROFILE"},
		{"windows git bash", "windows", map[string]string{"MSYSTEM": "MINGW64"}, "source ~/.bash_profile"},
		{"linux bash", "linux", map[string]string{"SHELL": "/bin/bash"}, "source ~/.bashrc"},
		{"darwin zsh", "darwin", map[string]string{"SHELL": "/bin/zsh"}, "source ~/.zshrc"},
		{"linux fish", "linux", map[string]string{"SHELL": "/usr/bin/fish"}, "exec fish"},
		{"windows cmd", "windows", map[string]string{"PROMPT": "$P$G"}, "Restart your terminal"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(tc.goos, t.TempDir(), tc.vars)
			got := refreshHintForShell(env)
			if !strings.Contains(got, tc.wantSubstr) {
				t.Fatalf("refreshHintForShell() = %q, expected substring %q", got, tc.wantSubstr)
			}
		})
	}
}

func TestRunUninstall_Flows(t *testing.T) {
	t.Run("cancelled by user", func(t *testing.T) {
		var buf bytes.Buffer
		removed := false
		deps := uninstallDeps{
			env:     newTestEnv("windows", t.TempDir(), nil),
			exePath: func() (string, error) { return `C:\tools\minhthetus-cli.exe`, nil },
			confirm: func(string) (bool, error) { return false, nil },
			removeBinary: func(string) (bool, error) {
				removed = true
				return false, nil
			},
			out: &buf,
		}
		if err := runUninstall(deps); err != nil {
			t.Fatal(err)
		}
		if removed {
			t.Fatal("removeBinary must not be called when cancelled")
		}
		if !strings.Contains(buf.String(), "cancelled") {
			t.Fatalf("expected cancellation message, got:\n%s", buf.String())
		}
	})

	t.Run("windows scheduled self-delete succeeds without sudo rm", func(t *testing.T) {
		var buf bytes.Buffer
		env := newTestEnv("windows", t.TempDir(), map[string]string{"PSModulePath": `C:\PS\Modules`})
		env.out = &buf
		deps := uninstallDeps{
			env:     env,
			exePath: func() (string, error) { return `C:\Users\Admin\go\bin\minhthetus-cli.exe`, nil },
			confirm: func(string) (bool, error) { return true, nil },
			removeBinary: func(string) (bool, error) {
				return true, nil
			},
			out: &buf,
		}
		if err := runUninstall(deps); err != nil {
			t.Fatal(err)
		}
		outStr := buf.String()
		if strings.Contains(outStr, "sudo rm") || strings.Contains(outStr, "source ~/.zshrc") {
			t.Fatalf("unexpected Unix instructions in Windows output:\n%s", outStr)
		}
		if !strings.Contains(outStr, "Scheduled automatic binary removal") || !strings.Contains(outStr, ". $PROFILE") {
			t.Fatalf("expected Windows scheduled delete and PowerShell hint, got:\n%s", outStr)
		}
	})

	t.Run("windows permission denied prints PowerShell and CMD instructions", func(t *testing.T) {
		var buf bytes.Buffer
		env := newTestEnv("windows", t.TempDir(), nil)
		env.out = &buf
		deps := uninstallDeps{
			env:     env,
			exePath: func() (string, error) { return `C:\Program Files\minhthetus-cli.exe`, nil },
			confirm: func(string) (bool, error) { return true, nil },
			removeBinary: func(string) (bool, error) {
				return false, errors.New("access denied")
			},
			out: &buf,
		}
		if err := runUninstall(deps); err == nil {
			t.Fatal("expected error when binary deletion fails")
		}
		outStr := buf.String()
		if strings.Contains(outStr, "sudo rm") {
			t.Fatalf("Windows output must never suggest sudo rm:\n%s", outStr)
		}
		if !strings.Contains(outStr, "Remove-Item -Force") || !strings.Contains(outStr, "del /f") {
			t.Fatalf("expected PowerShell/CMD manual delete commands, got:\n%s", outStr)
		}
	})

	t.Run("winget failure prints manual winget command", func(t *testing.T) {
		var buf bytes.Buffer
		env := newTestEnv("windows", t.TempDir(), nil)
		env.out = &buf
		deps := uninstallDeps{
			env: env,
			exePath: func() (string, error) {
				return `C:\Users\Admin\AppData\Local\Microsoft\WinGet\Packages\minhthetus-cli.exe`, nil
			},
			confirm:   func(string) (bool, error) { return true, nil },
			runWinget: func() error { return errors.New("exit status 1") },
			out:       &buf,
		}
		if err := runUninstall(deps); err == nil {
			t.Fatal("expected error when winget fails")
		}
		if !strings.Contains(buf.String(), "winget uninstall MinhTuLeHoang.minhthetus-cli") {
			t.Fatalf("expected manual winget command, got:\n%s", buf.String())
		}
	})

	t.Run("homebrew flow succeeds", func(t *testing.T) {
		var buf bytes.Buffer
		brewCalled := false
		env := newTestEnv("darwin", t.TempDir(), map[string]string{"SHELL": "/bin/zsh"})
		env.out = &buf
		deps := uninstallDeps{
			env:     env,
			exePath: func() (string, error) { return "/opt/homebrew/Cellar/minhthetus-cli/1.11.4/bin/minhthetus-cli", nil },
			confirm: func(string) (bool, error) { return true, nil },
			runBrew: func() error {
				brewCalled = true
				return nil
			},
			out: &buf,
		}
		if err := runUninstall(deps); err != nil {
			t.Fatal(err)
		}
		if !brewCalled {
			t.Fatal("expected runBrew to be called")
		}
		if !strings.Contains(buf.String(), "source ~/.zshrc") {
			t.Fatalf("expected zsh refresh hint, got:\n%s", buf.String())
		}
	})
}
