package sys

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setAliasTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestValidateAliasNameAndCommand(t *testing.T) {
	validNames := []string{"c", "pull", "master", "_gs", "git_log-1"}
	for _, n := range validNames {
		if err := validateAliasName(n); err != nil {
			t.Errorf("expected valid name %q, got error: %v", n, err)
		}
	}

	invalidNames := []string{
		"", "   ", "1abc", "-flag", "foo bar", "foo;rm", "a$b", "minhthetus-cli", "Minhthetus-CLI",
		"if", "IF", "for", "For", "foreach", "while", "switch", "exit", "return", "function", "in", "case", "esac", "fi", "done", "coproc",
	}
	for _, n := range invalidNames {
		if err := validateAliasName(n); err == nil {
			t.Errorf("expected invalid name %q to fail validation", n)
		}
	}

	validCommands := []string{
		"git status -sb",
		`cd "C:\Users\Admin\projects\" && dir`,
		`echo "a && b" && echo ok`,
		"git status # check status",
	}
	for _, c := range validCommands {
		if err := validateAliasCommand(c); err != nil {
			t.Errorf("expected valid command %q, got: %v", c, err)
		}
	}
	invalidCommands := []string{
		"", "   ", "echo hi\necho bye", "echo hi\recho bye",
		"git status &&", "git status ||", "&& git status", "|| git status",
		"git status && || echo hi", `cd "C:\Users\Admin\" &&`,
	}
	for _, c := range invalidCommands {
		if err := validateAliasCommand(c); err == nil {
			t.Errorf("expected invalid command %q to fail validation", c)
		}
	}
}

func TestAliasConfigPersistenceAndCRUDOperations(t *testing.T) {
	tmpHome := t.TempDir()
	setAliasTestHome(t, tmpHome)

	cfg, err := loadAliasConfig()
	if err != nil {
		t.Fatalf("loadAliasConfig on fresh home failed: %v", err)
	}
	if len(cfg.Aliases) != 0 || len(cfg.AppliedSuggestions) != 0 {
		t.Fatalf("expected empty initial config, got %#v", cfg)
	}

	// Add
	upsertAliasEntry(&cfg, AliasEntry{Name: "gs", Command: "git status"})
	upsertAliasEntry(&cfg, AliasEntry{Name: "master", Command: "preset", Kind: aliasKindPresetMaster})
	if err := saveAliasConfig(cfg); err != nil {
		t.Fatalf("saveAliasConfig failed: %v", err)
	}

	reloaded, err := loadAliasConfig()
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if len(reloaded.Aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(reloaded.Aliases))
	}

	// Filter/search (case-insensitive, sorted)
	matches := filterAliases(reloaded.Aliases, "STATUS")
	if len(matches) != 1 || matches[0].Name != "gs" {
		t.Fatalf("expected match [gs], got %#v", matches)
	}

	// Edit rename-only preserves preset Kind
	masterIdx := findAliasIndex(reloaded.Aliases, "master")
	if err := editAliasEntry(&reloaded, masterIdx, "m", reloaded.Aliases[masterIdx].Command); err != nil {
		t.Fatalf("editAliasEntry rename-only failed: %v", err)
	}
	if reloaded.Aliases[masterIdx].Kind != aliasKindPresetMaster || reloaded.Aliases[masterIdx].Name != "m" {
		t.Fatalf("expected rename-only to keep preset Kind and update name, got %#v", reloaded.Aliases[masterIdx])
	}

	// Edit command clears Kind
	if err := editAliasEntry(&reloaded, masterIdx, "master", "git checkout main && git pull"); err != nil {
		t.Fatalf("editAliasEntry failed: %v", err)
	}
	if reloaded.Aliases[masterIdx].Kind != "" {
		t.Fatalf("expected Kind cleared after command edit, got %q", reloaded.Aliases[masterIdx].Kind)
	}

	// Edit conflict check (case-insensitive)
	if err := editAliasEntry(&reloaded, masterIdx, "GS", "git status"); err == nil {
		t.Fatalf("expected duplicate name conflict error when renaming to existing alias")
	}

	// Delete keeps AppliedSuggestions sticky
	markSuggestionApplied(&reloaded, "c")
	upsertAliasEntry(&reloaded, AliasEntry{Name: "c", Command: "clear"})
	cIdx := findAliasIndex(reloaded.Aliases, "c")
	if _, err := deleteAliasEntry(&reloaded, cIdx); err != nil {
		t.Fatalf("deleteAliasEntry failed: %v", err)
	}
	if findAliasIndex(reloaded.Aliases, "c") != -1 {
		t.Fatalf("expected alias 'c' to be removed")
	}
	rem := remainingSuggestions(reloaded.AppliedSuggestions)
	for _, s := range rem {
		if s.ID == "c" {
			t.Fatalf("expected 'c' to remain sticky in applied_suggestions after deletion")
		}
	}
}

func TestSuggestionsFilteringAndMenuHiding(t *testing.T) {
	// Initially all 3 suggestions remain and "Suggestions" appears in menu.
	opts := mainMenuOptions(nil)
	if !containsString(opts, menuSuggestions) {
		t.Fatalf("expected %q in initial menu options: %#v", menuSuggestions, opts)
	}
	if len(opts) > 7 {
		t.Fatalf("expected <= 7 menu options for inline rendering, got %d", len(opts))
	}

	// Apply subset ("pull")
	applied := []string{"pull"}
	rem := remainingSuggestions(applied)
	if len(rem) != 2 {
		t.Fatalf("expected 2 remaining suggestions, got %d", len(rem))
	}
	for _, s := range rem {
		if s.ID == "pull" {
			t.Fatalf("did not expect 'pull' to reappear in remaining suggestions")
		}
	}
	if !containsString(mainMenuOptions(applied), menuSuggestions) {
		t.Fatalf("expected %q to still appear when 2 suggestions remain", menuSuggestions)
	}

	// Apply all 3 -> "Suggestions" completely hidden from main menu.
	allApplied := []string{"master", "pull", "c"}
	if remAll := remainingSuggestions(allApplied); len(remAll) != 0 {
		t.Fatalf("expected 0 remaining suggestions, got %#v", remAll)
	}
	optsAll := mainMenuOptions(allApplied)
	if containsString(optsAll, menuSuggestions) {
		t.Fatalf("expected %q to be completely hidden when all suggestions applied, got %#v", menuSuggestions, optsAll)
	}
}

func TestRenderBashAndPowerShellBlocks(t *testing.T) {
	if got := RenderBashBlock(nil); got != "" {
		t.Fatalf("expected empty bash block for nil aliases, got %q", got)
	}
	if got := RenderPowerShellBlock(nil); got != "" {
		t.Fatalf("expected empty ps block for nil aliases, got %q", got)
	}

	aliases := []AliasEntry{
		{Name: "master", Command: "preset", Kind: aliasKindPresetMaster},
		{Name: "pull", Command: "git pull origin"},
		{Name: "c", Command: "clear"},
		{Name: "say", Command: "echo 'hi there'"},
		{Name: "combo", Command: "git fetch origin && git status || echo 'fallback && ok'"},
	}

	bashBlock := RenderBashBlock(aliases)
	if !strings.HasPrefix(bashBlock, aliasMarkerStart) || !strings.HasSuffix(bashBlock, aliasMarkerEnd) {
		t.Fatalf("bash block missing markers:\n%s", bashBlock)
	}
	if !strings.Contains(bashBlock, `alias say='echo '\''hi there'\'''`) {
		t.Fatalf("bash block did not properly single-quote escape:\n%s", bashBlock)
	}
	if !strings.Contains(bashBlock, "unalias master 2>/dev/null || true") || !strings.Contains(bashBlock, "git symbolic-ref refs/remotes/origin/HEAD") {
		t.Fatalf("bash block missing master function:\n%s", bashBlock)
	}
	if !strings.Contains(bashBlock, "alias c='clear'") || !strings.Contains(bashBlock, "alias pull='git pull origin'") {
		t.Fatalf("bash block missing pull/c aliases:\n%s", bashBlock)
	}

	psBlock := RenderPowerShellBlock(aliases)
	if !strings.HasPrefix(psBlock, aliasMarkerStart) || !strings.HasSuffix(psBlock, aliasMarkerEnd) {
		t.Fatalf("powershell block missing markers:\n%s", psBlock)
	}
	// Assert PowerShell 5.1 safety: NO unquoted && or || anywhere in rendered .ps1 block.
	for _, line := range strings.Split(psBlock, "\n") {
		if strings.Contains(line, "'fallback && ok'") {
			continue
		}
		if strings.Contains(line, "&&") || strings.Contains(line, "||") {
			t.Fatalf("PowerShell block contains PS 5.1-incompatible && or || on line %q:\n%s", line, psBlock)
		}
	}
	if !strings.Contains(psBlock, "Remove-Item Alias:c -Force -ErrorAction SilentlyContinue") ||
		!strings.Contains(psBlock, "function c {\n    Clear-Host\n}") {
		t.Fatalf("PowerShell block missing multi-line c -> Clear-Host:\n%s", psBlock)
	}
	if !strings.Contains(psBlock, "function pull {\n    git pull origin @args\n}") {
		t.Fatalf("PowerShell block missing multi-line pull function:\n%s", psBlock)
	}
	if !strings.Contains(psBlock, "if ($_ok) {") || !strings.Contains(psBlock, "if (-not $_ok) {") || !strings.Contains(psBlock, "$global:LASTEXITCODE = $null") {
		t.Fatalf("PowerShell block did not rewrite && / || with $_ok state tracking:\n%s", psBlock)
	}
	if !strings.Contains(psBlock, "$LASTEXITCODE") || !strings.Contains(psBlock, "git symbolic-ref refs/remotes/origin/HEAD") {
		t.Fatalf("PowerShell block missing PS 5.1-safe master function:\n%s", psBlock)
	}
}

func TestUpsertMarkerBlock_IdempotencyBOMNewlinesAndErrors(t *testing.T) {
	block1 := RenderBashBlock([]AliasEntry{{Name: "c", Command: "clear"}})
	block2 := RenderBashBlock([]AliasEntry{{Name: "c", Command: "clear"}, {Name: "pull", Command: "git pull origin"}})

	// Insert into existing file preserving surrounding lines and CRLF.
	initial := "# user top\r\nexport FOO=1\r\n"
	out1, changed, err := UpsertMarkerBlock(initial, block1, "\r\n")
	if err != nil || !changed {
		t.Fatalf("initial insert failed (changed=%v, err=%v)", changed, err)
	}
	if !strings.Contains(out1, "\r\n") || strings.Contains(strings.ReplaceAll(out1, "\r\n", ""), "\n") {
		t.Fatalf("expected all CRLF line endings preserved, got %q", out1)
	}
	if !strings.HasPrefix(out1, "# user top\r\nexport FOO=1\r\n") {
		t.Fatalf("user header was modified: %q", out1)
	}

	// Idempotent re-insert of exact same block -> changed == false.
	outSame, changedSame, err := UpsertMarkerBlock(out1, block1, "\r\n")
	if err != nil {
		t.Fatalf("idempotent upsert failed: %v", err)
	}
	if changedSame || outSame != out1 {
		t.Fatalf("expected no change on identical block")
	}

	// Update existing block while preserving trailing user lines.
	withFooter := out1 + "# user footer\r\n"
	out2, changed2, err := UpsertMarkerBlock(withFooter, block2, "\r\n")
	if err != nil || !changed2 {
		t.Fatalf("update failed (changed=%v, err=%v)", changed2, err)
	}
	if !strings.HasPrefix(out2, "# user top\r\nexport FOO=1\r\n") || !strings.HasSuffix(out2, "# user footer\r\n") {
		t.Fatalf("surrounding user content corrupted:\n%q", out2)
	}

	// Remove block cleanly when aliases become empty.
	outRemoved, changedRemoved, err := UpsertMarkerBlock(out2, "", "\r\n")
	if err != nil || !changedRemoved {
		t.Fatalf("removal failed (changed=%v, err=%v)", changedRemoved, err)
	}
	if strings.Contains(outRemoved, aliasMarkerStart) || strings.Contains(outRemoved, aliasMarkerEnd) {
		t.Fatalf("markers still present after removal: %q", outRemoved)
	}
	if !strings.Contains(outRemoved, "# user top") || !strings.Contains(outRemoved, "# user footer") {
		t.Fatalf("user lines lost on removal: %q", outRemoved)
	}

	// Corrupted single marker must return error.
	if _, _, err := UpsertMarkerBlock("# top\n"+aliasMarkerStart+"\nalias c='clear'\n", block1, "\n"); err == nil {
		t.Fatalf("expected error when end marker is missing")
	}
	if _, _, err := UpsertMarkerBlock("# top\n"+aliasMarkerEnd+"\n", block1, "\n"); err == nil {
		t.Fatalf("expected error when start marker is missing")
	}
	if _, _, err := UpsertMarkerBlock(aliasMarkerEnd+"\n"+aliasMarkerStart+"\n", block1, "\n"); err == nil {
		t.Fatalf("expected error when end marker precedes start marker")
	}
}

func TestSyncProfileFile_PreservesUTF8BOMAndSkipsMissingWhenEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	psProfile := filepath.Join(tmpDir, "PowerShell", "Microsoft.PowerShell_profile.ps1")
	target := ProfileTarget{Path: psProfile, Shell: ShellKindPowerShell}

	// Missing profile + empty aliases -> must NOT create file.
	changed, err := syncProfileFile(target, nil)
	if err != nil || changed {
		t.Fatalf("expected no-op for missing file with empty aliases, got changed=%v err=%v", changed, err)
	}
	if _, err := os.Stat(psProfile); !os.IsNotExist(err) {
		t.Fatalf("expected profile file not to be created when aliases empty")
	}

	// Existing profile with UTF-8 BOM + CRLF -> preserve BOM and CRLF.
	if err := os.MkdirAll(filepath.Dir(psProfile), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	original := append([]byte{}, utf8BOM...)
	original = append(original, []byte("Write-Host 'hello'\r\n")...)
	if err := os.WriteFile(psProfile, original, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	changed, err = syncProfileFile(target, []AliasEntry{{Name: "c", Command: "clear"}})
	if err != nil || !changed {
		t.Fatalf("syncProfileFile failed: changed=%v err=%v", changed, err)
	}

	updated, err := os.ReadFile(psProfile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !bytes.HasPrefix(updated, utf8BOM) {
		t.Fatalf("expected UTF-8 BOM to be preserved")
	}
	if !strings.Contains(string(updated), "Write-Host 'hello'\r\n") {
		t.Fatalf("expected existing content and CRLF preserved: %q", string(updated))
	}
	if !strings.Contains(string(updated), "function c {\r\n    Clear-Host\r\n}") {
		t.Fatalf("expected PowerShell alias block appended: %q", string(updated))
	}
}

func TestResolveProfileTargets_UnixAndWindows(t *testing.T) {
	home := filepath.Join("tmp", "testuser")

	// Unix: macOS with existing .bash_profile and zsh shell
	unixTargets := resolveUnixProfileTargets(home, "/bin/zsh", "darwin", func(p string) bool {
		return strings.HasSuffix(p, ".bash_profile")
	})
	if len(unixTargets) != 2 {
		t.Fatalf("expected 2 macOS targets (.zshrc + .bash_profile), got %#v", unixTargets)
	}

	// Unix: Linux fallback when no rc files and empty SHELL
	linuxFallback := resolveUnixProfileTargets(home, "", "linux", func(string) bool { return false })
	if len(linuxFallback) != 1 || filepath.Base(linuxFallback[0].Path) != ".bashrc" {
		t.Fatalf("expected linux fallback ~/.bashrc, got %#v", linuxFallback)
	}

	// Windows: OneDrive Documents + alt Documents + PS 5.1 + PS 7 + Git Bash with deduplication
	oneDriveDocs := filepath.Join(home, "OneDrive", "Documents")
	localDocs := filepath.Join(home, "Documents")
	winTargets := resolveWindowsProfileTargets(
		home,
		oneDriveDocs,
		localDocs,
		`C:\Users\u\Documents\PowerShell\Modules;C:\Windows\system32\WindowsPowerShell\v1.0\Modules`,
		"MINGW64",
		"/usr/bin/bash",
		func(bin string) (string, error) {
			if bin == "powershell.exe" || bin == "pwsh" || bin == "bash.exe" {
				return bin, nil
			}
			return "", errors.New("not found")
		},
		func(p string) bool {
			return strings.Contains(p, ".bash_profile") || strings.Contains(p, filepath.Join("Documents", "PowerShell"))
		},
	)

	var hasPS51OneDrive, hasPS7OneDrive, hasPS7Alt, hasBashrc, hasBashProfile bool
	for _, tgt := range winTargets {
		switch {
		case strings.Contains(tgt.Path, filepath.Join("OneDrive", "Documents", "WindowsPowerShell")):
			hasPS51OneDrive = true
		case strings.Contains(tgt.Path, filepath.Join("OneDrive", "Documents", "PowerShell")):
			hasPS7OneDrive = true
		case strings.Contains(tgt.Path, filepath.Join(localDocs, "PowerShell")):
			hasPS7Alt = true
		case strings.HasSuffix(tgt.Path, ".bashrc"):
			hasBashrc = true
		case strings.HasSuffix(tgt.Path, ".bash_profile"):
			hasBashProfile = true
		}
	}
	if !hasPS51OneDrive || !hasPS7OneDrive || !hasPS7Alt || !hasBashrc || !hasBashProfile {
		t.Fatalf("missing expected Windows profile targets in %#v", winTargets)
	}
}

func TestDefaultReloadCommands(t *testing.T) {
	if got := defaultReloadCommands("darwin", "/bin/zsh", "", "", nil); len(got) != 1 || got[0] != "source ~/.zshrc" {
		t.Fatalf("unexpected macOS zsh reload hint: %#v", got)
	}
	if got := defaultReloadCommands("linux", "/bin/bash", "", "", nil); len(got) != 1 || got[0] != "source ~/.bashrc" {
		t.Fatalf("unexpected linux bash reload hint: %#v", got)
	}
	if got := defaultReloadCommands("windows", "", "", `C:\Windows\system32\WindowsPowerShell\v1.0\Modules`, nil); len(got) != 1 || got[0] != ". $PROFILE" {
		t.Fatalf("unexpected windows powershell reload hint: %#v", got)
	}
	if got := defaultReloadCommands("windows", "/usr/bin/bash", "MINGW64", "", nil); len(got) != 1 || got[0] != "source ~/.bashrc" {
		t.Fatalf("unexpected windows git bash reload hint: %#v", got)
	}
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func TestRenderPowerShellBlock_CommentsQuotedBackslashesNestingAndTrailingOps(t *testing.T) {
	aliases := []AliasEntry{
		{Name: "cmt", Command: "git status # check status"},
		{Name: "scmt", Command: "git status; # check status"},
		{Name: "wbs", Command: `cd "C:\Users\Admin\projects\" && dir`},
		{Name: "sub", Command: `(cd C:\Windows && dir)`},
		{Name: "sb", Command: `& { echo a && echo b }`},
		{Name: "trand", Command: "git status &&"},
		{Name: "aov", Command: "cmd /c exit 1 && echo B || echo C"},
		{Name: "redir", Command: `cmd /c "echo warn 1>&2 & exit 0" 2>$null && echo SHOULD_RUN`},
	}

	psBlock := RenderPowerShellBlock(aliases)

	// 1. Inline # comment must be multi-line and never place @args after #.
	if !strings.Contains(psBlock, "function cmt {\n    git status @args # check status\n}") {
		t.Fatalf("expected multi-line cmt with @args before comment, got:\n%s", psBlock)
	}
	if !strings.Contains(psBlock, "function scmt {\n    git status; # check status\n}") {
		t.Fatalf("expected multi-line scmt without @args after comment, got:\n%s", psBlock)
	}

	// 2. Windows double-quoted path ending in backslash must split && cleanly.
	if !strings.Contains(psBlock, `cd "C:\Users\Admin\projects\"`) {
		t.Fatalf("expected Windows path preserved, got:\n%s", psBlock)
	}

	// 3. No unquoted && or || anywhere in the rendered PowerShell block.
	if hasUnquotedChainOpPS(psBlock) {
		t.Fatalf("PowerShell block still contains unquoted && or ||:\n%s", psBlock)
	}

	// 4. Nested (cd C:\Windows && dir) and & { echo a && echo b } must preserve structure and rewrite inner &&.
	if !strings.Contains(psBlock, "function sub {\n    $(\n") {
		t.Fatalf("expected subshell chain rewritten inside $(...), got:\n%s", psBlock)
	}
	if !strings.Contains(psBlock, "function sb {\n    & {\n") {
		t.Fatalf("expected scriptblock chain rewritten inside & {...}, got:\n%s", psBlock)
	}
}

func TestRenderPowerShellBlock_PS51ASTAndRuntimeSemantics(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell 5.1 probe only runs on Windows")
	}
	psExe, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Skip("powershell.exe not available")
	}

	aliases := []AliasEntry{
		{Name: "master", Command: "preset", Kind: aliasKindPresetMaster},
		{Name: "c", Command: "clear"},
		{Name: "cmt", Command: "echo hi # inline comment"},
		{Name: "scmt", Command: "echo ok; # semicolon comment"},
		{Name: "wbs", Command: `cd "C:\Windows\" && echo WBS_OK`},
		{Name: "sub", Command: `(echo SUB_A && echo SUB_B)`},
		{Name: "sb", Command: `& { echo SB_A && echo SB_B }`},
		{Name: "trand", Command: "echo TRAND_OK &&"},
		{Name: "aov", Command: "cmd /c exit 1 && echo B || echo C"},
		{Name: "redir", Command: `cmd /c "echo warn 1>&2 & exit 0" 2>$null && echo SHOULD_RUN`},
	}

	psBlock := RenderPowerShellBlock(aliases)
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "probe.ps1")
	script := psBlock + "\n" + strings.Join([]string{
		"$tokens = $null",
		"$parseErrors = $null",
		"$raw = Get-Content -LiteralPath $PSCommandPath -Raw",
		"[void][System.Management.Automation.Language.Parser]::ParseInput($raw, [ref]$tokens, [ref]$parseErrors)",
		"if ($parseErrors.Count -gt 0) { $parseErrors | ForEach-Object { Write-Error $_.Message }; exit 2 }",
		"Write-Output ('CMT=' + (cmt world))",
		"Write-Output ('SCMT=' + (scmt))",
		"Write-Output ('WBS=' + ((wbs) -join ','))",
		"Write-Output ('SUB=' + ((sub) -join ','))",
		"Write-Output ('SB=' + ((sb) -join ','))",
		"Write-Output ('TRAND=' + (trand))",
		"Write-Output ('AOV=' + (aov))",
		"Write-Output ('REDIR=' + (redir))",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	out, err := exec.Command(psExe, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath).CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell 5.1 execution failed (%v):\n%s\nScript:\n%s", err, string(out), psBlock)
	}
	text := string(out)
	for _, want := range []string{
		"CMT=hi world",
		"SCMT=ok",
		"WBS=WBS_OK",
		"SUB=SUB_A,SUB_B",
		"SB=SB_A,SB_B",
		"TRAND=TRAND_OK",
		"AOV=C",
		"REDIR=SHOULD_RUN",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in PowerShell 5.1 output, got:\n%s\nBlock:\n%s", want, text, psBlock)
		}
	}
}

func TestLoadAliasConfig_EmptyLegacyAndMalformedRecovery(t *testing.T) {
	tmpHome := t.TempDir()
	setAliasTestHome(t, tmpHome)

	configDir := filepath.Join(tmpHome, ".minhthetus-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	cfgPath := filepath.Join(configDir, aliasConfigFile)

	// Empty, whitespace-only, and legacy "[]" must all load cleanly as empty config.
	for _, raw := range []string{"", "   \n\t ", "[]"} {
		if err := os.WriteFile(cfgPath, []byte(raw), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		cfg, err := loadAliasConfig()
		if err != nil {
			t.Fatalf("loadAliasConfig(%q) failed: %v", raw, err)
		}
		if len(cfg.Aliases) != 0 || len(cfg.AppliedSuggestions) != 0 {
			t.Fatalf("expected empty AliasConfig for %q, got %#v", raw, cfg)
		}
	}

	// Malformed JSON must return *aliasConfigParseError including full path.
	corruptContent := "{bad json"
	if err := os.WriteFile(cfgPath, []byte(corruptContent), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	_, err := loadAliasConfig()
	if err == nil {
		t.Fatalf("expected parse error for malformed JSON")
	}
	var parseErr *aliasConfigParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("expected *aliasConfigParseError, got %T: %v", err, err)
	}
	if parseErr.Path != cfgPath || !strings.Contains(err.Error(), cfgPath) {
		t.Fatalf("expected error to include full path %q, got %q", cfgPath, err.Error())
	}

	// Backup and reset must preserve malformed bytes in .bak and reset config to valid JSON.
	bakPath, err := backupAndResetAliasConfig(parseErr.Path)
	if err != nil {
		t.Fatalf("backupAndResetAliasConfig failed: %v", err)
	}
	if bakPath != cfgPath+".bak" {
		t.Fatalf("unexpected backup path %q", bakPath)
	}
	bakBytes, err := os.ReadFile(bakPath)
	if err != nil || string(bakBytes) != corruptContent {
		t.Fatalf("expected backup to preserve %q, got %q (err=%v)", corruptContent, string(bakBytes), err)
	}
	resetCfg, err := loadAliasConfig()
	if err != nil {
		t.Fatalf("loadAliasConfig after reset failed: %v", err)
	}
	if len(resetCfg.Aliases) != 0 || len(resetCfg.AppliedSuggestions) != 0 {
		t.Fatalf("expected empty reset config, got %#v", resetCfg)
	}
}

func TestFindAliasIndex_ExactMatchPrecedence(t *testing.T) {
	aliases := []AliasEntry{
		{Name: "GS", Command: "git status -s"},
		{Name: "gs", Command: "git status"},
	}
	if idx := findAliasIndex(aliases, "gs"); idx != 1 {
		t.Fatalf("expected exact match index 1 for 'gs', got %d", idx)
	}
	if idx := findAliasIndex(aliases, "GS"); idx != 0 {
		t.Fatalf("expected exact match index 0 for 'GS', got %d", idx)
	}
	if idx := findAliasIndex(aliases, "Gs"); idx != 0 {
		t.Fatalf("expected case-insensitive fallback index 0 for 'Gs', got %d", idx)
	}
}

func TestPartialSyncAndSuggestionOrdering(t *testing.T) {
	tmpHome := t.TempDir()
	setAliasTestHome(t, tmpHome)

	// 1. persistAndSyncWith: partial sync failure (len(targets) > 0 alongside err) must report changed=true.
	cfg := AliasConfig{
		Aliases:            []AliasEntry{{Name: "gs", Command: "git status"}},
		AppliedSuggestions: []string{},
	}
	partialTargets, changed, err := persistAndSyncWith(cfg, func([]AliasEntry) ([]ProfileTarget, error) {
		return []ProfileTarget{{Path: "/tmp/.bashrc", Shell: ShellKindBash}}, errors.New("corrupt .bash_profile")
	})
	if err == nil || !changed || len(partialTargets) != 1 {
		t.Fatalf("expected changed=true and 1 partial target on partial failure, got targets=%#v changed=%v err=%v", partialTargets, changed, err)
	}

	// 2. persistSuggestionsWith: total sync failure (len(targets) == 0 alongside err) must NOT persist applied_suggestions.
	nextCfg := AliasConfig{
		Aliases:            []AliasEntry{{Name: "pull", Command: "git pull origin"}},
		AppliedSuggestions: []string{"pull"},
	}
	targets, changed, err := persistSuggestionsWith(nextCfg, func([]AliasEntry) ([]ProfileTarget, error) {
		return nil, errors.New("all profiles broken")
	})
	if err == nil || changed || len(targets) != 0 {
		t.Fatalf("expected changed=false on total suggestion sync failure, got targets=%#v changed=%v err=%v", targets, changed, err)
	}

	reloaded, loadErr := loadAliasConfig()
	if loadErr != nil {
		t.Fatalf("loadAliasConfig failed: %v", loadErr)
	}
	if len(reloaded.AppliedSuggestions) != 0 {
		t.Fatalf("expected applied_suggestions NOT persisted on total sync failure, got %#v", reloaded.AppliedSuggestions)
	}

	// Verify JSON on disk still has empty applied_suggestions.
	cfgPath := filepath.Join(tmpHome, ".minhthetus-cli", aliasConfigFile)
	raw, _ := os.ReadFile(cfgPath)
	var diskCfg AliasConfig
	if err := json.Unmarshal(raw, &diskCfg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(diskCfg.AppliedSuggestions) != 0 {
		t.Fatalf("expected disk applied_suggestions empty, got %#v", diskCfg.AppliedSuggestions)
	}

	// 3. Partial suggestion sync success (len(targets) > 0) saves applied_suggestions and returns changed=true.
	targets, changed, err = persistSuggestionsWith(nextCfg, func([]AliasEntry) ([]ProfileTarget, error) {
		return []ProfileTarget{{Path: "/tmp/.bashrc", Shell: ShellKindBash}}, errors.New("one profile failed")
	})
	if err == nil || !changed || len(targets) != 1 {
		t.Fatalf("expected changed=true on partial suggestion sync, got targets=%#v changed=%v err=%v", targets, changed, err)
	}
}

func TestSyncProfileFile_PreservesSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "dotfiles")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	realFile := filepath.Join(realDir, ".bashrc")
	if err := os.WriteFile(realFile, []byte("export DOTFILES=1\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if got := resolveProfileWritePath(realFile); got != realFile {
		t.Fatalf("expected resolveProfileWritePath(%q) == %q for regular file, got %q", realFile, realFile, got)
	}
	missingFile := filepath.Join(tmpDir, "missing.bashrc")
	if got := resolveProfileWritePath(missingFile); got != missingFile {
		t.Fatalf("expected resolveProfileWritePath(%q) == %q for missing file, got %q", missingFile, missingFile, got)
	}

	linkPath := filepath.Join(tmpDir, ".bashrc")
	if err := os.Symlink(realFile, linkPath); err != nil {
		// On Windows without Developer Mode / Administrator privilege, os.Symlink returns ERROR_PRIVILEGE_NOT_HELD.
		t.Skipf("os.Symlink unavailable in this environment: %v", err)
	}

	if got := resolveProfileWritePath(linkPath); got != realFile {
		t.Fatalf("expected resolveProfileWritePath(%q) == %q, got %q", linkPath, realFile, got)
	}

	changed, err := syncProfileFile(ProfileTarget{Path: linkPath, Shell: ShellKindBash}, []AliasEntry{{Name: "c", Command: "clear"}})
	if err != nil || !changed {
		t.Fatalf("syncProfileFile on symlink failed: changed=%v err=%v", changed, err)
	}

	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to remain a symlink after syncProfileFile", linkPath)
	}
	content, err := os.ReadFile(realFile)
	if err != nil || !strings.Contains(string(content), "alias c='clear'") {
		t.Fatalf("expected symlink target updated, got %q (err=%v)", string(content), err)
	}
}

func TestMasterPresetAndGitChain_PS51RealGitRepo(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell 5.1 probe only runs on Windows")
	}
	psExe, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Skip("powershell.exe not available")
	}
	gitExe, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()
	originDir := filepath.Join(tmpDir, "origin.git")
	workDir := filepath.Join(tmpDir, "work")
	seedDir := filepath.Join(tmpDir, "seed")

	runGit := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command(gitExe, args...)
		if dir != "" {
			cmd.Dir = dir
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v (dir=%s) failed: %v\n%s", args, dir, err, string(out))
		}
		return strings.TrimSpace(string(out))
	}

	runGit("", "init", "--bare", originDir)
	runGit("", "clone", originDir, seedDir)
	runGit(seedDir, "config", "user.email", "test@example.com")
	runGit(seedDir, "config", "user.name", "Test")
	runGit(seedDir, "checkout", "-b", "main")
	if err := os.WriteFile(filepath.Join(seedDir, "README.md"), []byte("v1\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	runGit(seedDir, "add", ".")
	runGit(seedDir, "commit", "-m", "init")
	runGit(seedDir, "push", "-u", "origin", "main")
	runGit(originDir, "symbolic-ref", "HEAD", "refs/heads/main")

	runGit("", "clone", originDir, workDir)
	runGit(workDir, "config", "user.email", "test@example.com")
	runGit(workDir, "config", "user.name", "Test")
	runGit(workDir, "checkout", "-b", "feat/x")

	// Advance origin/main so master alias pulls v2.
	if err := os.WriteFile(filepath.Join(seedDir, "README.md"), []byte("v2\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	runGit(seedDir, "add", ".")
	runGit(seedDir, "commit", "-m", "v2")
	runGit(seedDir, "push", "origin", "main")

	psBlock := RenderPowerShellBlock([]AliasEntry{
		{Name: "master", Command: "preset", Kind: aliasKindPresetMaster},
		{Name: "gchain", Command: "git checkout main 2>$null && echo CHAIN_RAN"},
	})

	scriptPath := filepath.Join(tmpDir, "git_probe.ps1")
	script := psBlock + "\n" + strings.Join([]string{
		"Set-Location -LiteralPath '" + strings.ReplaceAll(workDir, "'", "''") + "'",
		"master",
		"$branchAfter = (git branch --show-current).Trim()",
		"$readme = (Get-Content -LiteralPath 'README.md' -Raw).Trim()",
		"git checkout feat/x 2>$null",
		"$gchainOut = (gchain)",
		"Write-Output \"BRANCH=$branchAfter README=$readme GCHAIN=$gchainOut\"",
	}, "\n")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	out, err := exec.Command(psExe, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath).CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell 5.1 git probe failed (%v):\n%s", err, string(out))
	}
	text := string(out)
	if !strings.Contains(text, "BRANCH=main") || !strings.Contains(text, "README=v2") || !strings.Contains(text, "CHAIN_RAN") {
		t.Fatalf("unexpected PS 5.1 git probe output:\n%s", text)
	}
}
