package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// ShellType represents a supported (or detected) shell environment.
type ShellType string

const (
	ShellPowerShell ShellType = "powershell"
	ShellBash       ShellType = "bash"
	ShellZsh        ShellType = "zsh"
	ShellFish       ShellType = "fish"
	ShellCmd        ShellType = "cmd"
	ShellUnknown    ShellType = ""
)

const (
	completionBlockStart = "# >>> minhthetus-cli completion >>>"
	completionBlockEnd   = "# <<< minhthetus-cli completion <<<"
)

type completionTarget struct {
	Label       string
	ProfilePath string
	Snippet     string
	RefreshHint string
	EOL         string
	BashBridge  bool // true when creating ~/.bash_profile to source ~/.bashrc on Windows Git Bash
}

type completionEnv struct {
	goos        string
	getenv      func(string) string
	userHomeDir func() (string, error)
	lookPath    func(string) (string, error)
	parentProc  func() (ShellType, string)
	winDocsDir  func(home string) string
	runCommand  func(name string, args ...string) ([]byte, error) // MUST NOT be called when silent==true
	statPath    func(string) (os.FileInfo, error)
	isTTY       func() bool
	chooseShell func(title string, options []string) (string, error)
	out         io.Writer
}

func defaultCompletionEnv() completionEnv {
	return completionEnv{
		goos:        runtime.GOOS,
		getenv:      os.Getenv,
		userHomeDir: os.UserHomeDir,
		lookPath:    exec.LookPath,
		parentProc:  detectParentShellProcess,
		winDocsDir:  windowsDocumentsDir,
		runCommand: func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).Output()
		},
		statPath: os.Stat,
		isTTY:    isInteractiveTerminal,
		chooseShell: func(title string, options []string) (string, error) {
			return ui.Choose(title, options)
		},
		out: os.Stdout,
	}
}

func isInteractiveTerminal() bool {
	inInfo, err := os.Stdin.Stat()
	if err != nil || (inInfo.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	outInfo, err := os.Stdout.Stat()
	if err != nil || (outInfo.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	return true
}

// setupCompletionCmd represents the setup-completion command
var setupCompletionCmd = &cobra.Command{
	Use:   "setup-completion",
	Short: "Install tab completion for your shell",
	Long: `Install shell tab completion for minhthetus-cli.

Auto-detects your current shell (PowerShell 5.1, PowerShell 7+, Git Bash,
Bash, Zsh, or Fish) or configures a specific shell via --shell.`,
	Example: `minhthetus-cli setup-completion
minhthetus-cli setup-completion --shell powershell
minhthetus-cli setup-completion --shell bash
minhthetus-cli setup-completion --silent`,
	Args: cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Setup Completion",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"-s", "--shell", "--silent", "-h", "--help"}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		silent, _ := cmd.Flags().GetBool("silent")
		shellFlag, _ := cmd.Flags().GetString("shell")
		if err := setupCompletionWithEnv(silent, shellFlag, defaultCompletionEnv()); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	setupCompletionCmd.Flags().Bool("silent", false, "Run setup without printing progress")
	setupCompletionCmd.Flags().StringP("shell", "s", "", "Target shell to configure (powershell, bash, zsh, fish)")
	_ = setupCompletionCmd.RegisterFlagCompletionFunc("shell", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"powershell", "bash", "zsh", "fish"}, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.AddCommand(setupCompletionCmd)
}

// SetupCompletion ensures that the shell completion is installed in the user's RC/profile file.
func SetupCompletion(silent bool) {
	_ = setupCompletionWithEnv(silent, "", defaultCompletionEnv())
}

// parseShellFlag validates and normalizes the --shell flag value.
func parseShellFlag(raw string) (ShellType, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ShellUnknown, nil
	}
	switch normalized {
	case "powershell", "pwsh", "ps", "powershell.exe", "pwsh.exe":
		return ShellPowerShell, nil
	case "bash", "git-bash", "gitbash", "bash.exe":
		return ShellBash, nil
	case "zsh", "zsh.exe":
		return ShellZsh, nil
	case "fish", "fish.exe":
		return ShellFish, nil
	default:
		return ShellUnknown, fmt.Errorf("unsupported shell %q (supported: powershell, bash, zsh, fish)", raw)
	}
}

// detectShell detects the active shell using in-process ancestor inspection and env vars.
// It returns the ShellType and, for PowerShell, an optional edition flavor ("pwsh" or "winps").
func detectShell(env completionEnv) (ShellType, string) {
	if env.goos == "windows" {
		if env.parentProc != nil {
			if procShell, flavor := env.parentProc(); procShell != ShellUnknown {
				return procShell, flavor
			}
		}

		shellEnv := strings.ToLower(env.getenv("SHELL"))
		switch {
		case strings.Contains(shellEnv, "zsh"):
			return ShellZsh, ""
		case strings.Contains(shellEnv, "fish"):
			return ShellFish, ""
		case strings.Contains(shellEnv, "bash"):
			return ShellBash, ""
		}

		if env.getenv("MSYSTEM") != "" || env.getenv("MINGW_PREFIX") != "" {
			return ShellBash, ""
		}
		if env.getenv("POWERSHELL_DISTRIBUTION_CHANNEL") != "" {
			return ShellPowerShell, "pwsh"
		}
		if env.getenv("PROMPT") != "" {
			return ShellCmd, ""
		}
		if psMod := strings.ToLower(env.getenv("PSModulePath")); psMod != "" {
			if strings.Contains(psMod, `powershell\7`) || strings.Contains(psMod, `powershell/7`) {
				return ShellPowerShell, "pwsh"
			}
			return ShellPowerShell, "winps"
		}
		return ShellUnknown, ""
	}

	shellEnv := strings.ToLower(env.getenv("SHELL"))
	switch {
	case strings.Contains(shellEnv, "zsh"):
		return ShellZsh, ""
	case strings.Contains(shellEnv, "bash"):
		return ShellBash, ""
	case strings.Contains(shellEnv, "fish"):
		return ShellFish, ""
	case strings.Contains(shellEnv, "pwsh"):
		return ShellPowerShell, "pwsh"
	}
	if env.getenv("POWERSHELL_DISTRIBUTION_CHANNEL") != "" {
		return ShellPowerShell, "pwsh"
	}
	return ShellUnknown, ""
}

func xdgConfigOrDefault(home string, getenv func(string) string) string {
	if getenv != nil {
		if xdg := strings.TrimSpace(getenv("XDG_CONFIG_HOME")); xdg != "" {
			return xdg
		}
	}
	return filepath.Join(home, ".config")
}

func resolveWindowsDocsDir(home string, env completionEnv) string {
	if env.winDocsDir != nil {
		if docs := strings.TrimSpace(env.winDocsDir(home)); docs != "" {
			return filepath.Clean(docs)
		}
	}
	defaultDocs := filepath.Join(home, "Documents")
	if env.statPath != nil {
		if _, err := env.statPath(defaultDocs); err == nil {
			return defaultDocs
		}
		if env.getenv != nil {
			for _, varName := range []string{"OneDrive", "OneDriveConsumer", "OneDriveCommercial"} {
				if base := strings.TrimSpace(env.getenv(varName)); base != "" {
					candidate := filepath.Join(base, "Documents")
					if _, err := env.statPath(candidate); err == nil {
						return filepath.Clean(candidate)
					}
				}
			}
		}
	}
	return defaultDocs
}

func powerShellSnippet() string {
	return completionBlockStart + "\n" +
		"if (Get-Command minhthetus-cli -ErrorAction SilentlyContinue) {\n" +
		"    minhthetus-cli completion powershell | Out-String | Invoke-Expression\n" +
		"}\n" +
		completionBlockEnd
}

func posixSnippet(shellName string) string {
	return completionBlockStart + "\n" +
		"if command -v minhthetus-cli >/dev/null 2>&1; then\n" +
		fmt.Sprintf("    source <(minhthetus-cli completion %s)\n", shellName) +
		"fi\n" +
		completionBlockEnd
}

func fishSnippet() string {
	return completionBlockStart + "\n" +
		"if type -q minhthetus-cli\n" +
		"    minhthetus-cli completion fish | source\n" +
		"end\n" +
		completionBlockEnd
}

func resolvePowerShellTargets(psFlavor string, silent bool, env completionEnv, home string) []completionTarget {
	snippet := powerShellSnippet()
	if env.goos != "windows" {
		return []completionTarget{
			{
				Label:       "PowerShell",
				ProfilePath: filepath.Join(xdgConfigOrDefault(home, env.getenv), "powershell", "Microsoft.PowerShell_profile.ps1"),
				Snippet:     snippet,
				RefreshHint: ". $PROFILE",
				EOL:         "\n",
			},
		}
	}

	docsDir := resolveWindowsDocsDir(home, env)
	ps7Path := filepath.Join(docsDir, "PowerShell", "Microsoft.PowerShell_profile.ps1")
	ps5Path := filepath.Join(docsDir, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	if env.getenv != nil {
		if profileEnv := strings.TrimSpace(env.getenv("PROFILE")); profileEnv != "" &&
			strings.HasSuffix(strings.ToLower(profileEnv), ".ps1") {
			lowerProf := strings.ToLower(profileEnv)
			if strings.Contains(lowerProf, "windowspowershell") {
				ps5Path = profileEnv
			} else if strings.Contains(lowerProf, `powershell\`) || strings.Contains(lowerProf, "powershell/") {
				ps7Path = profileEnv
			}
		}
	}

	ps7Target := completionTarget{
		Label:       "PowerShell 7+",
		ProfilePath: ps7Path,
		Snippet:     snippet,
		RefreshHint: ". $PROFILE",
		EOL:         "\r\n",
	}
	ps5Target := completionTarget{
		Label:       "Windows PowerShell 5.1",
		ProfilePath: ps5Path,
		Snippet:     snippet,
		RefreshHint: ". $PROFILE",
		EOL:         "\r\n",
	}

	if silent {
		if psFlavor == "pwsh" {
			return []completionTarget{ps7Target}
		}
		return []completionTarget{ps5Target}
	}

	// Non-silent explicit/interactive mode: configure available PowerShell editions
	hasPS7 := psFlavor == "pwsh"
	if !hasPS7 && env.lookPath != nil {
		if _, err := env.lookPath("pwsh"); err == nil {
			hasPS7 = true
		}
	}
	if !hasPS7 && env.statPath != nil {
		if _, err := env.statPath(filepath.Dir(ps7Path)); err == nil {
			hasPS7 = true
		}
	}

	hasPS5 := psFlavor == "winps"
	if !hasPS5 && env.lookPath != nil {
		if _, err := env.lookPath("powershell"); err == nil {
			hasPS5 = true
		} else if _, err := env.lookPath("powershell.exe"); err == nil {
			hasPS5 = true
		}
	}
	if !hasPS5 && env.statPath != nil {
		if _, err := env.statPath(filepath.Dir(ps5Path)); err == nil {
			hasPS5 = true
		}
	}

	var targets []completionTarget
	if hasPS7 {
		targets = append(targets, ps7Target)
	}
	if hasPS5 {
		targets = append(targets, ps5Target)
	}
	if len(targets) == 0 {
		targets = append(targets, ps5Target)
	}
	return targets
}

func resolveCompletionTargets(shell ShellType, psFlavor string, silent bool, env completionEnv, home string) ([]completionTarget, error) {
	switch shell {
	case ShellPowerShell:
		return resolvePowerShellTargets(psFlavor, silent, env, home), nil

	case ShellBash:
		bashSnippet := posixSnippet("bash")
		bashrc := filepath.Join(home, ".bashrc")
		hint := "source ~/.bashrc"

		targets := []completionTarget{
			{
				Label:       "Bash",
				ProfilePath: bashrc,
				Snippet:     bashSnippet,
				RefreshHint: hint,
				EOL:         "\n",
			},
		}

		if env.goos == "windows" {
			bashProfile := filepath.Join(home, ".bash_profile")
			bashLogin := filepath.Join(home, ".bash_login")
			profile := filepath.Join(home, ".profile")

			statExists := func(p string) bool {
				if env.statPath == nil {
					_, err := os.Stat(p)
					return err == nil
				}
				_, err := env.statPath(p)
				return err == nil
			}

			hasBP := statExists(bashProfile)
			hasBL := statExists(bashLogin)
			hasP := statExists(profile)

			targets[0].RefreshHint = "source ~/.bash_profile"

			if !hasBP && !hasBL && !hasP {
				targets = append(targets, completionTarget{
					Label:       "Git Bash login profile",
					ProfilePath: bashProfile,
					Snippet: "# generated by minhthetus-cli setup-completion\n" +
						"if [ -f ~/.bashrc ]; then . ~/.bashrc; fi",
					RefreshHint: "source ~/.bash_profile",
					EOL:         "\n",
					BashBridge:  true,
				})
			} else if hasBP {
				if data, err := os.ReadFile(bashProfile); err == nil && !bashProfileSourcesBashrc(string(data)) {
					targets = append(targets, completionTarget{
						Label:       "Git Bash login profile",
						ProfilePath: bashProfile,
						Snippet:     bashSnippet,
						RefreshHint: "source ~/.bash_profile",
						EOL:         "\n",
					})
				}
			}
		}
		return targets, nil

	case ShellZsh:
		return []completionTarget{
			{
				Label:       "Zsh",
				ProfilePath: filepath.Join(home, ".zshrc"),
				Snippet:     posixSnippet("zsh"),
				RefreshHint: "source ~/.zshrc",
				EOL:         "\n",
			},
		}, nil

	case ShellFish:
		return []completionTarget{
			{
				Label:       "Fish",
				ProfilePath: filepath.Join(xdgConfigOrDefault(home, env.getenv), "fish", "completions", "minhthetus-cli.fish"),
				Snippet:     fishSnippet(),
				RefreshHint: "exec fish",
				EOL:         "\n",
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported shell %q", shell)
	}
}

func shortenHomePath(path, home string) string {
	if home == "" {
		return path
	}
	cleanPath := filepath.Clean(path)
	cleanHome := filepath.Clean(home)
	if cleanPath == cleanHome {
		return "~"
	}
	prefix := cleanHome + string(filepath.Separator)
	if strings.HasPrefix(cleanPath, prefix) {
		return "~" + string(filepath.Separator) + strings.TrimPrefix(cleanPath, prefix)
	}
	// Handle slash differences when testing Windows paths on Unix
	normPath := strings.ReplaceAll(cleanPath, "\\", "/")
	normHome := strings.TrimSuffix(strings.ReplaceAll(cleanHome, "\\", "/"), "/")
	if strings.HasPrefix(normPath, normHome+"/") {
		return "~/" + strings.TrimPrefix(normPath, normHome+"/")
	}
	return path
}

func isShellDelimiterOrSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\r', '\n', ';', '&', '|', '(', ')', '{', '}':
		return true
	default:
		return false
	}
}

func stripShellLineComment(line string) string {
	inSingle := false
	inDouble := false
	escaped := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			escaped = false
			continue
		}
		if inSingle {
			if ch == '\'' {
				inSingle = false
			}
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if inDouble {
			if ch == '"' {
				inDouble = false
			}
			continue
		}
		if ch == '\'' {
			inSingle = true
			continue
		}
		if ch == '"' {
			inDouble = true
			continue
		}
		if ch == '#' && (i == 0 || isShellDelimiterOrSpace(line[i-1])) {
			return line[:i]
		}
	}
	return line
}

// bashProfileSourcesBashrc reports whether a profile references .bashrc on an active (non-comment) line.
func bashProfileSourcesBashrc(content string) bool {
	norm := strings.ReplaceAll(content, "\r\n", "\n")
	for _, line := range strings.Split(norm, "\n") {
		if strings.Contains(stripShellLineComment(line), ".bashrc") {
			return true
		}
	}
	return false
}

func hasCompletionInstalled(content string, target completionTarget) bool {
	if target.BashBridge {
		return bashProfileSourcesBashrc(content) || strings.Contains(content, "minhthetus-cli completion")
	}
	return strings.Contains(content, completionBlockStart) || strings.Contains(content, "minhthetus-cli completion")
}

func appendSnippetToFile(target completionTarget) (alreadyInstalled bool, err error) {
	existingBytes, readErr := os.ReadFile(target.ProfilePath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return false, readErr
	}

	existing := string(existingBytes)
	if readErr == nil && hasCompletionInstalled(existing, target) {
		return true, nil
	}

	if err := os.MkdirAll(filepath.Dir(target.ProfilePath), 0755); err != nil {
		return false, err
	}

	// Preserve existing line endings on PowerShell profiles if already present
	eol := target.EOL
	if eol == "" {
		eol = "\n"
	}
	if len(existingBytes) > 0 && target.EOL == "\r\n" && !strings.Contains(existing, "\r\n") && strings.Contains(existing, "\n") {
		eol = "\n"
	}

	snippet := strings.ReplaceAll(target.Snippet, "\r\n", "\n")
	if eol != "\n" {
		snippet = strings.ReplaceAll(snippet, "\n", eol)
	}

	var payload string
	if len(existingBytes) == 0 {
		payload = snippet + eol
	} else {
		if strings.HasSuffix(existing, "\n") {
			payload = eol + snippet + eol
		} else {
			payload = eol + eol + snippet + eol
		}
	}

	f, err := os.OpenFile(target.ProfilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if _, err := f.WriteString(payload); err != nil {
		return false, err
	}
	return false, nil
}

func setupCompletionWithEnv(silent bool, shellFlag string, env completionEnv) error {
	out := env.out
	if out == nil {
		out = os.Stdout
	}

	var targetShell ShellType
	var psFlavor string

	if strings.TrimSpace(shellFlag) != "" {
		parsed, err := parseShellFlag(shellFlag)
		if err != nil {
			fmt.Fprintln(out, ui.ErrorMessage(err.Error()))
			return err
		}
		targetShell = parsed
		// If current shell happens to match PowerShell, preserve detected flavor
		if detected, flavor := detectShell(env); detected == ShellPowerShell {
			psFlavor = flavor
		}
	} else {
		targetShell, psFlavor = detectShell(env)
	}

	if targetShell == ShellUnknown || targetShell == ShellCmd {
		if silent {
			return nil
		}
		if targetShell == ShellCmd {
			fmt.Fprintln(out, ui.WarningMessage("Windows Command Prompt (cmd.exe) does not support native tab completion."))
		} else {
			fmt.Fprintln(out, ui.WarningMessage("Could not automatically detect your shell."))
		}

		if env.isTTY != nil && env.isTTY() && env.chooseShell != nil {
			choice, err := env.chooseShell(
				"Select a shell to configure completion for:",
				[]string{"powershell", "bash", "zsh", "fish", "Cancel"},
			)
			if err != nil || choice == "" || choice == "Cancel" {
				fmt.Fprintln(out, ui.InfoMessage("Completion setup cancelled."))
				return nil
			}
			targetShell, _ = parseShellFlag(choice)
		} else {
			fmt.Fprintln(out, ui.HintMessage("Run setup-completion with --shell to configure a specific shell:"))
			for _, ex := range []string{
				"minhthetus-cli setup-completion --shell powershell",
				"minhthetus-cli setup-completion --shell bash",
				"minhthetus-cli setup-completion --shell zsh",
				"minhthetus-cli setup-completion --shell fish",
			} {
				fmt.Fprintf(out, "   %s\n", ui.CyanStyle().Render(ex))
			}
			return nil
		}
	}

	home, err := env.userHomeDir()
	if err != nil {
		if !silent {
			fmt.Fprintln(out, ui.ErrorMessage(fmt.Sprintf("Failed to locate home directory: %v", err)))
		}
		return err
	}

	targets, err := resolveCompletionTargets(targetShell, psFlavor, silent, env, home)
	if err != nil {
		if !silent {
			fmt.Fprintln(out, ui.ErrorMessage(err.Error()))
		}
		return err
	}

	if !silent {
		fmt.Fprintf(out, "%s Setting up completion for %s...\n", ui.CyanStyle().Render(ui.HourglassIcon), ui.BoldStyle.Render(string(targetShell)))
	}

	addedAny := false
	var lastHint string
	for _, target := range targets {
		displayPath := shortenHomePath(target.ProfilePath, home)
		already, err := appendSnippetToFile(target)
		if err != nil {
			if !silent {
				fmt.Fprintln(out, ui.ErrorMessage(fmt.Sprintf("Error updating %s: %v", displayPath, err)))
			}
			return err
		}
		if already {
			if !silent && !target.BashBridge {
				fmt.Fprintln(out, ui.SuccessMessage(fmt.Sprintf("Completion already exists in %s", displayPath)))
			}
		} else {
			addedAny = true
			lastHint = target.RefreshHint
			if !silent {
				if target.BashBridge {
					fmt.Fprintln(out, ui.SuccessMessage(fmt.Sprintf("Created %s to load ~/.bashrc in Git Bash login shells", displayPath)))
				} else {
					fmt.Fprintln(out, ui.SuccessMessage(fmt.Sprintf("Added completion to %s (%s)", displayPath, target.Label)))
				}
			}
		}
	}

	if addedAny && !silent && lastHint != "" {
		fmt.Fprintln(out, ui.HintMessage(fmt.Sprintf("Run '%s' or restart your shell to activate tab completion.", lastHint)))
	}
	return nil
}
