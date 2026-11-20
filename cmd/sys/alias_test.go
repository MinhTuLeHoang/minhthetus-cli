package sys

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
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

	invalidNames := []string{"", "   ", "1abc", "-flag", "foo bar", "foo;rm", "a$b", "minhthetus-cli", "Minhthetus-CLI"}
	for _, n := range invalidNames {
		if err := validateAliasName(n); err == nil {
			t.Errorf("expected invalid name %q to fail validation", n)
		}
	}

	if err := validateAliasCommand("git status -sb"); err != nil {
		t.Errorf("expected valid command, got: %v", err)
	}
	for _, c := range []string{"", "   ", "echo hi\necho bye", "echo hi\recho bye"} {
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
		!strings.Contains(psBlock, "function c { Clear-Host }") {
		t.Fatalf("PowerShell block missing c -> Clear-Host:\n%s", psBlock)
	}
	if !strings.Contains(psBlock, "function pull { git pull origin @args }") {
		t.Fatalf("PowerShell block missing pull function:\n%s", psBlock)
	}
	if !strings.Contains(psBlock, "if (-not $?) { return }") || !strings.Contains(psBlock, "if ($?) { return }") {
		t.Fatalf("PowerShell block did not rewrite && / || for combo alias:\n%s", psBlock)
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
	if !strings.Contains(string(updated), "function c { Clear-Host }") {
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
