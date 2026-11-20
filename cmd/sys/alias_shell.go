package sys

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
)

const (
	aliasMarkerStart = "# >>> minhthetus-cli aliases >>>"
	aliasMarkerEnd   = "# <<< minhthetus-cli aliases <<<"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// ShellKind identifies the syntax dialect used by a target shell profile.
type ShellKind string

const (
	ShellKindBash       ShellKind = "bash"
	ShellKindZsh        ShellKind = "zsh"
	ShellKindPowerShell ShellKind = "powershell"
)

// ProfileTarget pairs a shell profile file path with its shell syntax kind.
type ProfileTarget struct {
	Path  string
	Shell ShellKind
}

func resolveUnixProfileTargets(home, shellEnv, goos string, fileExists func(string) bool) []ProfileTarget {
	if home == "" {
		return nil
	}

	zshrc := filepath.Join(home, ".zshrc")
	bashrc := filepath.Join(home, ".bashrc")
	bashProfile := filepath.Join(home, ".bash_profile")

	shellLower := strings.ToLower(shellEnv)
	var targets []ProfileTarget

	if strings.Contains(shellLower, "zsh") || fileExists(zshrc) {
		targets = append(targets, ProfileTarget{Path: zshrc, Shell: ShellKindZsh})
	}
	if strings.Contains(shellLower, "bash") || fileExists(bashrc) {
		targets = append(targets, ProfileTarget{Path: bashrc, Shell: ShellKindBash})
	}
	if fileExists(bashProfile) {
		targets = append(targets, ProfileTarget{Path: bashProfile, Shell: ShellKindBash})
	}

	if len(targets) == 0 {
		if goos == "darwin" {
			targets = append(targets, ProfileTarget{Path: zshrc, Shell: ShellKindZsh})
		} else {
			targets = append(targets, ProfileTarget{Path: bashrc, Shell: ShellKindBash})
		}
	}

	return deduplicateTargets(targets, false)
}

func resolveWindowsProfileTargets(
	home, docsDir, altDocsDir, psModulePath, msystem, shellEnv string,
	lookPath func(string) (string, error),
	fileExists func(string) bool,
) []ProfileTarget {
	var targets []ProfileTarget

	addPS51 := func(baseDocs string) {
		if baseDocs == "" {
			return
		}
		p := filepath.Join(baseDocs, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
		targets = append(targets, ProfileTarget{Path: p, Shell: ShellKindPowerShell})
	}
	addPS7 := func(baseDocs string) {
		if baseDocs == "" {
			return
		}
		p := filepath.Join(baseDocs, "PowerShell", "Microsoft.PowerShell_profile.ps1")
		targets = append(targets, ProfileTarget{Path: p, Shell: ShellKindPowerShell})
	}

	psModuleLower := strings.ToLower(psModulePath)

	primaryPS51 := filepath.Join(docsDir, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	_, psExeErr := lookPath("powershell.exe")
	if psExeErr != nil {
		_, psExeErr = lookPath("powershell")
	}
	if strings.Contains(psModuleLower, "windowspowershell") || fileExists(primaryPS51) || psExeErr == nil {
		addPS51(docsDir)
	}
	if altDocsDir != "" && !strings.EqualFold(filepath.Clean(altDocsDir), filepath.Clean(docsDir)) {
		altPS51 := filepath.Join(altDocsDir, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
		if fileExists(altPS51) {
			addPS51(altDocsDir)
		}
	}

	primaryPS7 := filepath.Join(docsDir, "PowerShell", "Microsoft.PowerShell_profile.ps1")
	_, pwshErr := lookPath("pwsh")
	if pwshErr != nil {
		_, pwshErr = lookPath("pwsh.exe")
	}
	if strings.Contains(psModuleLower, `\powershell\modules`) || strings.Contains(psModuleLower, `/powershell/modules`) || fileExists(primaryPS7) || pwshErr == nil {
		addPS7(docsDir)
	}
	if altDocsDir != "" && !strings.EqualFold(filepath.Clean(altDocsDir), filepath.Clean(docsDir)) {
		altPS7 := filepath.Join(altDocsDir, "PowerShell", "Microsoft.PowerShell_profile.ps1")
		if fileExists(altPS7) {
			addPS7(altDocsDir)
		}
	}

	if home != "" {
		bashrc := filepath.Join(home, ".bashrc")
		bashProfile := filepath.Join(home, ".bash_profile")
		_, bashExeErr := lookPath("bash.exe")
		gitBashDefault := fileExists(`C:\Program Files\Git\bin\bash.exe`) ||
			fileExists(`C:\Program Files (x86)\Git\bin\bash.exe`) ||
			fileExists(`C:\Program Files\Git\usr\bin\bash.exe`)
		_, gitErr := lookPath("git")

		hasGitBash := msystem != "" ||
			strings.Contains(strings.ToLower(shellEnv), "bash") ||
			fileExists(bashrc) ||
			fileExists(bashProfile) ||
			bashExeErr == nil ||
			(gitErr == nil && gitBashDefault)

		if hasGitBash {
			if msystem != "" || strings.Contains(strings.ToLower(shellEnv), "bash") || fileExists(bashrc) || !fileExists(bashProfile) {
				targets = append(targets, ProfileTarget{Path: bashrc, Shell: ShellKindBash})
			}
			if fileExists(bashProfile) {
				targets = append(targets, ProfileTarget{Path: bashProfile, Shell: ShellKindBash})
			}
		}
	}

	if len(targets) == 0 && docsDir != "" {
		addPS51(docsDir)
	}

	return deduplicateTargets(targets, true)
}

func deduplicateTargets(targets []ProfileTarget, caseInsensitive bool) []ProfileTarget {
	seen := make(map[string]bool, len(targets))
	out := make([]ProfileTarget, 0, len(targets))
	for _, t := range targets {
		cleaned := filepath.Clean(t.Path)
		key := cleaned
		if caseInsensitive {
			key = strings.ToLower(cleaned)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ProfileTarget{Path: cleaned, Shell: t.Shell})
	}
	return out
}

func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", `'\''`)
}

// RenderBashBlock renders the complete delimited alias block for Bash/Zsh profiles.
// Returns "" when aliases is empty.
func RenderBashBlock(aliases []AliasEntry) string {
	if len(aliases) == 0 {
		return ""
	}

	sorted := sortedAliases(aliases)
	var lines []string
	lines = append(lines, aliasMarkerStart)

	for _, a := range sorted {
		if a.Kind == aliasKindPresetMaster {
			lines = append(lines,
				fmt.Sprintf("unalias %s 2>/dev/null || true", a.Name),
				fmt.Sprintf("%s() {", a.Name),
				"  local branch",
				"  branch=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null)",
				`  branch="${branch#refs/remotes/origin/}"`,
				`  if [ -z "$branch" ]; then`,
				"    if git show-ref --verify --quiet refs/heads/main 2>/dev/null || git show-ref --verify --quiet refs/remotes/origin/main 2>/dev/null; then",
				`      branch="main"`,
				"    else",
				`      branch="master"`,
				"    fi",
				"  fi",
				`  git checkout "$branch" && git fetch origin && git pull origin "$branch"`,
				"}",
			)
			continue
		}
		lines = append(lines, fmt.Sprintf("alias %s='%s'", a.Name, escapeSingleQuotes(a.Command)))
	}

	lines = append(lines, aliasMarkerEnd)
	return strings.Join(lines, "\n")
}

type shellChainPart struct {
	Command string
	OpAfter string // "", "&&", or "||"
}

type psDelimSpan struct {
	ch    byte
	start int
}

func isPSLineCommentStart(cmd string, idx int) bool {
	if idx == 0 {
		return true
	}
	switch cmd[idx-1] {
	case ' ', '\t', '\r', '\n', ';', '|', '&', '(', '{', ')', '}', '<', '>':
		return true
	}
	return false
}

// skipPSBlockComment advances from '<' of "<#" to the closing '>' of "#>" (or end of string).
func skipPSBlockComment(cmd string, i int) int {
	for j := i + 2; j+1 < len(cmd); j++ {
		if cmd[j] == '#' && cmd[j+1] == '>' {
			return j + 1
		}
	}
	return len(cmd) - 1
}

// skipPSLineComment advances from '#' to the newline (or end of string).
func skipPSLineComment(cmd string, i int) int {
	for j := i + 1; j < len(cmd); j++ {
		if cmd[j] == '\n' {
			return j
		}
	}
	return len(cmd) - 1
}

func splitPowerShellChain(cmd string, ignoreParenDepth bool) []shellChainPart {
	var parts []shellChainPart
	inSingle := false
	inDouble := false
	var delims []byte
	start := 0

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]
		if inSingle {
			if ch == '\'' {
				if i+1 < len(cmd) && cmd[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			if ch == '`' {
				if i+1 < len(cmd) {
					i++
				}
				continue
			}
			if ch == '"' {
				if i+1 < len(cmd) && cmd[i+1] == '"' {
					i++
					continue
				}
				inDouble = false
			}
			continue
		}

		// Outside quotes in PowerShell: backtick (`) is escape, backslash (\) is literal.
		if ch == '`' {
			if i+1 < len(cmd) {
				i++
			}
			continue
		}
		if ch == '<' && i+1 < len(cmd) && cmd[i+1] == '#' {
			i = skipPSBlockComment(cmd, i)
			continue
		}
		if ch == '#' && isPSLineCommentStart(cmd, i) {
			i = skipPSLineComment(cmd, i)
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
		if !ignoreParenDepth {
			if ch == '(' || ch == '{' {
				delims = append(delims, ch)
				continue
			}
			if ch == ')' && len(delims) > 0 && delims[len(delims)-1] == '(' {
				delims = delims[:len(delims)-1]
				continue
			}
			if ch == '}' && len(delims) > 0 && delims[len(delims)-1] == '{' {
				delims = delims[:len(delims)-1]
				continue
			}
		}
		if (ignoreParenDepth || len(delims) == 0) && i+1 < len(cmd) {
			two := cmd[i : i+2]
			if two == "&&" || two == "||" {
				seg := strings.TrimSpace(cmd[start:i])
				parts = append(parts, shellChainPart{Command: seg, OpAfter: two})
				i++
				start = i + 1
			}
		}
	}

	tail := strings.TrimSpace(cmd[start:])
	parts = append(parts, shellChainPart{Command: tail})
	return parts
}

func validatePowerShellChainSyntax(cmd string) error {
	parts := splitPowerShellChain(cmd, false)
	if len(parts) <= 1 && parts[0].OpAfter == "" {
		return nil
	}
	for _, p := range parts {
		if strings.TrimSpace(p.Command) == "" {
			return fmt.Errorf("alias command cannot start, end, or have empty segments around '&&' or '||'")
		}
	}
	return nil
}

func hasUnquotedShellMetaPS(cmd string) bool {
	inSingle := false
	inDouble := false

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]
		if inSingle {
			if ch == '\'' {
				if i+1 < len(cmd) && cmd[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			if ch == '`' {
				if i+1 < len(cmd) {
					i++
				}
				continue
			}
			if ch == '"' {
				if i+1 < len(cmd) && cmd[i+1] == '"' {
					i++
					continue
				}
				inDouble = false
			}
			continue
		}

		if ch == '`' {
			if i+1 < len(cmd) {
				i++
			}
			continue
		}
		if ch == '<' && i+1 < len(cmd) && cmd[i+1] == '#' {
			return true
		}
		if ch == '#' && isPSLineCommentStart(cmd, i) {
			i = skipPSLineComment(cmd, i)
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
		switch ch {
		case ';', '|', '&', '>', '<', '(', ')', '{', '}':
			return true
		}
	}
	return false
}

func findUnquotedPSCommentIndex(cmd string) int {
	inSingle := false
	inDouble := false

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]
		if inSingle {
			if ch == '\'' {
				if i+1 < len(cmd) && cmd[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			if ch == '`' {
				if i+1 < len(cmd) {
					i++
				}
				continue
			}
			if ch == '"' {
				if i+1 < len(cmd) && cmd[i+1] == '"' {
					i++
					continue
				}
				inDouble = false
			}
			continue
		}

		if ch == '`' {
			if i+1 < len(cmd) {
				i++
			}
			continue
		}
		if ch == '<' && i+1 < len(cmd) && cmd[i+1] == '#' {
			i = skipPSBlockComment(cmd, i)
			continue
		}
		if ch == '#' && isPSLineCommentStart(cmd, i) {
			return i
		}
		if ch == '\'' {
			inSingle = true
			continue
		}
		if ch == '"' {
			inDouble = true
			continue
		}
	}
	return -1
}

func hasUnquotedChainOpPS(cmd string) bool {
	parts := splitPowerShellChain(cmd, true)
	return len(parts) > 1 || (len(parts) == 1 && parts[0].OpAfter != "")
}

func isChainPartsNonTrivial(parts []shellChainPart) bool {
	return len(parts) > 1 || (len(parts) == 1 && parts[0].OpAfter != "")
}

func rewriteNestedPowerShellChains(expr string, indent string) string {
	var out strings.Builder
	inSingle := false
	inDouble := false
	var delims []psDelimSpan
	lastCopied := 0

	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if inSingle {
			if ch == '\'' {
				if i+1 < len(expr) && expr[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			if ch == '`' {
				if i+1 < len(expr) {
					i++
				}
				continue
			}
			if ch == '"' {
				if i+1 < len(expr) && expr[i+1] == '"' {
					i++
					continue
				}
				inDouble = false
			}
			continue
		}

		if ch == '`' {
			if i+1 < len(expr) {
				i++
			}
			continue
		}
		if ch == '<' && i+1 < len(expr) && expr[i+1] == '#' {
			i = skipPSBlockComment(expr, i)
			continue
		}
		if ch == '#' && isPSLineCommentStart(expr, i) {
			i = skipPSLineComment(expr, i)
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

		if ch == '(' || ch == '{' {
			delims = append(delims, psDelimSpan{ch: ch, start: i})
			continue
		}
		if (ch == ')' && len(delims) > 0 && delims[len(delims)-1].ch == '(') ||
			(ch == '}' && len(delims) > 0 && delims[len(delims)-1].ch == '{') {
			span := delims[len(delims)-1]
			delims = delims[:len(delims)-1]
			if len(delims) == 0 {
				inner := expr[span.start+1 : i]
				innerParts := splitPowerShellChain(inner, false)
				if isChainPartsNonTrivial(innerParts) {
					out.WriteString(expr[lastCopied:span.start])
					innerIndent := indent + "    "
					bodyLines := emitPSChainLines(innerParts, innerIndent)
					if span.ch == '{' {
						out.WriteString("{\n")
						out.WriteString(strings.Join(bodyLines, "\n"))
						out.WriteString("\n" + indent + "}")
					} else {
						isSubExprOrArray := span.start > 0 && (expr[span.start-1] == '$' || expr[span.start-1] == '@')
						if isSubExprOrArray {
							out.WriteString("(\n")
						} else {
							out.WriteString("$(\n")
						}
						out.WriteString(strings.Join(bodyLines, "\n"))
						out.WriteString("\n" + indent + ")")
					}
					lastCopied = i + 1
				} else {
					rewrittenInner := rewriteNestedPowerShellChains(inner, indent)
					if rewrittenInner != inner {
						out.WriteString(expr[lastCopied : span.start+1])
						out.WriteString(rewrittenInner)
						out.WriteByte(ch)
						lastCopied = i + 1
					}
				}
			}
		}
	}

	out.WriteString(expr[lastCopied:])
	return out.String()
}

func emitPSChainLines(parts []shellChainPart, indent string) []string {
	var lines []string
	hasPrev := false
	prevOp := ""

	for _, part := range parts {
		cmd := strings.TrimSpace(part.Command)
		if cmd != "" {
			stepIndent := indent
			if hasPrev && (prevOp == "&&" || prevOp == "||") {
				stepIndent = indent + "    "
			}
			rewrittenCmd := rewriteNestedPowerShellChains(cmd, stepIndent)
			if !hasPrev {
				lines = append(lines,
					fmt.Sprintf("%s$global:LASTEXITCODE = $null", indent),
					fmt.Sprintf("%s%s", indent, rewrittenCmd),
					fmt.Sprintf("%sif ($null -ne $LASTEXITCODE) { $_ok = ($LASTEXITCODE -eq 0) } else { $_ok = $? }", indent),
				)
				hasPrev = true
			} else if prevOp == "&&" {
				lines = append(lines,
					fmt.Sprintf("%sif ($_ok) {", indent),
					fmt.Sprintf("%s$global:LASTEXITCODE = $null", stepIndent),
					fmt.Sprintf("%s%s", stepIndent, rewrittenCmd),
					fmt.Sprintf("%sif ($null -ne $LASTEXITCODE) { $_ok = ($LASTEXITCODE -eq 0) } else { $_ok = $? }", stepIndent),
					fmt.Sprintf("%s}", indent),
				)
			} else if prevOp == "||" {
				lines = append(lines,
					fmt.Sprintf("%sif (-not $_ok) {", indent),
					fmt.Sprintf("%s$global:LASTEXITCODE = $null", stepIndent),
					fmt.Sprintf("%s%s", stepIndent, rewrittenCmd),
					fmt.Sprintf("%sif ($null -ne $LASTEXITCODE) { $_ok = ($LASTEXITCODE -eq 0) } else { $_ok = $? }", stepIndent),
					fmt.Sprintf("%s}", indent),
				)
			}
		}
		prevOp = part.OpAfter
	}
	return lines
}

func formatSinglePSCommand(trimmed string) string {
	rewritten := rewriteNestedPowerShellChains(trimmed, "    ")
	if commentIdx := findUnquotedPSCommentIndex(rewritten); commentIdx >= 0 {
		before := strings.TrimRight(rewritten[:commentIdx], " \t")
		comment := rewritten[commentIdx:]
		if before == "" || hasUnquotedShellMetaPS(before) || strings.HasSuffix(before, "@args") {
			return rewritten
		}
		return before + " @args " + comment
	}
	if hasUnquotedShellMetaPS(rewritten) || strings.HasSuffix(strings.TrimSpace(rewritten), "@args") {
		return rewritten
	}
	return rewritten + " @args"
}

// RenderPowerShellBlock renders the complete delimited alias block compatible with both
// Windows PowerShell 5.1 and PowerShell 7+. Returns "" when aliases is empty.
func RenderPowerShellBlock(aliases []AliasEntry) string {
	if len(aliases) == 0 {
		return ""
	}

	sorted := sortedAliases(aliases)
	var lines []string
	lines = append(lines, aliasMarkerStart)

	for _, a := range sorted {
		lines = append(lines, fmt.Sprintf("Remove-Item Alias:%s -Force -ErrorAction SilentlyContinue", a.Name))

		if a.Kind == aliasKindPresetMaster {
			lines = append(lines,
				fmt.Sprintf("function %s {", a.Name),
				"    $branch = (git symbolic-ref refs/remotes/origin/HEAD 2>$null)",
				"    if ($LASTEXITCODE -eq 0 -and $branch) {",
				"        $branch = ($branch.Trim() -replace '^refs/remotes/origin/', '')",
				"    } else {",
				"        $branch = $null",
				"    }",
				"    if (-not $branch) {",
				"        git show-ref --verify --quiet refs/heads/main 2>$null",
				"        $hasLocalMain = ($LASTEXITCODE -eq 0)",
				"        git show-ref --verify --quiet refs/remotes/origin/main 2>$null",
				"        $hasRemoteMain = ($LASTEXITCODE -eq 0)",
				"        if ($hasLocalMain -or $hasRemoteMain) {",
				"            $branch = 'main'",
				"        } else {",
				"            $branch = 'master'",
				"        }",
				"    }",
				"    git checkout $branch",
				"    if ($LASTEXITCODE -ne 0) { return }",
				"    git fetch origin",
				"    if ($LASTEXITCODE -ne 0) { return }",
				"    git pull origin $branch",
				"}",
			)
			continue
		}

		trimmed := strings.TrimSpace(a.Command)
		if strings.EqualFold(trimmed, "clear") || strings.EqualFold(trimmed, "Clear-Host") {
			lines = append(lines,
				fmt.Sprintf("function %s {", a.Name),
				"    Clear-Host",
				"}",
			)
			continue
		}

		parts := splitPowerShellChain(trimmed, false)
		if !isChainPartsNonTrivial(parts) {
			single := formatSinglePSCommand(trimmed)
			if !hasUnquotedChainOpPS(single) {
				lines = append(lines,
					fmt.Sprintf("function %s {", a.Name),
					fmt.Sprintf("    %s", single),
					"}",
				)
				continue
			}
			// Fallback if unbalanced parentheses/braces hid an unquoted && or ||.
			parts = splitPowerShellChain(trimmed, true)
		}

		lines = append(lines, fmt.Sprintf("function %s {", a.Name))
		lines = append(lines, emitPSChainLines(parts, "    ")...)
		lines = append(lines, "}")
	}

	lines = append(lines, aliasMarkerEnd)
	return strings.Join(lines, "\n")
}

func renderBlockForShell(shell ShellKind, aliases []AliasEntry) string {
	if shell == ShellKindPowerShell {
		return RenderPowerShellBlock(aliases)
	}
	return RenderBashBlock(aliases)
}

func splitNormalizedLines(content string) []string {
	if content == "" {
		return nil
	}
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// UpsertMarkerBlock inserts, updates, or removes the delimited alias block in existing content.
// Returns the updated content, whether a change occurred, and any marker-corruption error.
func UpsertMarkerBlock(existing string, block string, lineEnding string) (string, bool, error) {
	if lineEnding == "" {
		lineEnding = "\n"
	}

	lines := splitNormalizedLines(existing)
	startIdx := -1
	endIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == aliasMarkerStart {
			if startIdx >= 0 {
				return "", false, fmt.Errorf("corrupted alias marker block: duplicate start marker")
			}
			startIdx = i
		} else if trimmed == aliasMarkerEnd {
			if endIdx >= 0 {
				return "", false, fmt.Errorf("corrupted alias marker block: duplicate end marker")
			}
			endIdx = i
		}
	}

	if (startIdx >= 0 && endIdx < 0) || (startIdx < 0 && endIdx >= 0) {
		return "", false, fmt.Errorf("corrupted alias marker block: missing matching start/end marker")
	}
	if startIdx >= 0 && endIdx < startIdx {
		return "", false, fmt.Errorf("corrupted alias marker block: end marker appears before start marker")
	}

	if strings.TrimSpace(block) == "" {
		if startIdx < 0 {
			return existing, false, nil
		}
		before := append([]string{}, lines[:startIdx]...)
		after := append([]string{}, lines[endIdx+1:]...)
		// Drop the separator blank line we added before the block when appropriate.
		if len(before) > 0 && strings.TrimSpace(before[len(before)-1]) == "" {
			if len(after) == 0 || strings.TrimSpace(after[0]) == "" {
				before = before[:len(before)-1]
			}
		}
		for len(before) == 0 && len(after) > 0 && strings.TrimSpace(after[0]) == "" {
			after = after[1:]
		}
		combined := append(before, after...)
		if len(combined) == 0 {
			return "", existing != "", nil
		}
		out := strings.Join(combined, lineEnding) + lineEnding
		return out, out != existing, nil
	}

	blockLines := splitNormalizedLines(block)
	var updated []string
	if startIdx >= 0 {
		updated = make([]string, 0, len(lines)-(endIdx-startIdx+1)+len(blockLines))
		updated = append(updated, lines[:startIdx]...)
		updated = append(updated, blockLines...)
		updated = append(updated, lines[endIdx+1:]...)
	} else {
		updated = append(updated, lines...)
		if len(updated) > 0 && strings.TrimSpace(updated[len(updated)-1]) != "" {
			updated = append(updated, "")
		}
		updated = append(updated, blockLines...)
	}

	out := strings.Join(updated, lineEnding) + lineEnding
	return out, out != existing, nil
}

func detectLineEnding(existing string, shell ShellKind) string {
	if strings.Contains(existing, "\r\n") {
		return "\r\n"
	}
	if strings.Contains(existing, "\n") {
		return "\n"
	}
	if shell == ShellKindPowerShell {
		return "\r\n"
	}
	return "\n"
}

func resolveProfileWritePath(path string) string {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return path
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil && resolved != "" {
		return resolved
	}
	if linkTarget, err := os.Readlink(path); err == nil && linkTarget != "" {
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Join(filepath.Dir(path), linkTarget)
		}
		return filepath.Clean(linkTarget)
	}
	return path
}

func syncProfileFile(target ProfileTarget, aliases []AliasEntry) (bool, error) {
	block := renderBlockForShell(target.Shell, aliases)
	writePath := resolveProfileWritePath(target.Path)

	raw, err := os.ReadFile(writePath)
	fileExists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to read %s: %w", target.Path, err)
	}

	if !fileExists && block == "" {
		return false, nil
	}

	hasBOM := bytes.HasPrefix(raw, utf8BOM)
	content := string(raw)
	if hasBOM {
		content = string(raw[len(utf8BOM):])
	}

	lineEnding := detectLineEnding(content, target.Shell)
	nextContent, changed, err := UpsertMarkerBlock(content, block, lineEnding)
	if err != nil {
		return false, fmt.Errorf("%s: %w", target.Path, err)
	}
	if !changed {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(writePath), 0755); err != nil {
		return false, fmt.Errorf("failed to create directory for %s: %w", target.Path, err)
	}

	mode := os.FileMode(0644)
	if fileExists {
		if info, statErr := os.Stat(writePath); statErr == nil {
			mode = info.Mode()
		}
	}

	var outBytes []byte
	if hasBOM {
		outBytes = append(outBytes, utf8BOM...)
	}
	outBytes = append(outBytes, []byte(nextContent)...)

	tmpFile, err := os.CreateTemp(filepath.Dir(writePath), ".minhthetus-alias-*")
	if err != nil {
		return false, fmt.Errorf("failed to create temp file for %s: %w", target.Path, err)
	}
	tmpPath := tmpFile.Name()

	writeOK := false
	defer func() {
		if !writeOK {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(outBytes); err != nil {
		_ = tmpFile.Close()
		return false, fmt.Errorf("failed to write temp file for %s: %w", target.Path, err)
	}
	if err := tmpFile.Close(); err != nil {
		return false, fmt.Errorf("failed to close temp file for %s: %w", target.Path, err)
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return false, fmt.Errorf("failed to set permissions on %s: %w", target.Path, err)
	}
	if err := os.Rename(tmpPath, writePath); err != nil {
		return false, fmt.Errorf("failed to replace %s: %w", target.Path, err)
	}
	writeOK = true
	return true, nil
}

func syncAllProfiles(aliases []AliasEntry) ([]ProfileTarget, error) {
	targets := resolveCurrentProfileTargets()
	var updated []ProfileTarget
	var errs []string

	for _, t := range targets {
		changed, err := syncProfileFile(t, aliases)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		if changed {
			updated = append(updated, t)
		}
	}

	if len(errs) > 0 {
		return updated, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return updated, nil
}

func defaultReloadCommands(goos, shellEnv, msystem, psModulePath string, updated []ProfileTarget) []string {
	shellLower := strings.ToLower(shellEnv)
	if goos == "windows" {
		if msystem != "" || strings.Contains(shellLower, "bash") {
			for _, t := range updated {
				if t.Shell == ShellKindBash && strings.HasSuffix(t.Path, ".bash_profile") {
					return []string{"source ~/.bash_profile"}
				}
			}
			return []string{"source ~/.bashrc"}
		}
		if psModulePath != "" {
			return []string{". $PROFILE"}
		}
		var cmds []string
		hasPS := false
		hasBash := false
		for _, t := range updated {
			if t.Shell == ShellKindPowerShell && !hasPS {
				cmds = append(cmds, ". $PROFILE")
				hasPS = true
			} else if t.Shell == ShellKindBash && !hasBash {
				cmds = append(cmds, "source ~/.bashrc")
				hasBash = true
			}
		}
		if len(cmds) == 0 {
			return []string{". $PROFILE"}
		}
		return cmds
	}

	if strings.Contains(shellLower, "zsh") {
		return []string{"source ~/.zshrc"}
	}
	if strings.Contains(shellLower, "bash") {
		hasBashrc := false
		hasBashProfile := false
		for _, t := range updated {
			if strings.HasSuffix(t.Path, ".bashrc") {
				hasBashrc = true
			} else if strings.HasSuffix(t.Path, ".bash_profile") {
				hasBashProfile = true
			}
		}
		if !hasBashrc && hasBashProfile {
			return []string{"source ~/.bash_profile"}
		}
		return []string{"source ~/.bashrc"}
	}
	if goos == "darwin" {
		return []string{"source ~/.zshrc"}
	}
	return []string{"source ~/.bashrc"}
}

func printReloadHint(updated []ProfileTarget) {
	cmds := defaultReloadCommands(
		runtime.GOOS,
		os.Getenv("SHELL"),
		os.Getenv("MSYSTEM"),
		os.Getenv("PSModulePath"),
		updated,
	)
	fmt.Println(ui.InfoMessage("Reload your shell to apply alias changes in the current session:"))
	for _, cmd := range cmds {
		fmt.Printf("  %s\n", ui.CyanStyle().Render(cmd))
	}
	fmt.Println(ui.GrayStyle().Render("  (Or open a new terminal session for aliases to load automatically.)"))
}
