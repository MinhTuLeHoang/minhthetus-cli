package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
	"github.com/spf13/cobra"
)

// InstallMethod represents the way the CLI was installed
type InstallMethod string

const (
	Homebrew    InstallMethod = "Homebrew"
	GoInstall   InstallMethod = "Go"
	WinGet      InstallMethod = "WinGet"
	ManualBuild InstallMethod = "Manual/Makefile"
)

type cleanupTarget struct {
	Path       string
	DeleteFile bool
}

type uninstallDeps struct {
	env          completionEnv
	exePath      func() (string, error)
	confirm      func(prompt string) (bool, error)
	runBrew      func() error
	runWinget    func() error
	removeBinary func(exePath string) (scheduled bool, err error)
	out          io.Writer
}

func defaultUninstallDeps() uninstallDeps {
	env := defaultCompletionEnv()
	return uninstallDeps{
		env:     env,
		exePath: os.Executable,
		confirm: func(prompt string) (bool, error) {
			return ui.ConfirmWithDefault(prompt, 0, false, false)
		},
		runBrew:   runBrewUninstall,
		runWinget: runWingetUninstall,
		removeBinary: func(exePath string) (bool, error) {
			err := os.Remove(exePath)
			if err == nil {
				return false, nil
			}
			if env.goos == "windows" {
				if schedErr := scheduleWindowsSelfDelete(exePath); schedErr == nil {
					return true, nil
				}
			}
			return false, err
		},
		out: os.Stdout,
	}
}

func detectInstallMethodDirect(path string) InstallMethod {
	lowerPath := strings.ToLower(path)
	if strings.Contains(path, "Cellar") || strings.Contains(lowerPath, "homebrew") || strings.Contains(lowerPath, "opt/homebrew") {
		return Homebrew
	}
	if strings.Contains(lowerPath, "winget") {
		return WinGet
	}
	if strings.Contains(lowerPath, "go/bin") || strings.Contains(lowerPath, `go\bin`) {
		return GoInstall
	}
	return ManualBuild
}

// detectInstallMethod inspects the executable path (and symlink target) to determine the installation type.
func detectInstallMethod(path string) InstallMethod {
	if m := detectInstallMethodDirect(path); m != ManualBuild {
		return m
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil && resolved != path {
		return detectInstallMethodDirect(resolved)
	}
	return ManualBuild
}

// runBrewUninstall triggers Homebrew's uninstallation database command
func runBrewUninstall() error {
	brewCmd := exec.Command("brew", "uninstall", "minhthetus-cli")
	brewCmd.Stdout = os.Stdout
	brewCmd.Stderr = os.Stderr
	brewCmd.Stdin = os.Stdin
	return brewCmd.Run()
}

// runWingetUninstall triggers WinGet's uninstallation database command
func runWingetUninstall() error {
	wingetCmd := exec.Command("winget", "uninstall", "MinhTuLeHoang.minhthetus-cli")
	wingetCmd.Stdout = os.Stdout
	wingetCmd.Stderr = os.Stderr
	wingetCmd.Stdin = os.Stdin
	return wingetCmd.Run()
}

type psScanState struct {
	inSingle       bool
	inDouble       bool
	inHereSingle   bool
	inHereDouble   bool
	inBlockComment bool
}

func (s *psScanState) inMultiline() bool {
	return s.inSingle || s.inDouble || s.inHereSingle || s.inHereDouble || s.inBlockComment
}

func (s *psScanState) scanLine(line string) (opens int, closes int, tailHasElse bool) {
	i := 0
	if s.inHereSingle {
		if strings.HasPrefix(line, "'@") {
			s.inHereSingle = false
			i = 2
		} else {
			return 0, 0, false
		}
	} else if s.inHereDouble {
		if strings.HasPrefix(line, "\"@") {
			s.inHereDouble = false
			i = 2
		} else {
			return 0, 0, false
		}
	}

	escaped := false
	lastCloseIdx := -1
	for ; i < len(line); i++ {
		ch := line[i]
		if s.inBlockComment {
			if ch == '#' && i+1 < len(line) && line[i+1] == '>' {
				s.inBlockComment = false
				i++
			}
			continue
		}
		if escaped {
			escaped = false
			continue
		}
		if s.inSingle {
			if ch == '\'' {
				if i+1 < len(line) && line[i+1] == '\'' {
					i++
					continue
				}
				s.inSingle = false
			}
			continue
		}
		if s.inDouble {
			if ch == '`' {
				escaped = true
				continue
			}
			if ch == '"' {
				s.inDouble = false
			}
			continue
		}

		// Outside all strings/comments
		if ch == '@' && i+1 < len(line) && (line[i+1] == '"' || line[i+1] == '\'') && strings.TrimRight(line[i+2:], " \t") == "" {
			if line[i+1] == '"' {
				s.inHereDouble = true
			} else {
				s.inHereSingle = true
			}
			break
		}
		if ch == '<' && i+1 < len(line) && line[i+1] == '#' {
			s.inBlockComment = true
			i++
			continue
		}
		if ch == '#' {
			break
		}
		if ch == '`' {
			escaped = true
			continue
		}
		if ch == '\'' {
			s.inSingle = true
			continue
		}
		if ch == '"' {
			s.inDouble = true
			continue
		}
		if ch == '{' {
			opens++
			lastCloseIdx = -1
		} else if ch == '}' {
			closes++
			lastCloseIdx = i
		}
	}

	if lastCloseIdx != -1 && !s.inMultiline() {
		tail := strings.TrimSpace(line[lastCloseIdx+1:])
		if idx := strings.IndexByte(tail, '#'); idx != -1 {
			tail = strings.TrimSpace(tail[:idx])
		}
		lowerTail := strings.ToLower(tail)
		if lowerTail == "else" || strings.HasPrefix(lowerTail, "else ") ||
			strings.HasPrefix(lowerTail, "elseif ") || strings.HasPrefix(lowerTail, "elseif(") {
			tailHasElse = true
		}
	}
	return opens, closes, tailHasElse
}

func isPSLineBlankOrComment(line string, st *psScanState) bool {
	if st.inHereSingle || st.inHereDouble {
		st.scanLine(line)
		return true
	}
	wasBlockComment := st.inBlockComment
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || (!wasBlockComment && !st.inSingle && !st.inDouble && strings.HasPrefix(trimmed, "#")) {
		if wasBlockComment {
			st.scanLine(line)
		}
		return true
	}
	if !wasBlockComment && !st.inSingle && !st.inDouble && strings.HasPrefix(trimmed, "<#") {
		o, c, _ := st.scanLine(line)
		if o == 0 && c == 0 {
			return true
		}
	}
	if wasBlockComment {
		o, c, _ := st.scanLine(line)
		if o == 0 && c == 0 {
			return true
		}
	}
	return false
}

func isPSElseHeader(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	return lower == "else" || strings.HasPrefix(lower, "else ") || strings.HasPrefix(lower, "else{") ||
		strings.HasPrefix(lower, "elseif ") || strings.HasPrefix(lower, "elseif(")
}

// scanPSIfBlock finds the closing line index of a PowerShell if (...) { ... } block (including any attached elseif/else).
func scanPSIfBlock(lines []string, startIdx int) (endIdx int, ok bool) {
	var st psScanState
	opens, closes, tailHasElse := st.scanLine(lines[startIdx])
	if closes > 0 && opens == 0 {
		return -1, false
	}

	curIdx := startIdx
	depth := opens - closes
	openedBrace := opens > 0

	if !openedBrace {
		for k := startIdx + 1; k < len(lines); k++ {
			if isPSLineBlankOrComment(lines[k], &st) {
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(lines[k]), "{") {
				o2, c2, th2 := st.scanLine(lines[k])
				depth = o2 - c2
				openedBrace = o2 > 0
				tailHasElse = th2
				curIdx = k
			}
			break
		}
	}

	if !openedBrace {
		return -1, false
	}

	for depth > 0 || st.inMultiline() {
		curIdx++
		if curIdx >= len(lines) {
			return -1, false
		}
		oj, cj, thj := st.scanLine(lines[curIdx])
		depth += oj - cj
		tailHasElse = thj
	}

	// Consume attached elseif/else blocks so no orphan else/elseif remains
	for {
		if tailHasElse {
			foundOpen := false
			for k := curIdx + 1; k < len(lines); k++ {
				if isPSLineBlankOrComment(lines[k], &st) {
					continue
				}
				if strings.HasPrefix(strings.TrimSpace(lines[k]), "{") {
					o2, c2, th2 := st.scanLine(lines[k])
					depth = o2 - c2
					tailHasElse = th2
					curIdx = k
					foundOpen = o2 > 0
				}
				break
			}
			if !foundOpen {
				return -1, false
			}
			for depth > 0 || st.inMultiline() {
				curIdx++
				if curIdx >= len(lines) {
					return -1, false
				}
				oj, cj, thj := st.scanLine(lines[curIdx])
				depth += oj - cj
				tailHasElse = thj
			}
			continue
		}

		var probeSt psScanState = st
		nextCodeIdx := -1
		for k := curIdx + 1; k < len(lines); k++ {
			if isPSLineBlankOrComment(lines[k], &probeSt) {
				continue
			}
			if isPSElseHeader(lines[k]) {
				nextCodeIdx = k
			}
			break
		}
		if nextCodeIdx == -1 {
			break
		}

		st = probeSt
		oe, ce, the := st.scanLine(lines[nextCodeIdx])
		curIdx = nextCodeIdx
		depth = oe - ce
		tailHasElse = the
		if oe == 0 {
			foundOpen := false
			for k := curIdx + 1; k < len(lines); k++ {
				if isPSLineBlankOrComment(lines[k], &st) {
					continue
				}
				if strings.HasPrefix(strings.TrimSpace(lines[k]), "{") {
					o2, c2, th2 := st.scanLine(lines[k])
					depth = o2 - c2
					tailHasElse = th2
					curIdx = k
					foundOpen = o2 > 0
				}
				break
			}
			if !foundOpen {
				return -1, false
			}
		}
		for depth > 0 || st.inMultiline() {
			curIdx++
			if curIdx >= len(lines) {
				return -1, false
			}
			oj, cj, thj := st.scanLine(lines[curIdx])
			depth += oj - cj
			tailHasElse = thj
		}
	}

	return curIdx, true
}

type shellScanState struct {
	inSingle bool
	inDouble bool
}

func isWordBoundaryDelim(ch byte) bool {
	switch ch {
	case ' ', '\t', '\r', '\n', '#', ';', '|', '&', '(', ')', '{', '}', '\'', '"', '\\':
		return true
	default:
		return false
	}
}

func (s *shellScanState) scanLine(line string) (ifCount int, fiCount int, hasThen bool, firstToken string) {
	cmdPos := !s.inSingle && !s.inDouble
	for i := 0; i < len(line); {
		ch := line[i]
		if s.inSingle {
			if ch == '\'' {
				s.inSingle = false
			}
			i++
			continue
		}
		if s.inDouble {
			if ch == '\\' && i+1 < len(line) {
				i += 2
				continue
			}
			if ch == '"' {
				s.inDouble = false
			}
			i++
			continue
		}
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}
		if ch == '#' {
			break
		}
		if ch == '\\' {
			if i+1 < len(line) {
				i += 2
			} else {
				i++
			}
			cmdPos = false
			continue
		}
		if ch == '\'' {
			s.inSingle = true
			cmdPos = false
			i++
			continue
		}
		if ch == '"' {
			s.inDouble = true
			cmdPos = false
			i++
			continue
		}
		if ch == ';' || ch == '|' || ch == '&' || ch == '(' || ch == '{' {
			cmdPos = true
			i++
			continue
		}
		if ch == ')' || ch == '}' {
			cmdPos = false
			i++
			continue
		}

		start := i
		for i < len(line) && !isWordBoundaryDelim(line[i]) {
			i++
		}
		if i == start {
			i++
			continue
		}
		word := strings.ToLower(line[start:i])
		if firstToken == "" {
			firstToken = word
		}
		if cmdPos {
			switch word {
			case "if":
				ifCount++
				cmdPos = true
			case "then":
				hasThen = true
				cmdPos = true
			case "else", "elif", "do", "!":
				cmdPos = true
			case "fi":
				fiCount++
				cmdPos = false
			default:
				cmdPos = false
			}
		}
	}
	return ifCount, fiCount, hasThen, firstToken
}

func isBashIfThenLine(lines []string, idx int) bool {
	var st shellScanState
	ifCount, _, hasThen, firstToken := st.scanLine(lines[idx])
	if firstToken != "if" || ifCount == 0 {
		return false
	}
	if hasThen {
		return true
	}
	if !st.inSingle && !st.inDouble {
		code := extractExecutableCode(lines[idx], true)
		if strings.Count(code, "{") != strings.Count(code, "}") {
			return false
		}
	}
	for k := idx + 1; k < len(lines); k++ {
		wasInQuote := st.inSingle || st.inDouble
		trimmed := strings.TrimSpace(lines[k])
		if !wasInQuote && (trimmed == "" || strings.HasPrefix(trimmed, "#")) {
			continue
		}
		if !wasInQuote && (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "}")) {
			return false
		}
		ifc, fic, hasThenK, firstTokK := st.scanLine(lines[k])
		if !wasInQuote {
			if firstTokK == "then" {
				return true
			}
			if firstTokK == "if" || firstTokK == "end" || firstTokK == "fi" || firstTokK == "else" || firstTokK == "elif" {
				return false
			}
			if hasThenK && ifc == 0 {
				return true
			}
			if fic > ifc {
				return false
			}
			if !st.inSingle && !st.inDouble {
				code := extractExecutableCode(lines[k], true)
				if strings.Count(code, "{") != strings.Count(code, "}") {
					return false
				}
			}
		}
		if hasThenK && !wasInQuote && ifc > 0 {
			// Check if an outer `then` precedes a nested `if` on the same line (e.g. `[ -n "$B" ]; then if ...`).
			if strings.Contains(strings.ToLower(extractExecutableCode(lines[k], true)), "; then") {
				return true
			}
		}
	}
	return false
}

func scanBashIfBlock(lines []string, startIdx int) (endIdx int, ok bool) {
	var st shellScanState
	ifCount, fiCount, _, _ := st.scanLine(lines[startIdx])
	depth := ifCount - fiCount
	if depth <= 0 && !st.inSingle && !st.inDouble {
		return startIdx, true
	}
	for j := startIdx + 1; j < len(lines); j++ {
		ifc, fic, _, _ := st.scanLine(lines[j])
		depth += ifc - fic
		if depth <= 0 && !st.inSingle && !st.inDouble {
			return j, true
		}
	}
	return -1, false
}

type fishScanState struct {
	inSingle bool
	inDouble bool
}

func (s *fishScanState) scanLine(line string) (opens int, closes int, firstToken string) {
	cmdPos := !s.inSingle && !s.inDouble
	for i := 0; i < len(line); {
		ch := line[i]
		if s.inSingle {
			if ch == '\\' && i+1 < len(line) {
				i += 2
				continue
			}
			if ch == '\'' {
				s.inSingle = false
			}
			i++
			continue
		}
		if s.inDouble {
			if ch == '\\' && i+1 < len(line) {
				i += 2
				continue
			}
			if ch == '"' {
				s.inDouble = false
			}
			i++
			continue
		}
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}
		if ch == '#' {
			break
		}
		if ch == '\\' {
			if i+1 < len(line) {
				i += 2
			} else {
				i++
			}
			cmdPos = false
			continue
		}
		if ch == '\'' {
			s.inSingle = true
			cmdPos = false
			i++
			continue
		}
		if ch == '"' {
			s.inDouble = true
			cmdPos = false
			i++
			continue
		}
		if ch == ';' || ch == '|' || ch == '&' || ch == '(' {
			cmdPos = true
			i++
			continue
		}
		if ch == ')' || ch == '{' || ch == '}' {
			cmdPos = false
			i++
			continue
		}

		start := i
		for i < len(line) && !isWordBoundaryDelim(line[i]) {
			i++
		}
		if i == start {
			i++
			continue
		}
		word := strings.ToLower(line[start:i])
		if firstToken == "" {
			firstToken = word
		}
		if cmdPos {
			switch word {
			case "if", "begin", "for", "while", "function", "switch":
				opens++
				if word == "if" || word == "while" {
					cmdPos = true
				} else {
					cmdPos = false
				}
			case "else", "and", "or", "not", "!":
				cmdPos = true
			case "end":
				closes++
				cmdPos = false
			default:
				cmdPos = false
			}
		}
	}
	return opens, closes, firstToken
}

func scanFishIfBlock(lines []string, startIdx int) (endIdx int, ok bool) {
	var st fishScanState
	opens, closes, firstToken := st.scanLine(lines[startIdx])
	if firstToken != "if" || opens == 0 {
		return -1, false
	}
	depth := opens - closes
	if depth <= 0 && !st.inSingle && !st.inDouble {
		return startIdx, true
	}
	for j := startIdx + 1; j < len(lines); j++ {
		oj, cj, _ := st.scanLine(lines[j])
		depth += oj - cj
		if depth <= 0 && !st.inSingle && !st.inDouble {
			return j, true
		}
	}
	return -1, false
}

func isIdentByte(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
}

func extractExecutableCode(line string, allowBackslashEscape bool) string {
	var b strings.Builder
	inSingle := false
	inDouble := false
	escaped := false
	parenDepth := 0
	var doubleStack []int

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			if !inSingle && !inDouble {
				b.WriteByte(ch)
			}
			escaped = false
			continue
		}
		if inSingle {
			if ch == '\'' {
				if !allowBackslashEscape && i+1 < len(line) && line[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
				b.WriteByte(' ')
			}
			continue
		}
		if inDouble {
			if !allowBackslashEscape && ch == '`' {
				escaped = true
				continue
			}
			if allowBackslashEscape && ch == '\\' {
				escaped = true
				continue
			}
			if ch == '$' && i+1 < len(line) && line[i+1] == '(' {
				doubleStack = append(doubleStack, parenDepth)
				parenDepth++
				inDouble = false
				b.WriteByte(' ')
				i++
				continue
			}
			if ch == '"' {
				if !allowBackslashEscape && i+1 < len(line) && line[i+1] == '"' {
					i++
					continue
				}
				inDouble = false
				b.WriteByte(' ')
			}
			continue
		}
		if !allowBackslashEscape && ch == '`' {
			escaped = true
			continue
		}
		if allowBackslashEscape && ch == '\\' && i+1 < len(line) && (line[i+1] == '"' || line[i+1] == '\'' || line[i+1] == '\\' || line[i+1] == '$') {
			escaped = true
			continue
		}
		if ch == '\'' {
			inSingle = true
			b.WriteByte(' ')
			continue
		}
		if ch == '"' {
			inDouble = true
			b.WriteByte(' ')
			continue
		}
		if len(doubleStack) > 0 {
			if ch == '(' {
				parenDepth++
			} else if ch == ')' {
				parenDepth--
				if parenDepth == doubleStack[len(doubleStack)-1] {
					doubleStack = doubleStack[:len(doubleStack)-1]
					inDouble = true
					b.WriteByte(' ')
					continue
				}
			}
		}
		if ch == '#' && len(doubleStack) == 0 && (i == 0 || isShellDelimiterOrSpace(line[i-1])) {
			break
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func hasCompletionCommand(line string, allowBackslashEscape bool) bool {
	exec := strings.ToLower(extractExecutableCode(line, allowBackslashEscape))
	trimmed := strings.TrimSpace(exec)
	if trimmed == "" {
		return false
	}
	for _, prefix := range []string{"echo ", "printf ", "write-host ", "write-output "} {
		if strings.HasPrefix(trimmed, prefix) {
			return false
		}
	}

	const cliName = "minhthetus-cli"
	const compWord = "completion"
	pos := 0
	for {
		idx := strings.Index(exec[pos:], cliName)
		if idx == -1 {
			return false
		}
		abs := pos + idx
		if abs > 0 && isIdentByte(exec[abs-1]) {
			pos = abs + len(cliName)
			continue
		}
		rest := exec[abs+len(cliName):]
		if strings.HasPrefix(rest, ".exe") {
			rest = rest[4:]
		}
		if len(rest) == 0 || (rest[0] != ' ' && rest[0] != '\t') {
			pos = abs + len(cliName)
			continue
		}
		rest = strings.TrimLeft(rest, " \t")
		if strings.HasPrefix(rest, compWord) {
			after := rest[len(compWord):]
			if len(after) == 0 || !isIdentByte(after[0]) {
				return true
			}
		}
		pos = abs + len(cliName)
	}
}

func isCommandExistenceIfHeader(line string, allowBackslashEscape bool) bool {
	code := strings.ToLower(extractExecutableCode(line, allowBackslashEscape))
	if !strings.Contains(code, "minhthetus-cli") {
		return false
	}
	for _, cmdCheck := range []string{"get-command", "gcm ", "command -v", "type -q", "type ", "which ", "hash "} {
		if strings.Contains(code, cmdCheck) {
			return true
		}
	}
	return false
}

func isCompletionIfBlock(blockLines []string, hasPrecedingCLIComment bool, allowBackslashEscape bool) bool {
	for _, l := range blockLines {
		if hasCompletionCommand(l, allowBackslashEscape) {
			return true
		}
	}
	return hasPrecedingCLIComment && len(blockLines) > 0 && isCommandExistenceIfHeader(blockLines[0], allowBackslashEscape)
}

func isCLICommentLine(line string) bool {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "#") {
		return false
	}
	lower := strings.ToLower(t)
	if !strings.Contains(lower, "minhthetus-cli") {
		return false
	}
	return strings.Contains(lower, "completion") || strings.Contains(lower, "autocomplete")
}

func isPSVarAssignmentLine(trimmed string) bool {
	if len(trimmed) < 3 || trimmed[0] != '$' {
		return false
	}
	i := 1
	if !((trimmed[i] >= 'a' && trimmed[i] <= 'z') || (trimmed[i] >= 'A' && trimmed[i] <= 'Z') || trimmed[i] == '_') {
		return false
	}
	for i < len(trimmed) && ((trimmed[i] >= 'a' && trimmed[i] <= 'z') || (trimmed[i] >= 'A' && trimmed[i] <= 'Z') || (trimmed[i] >= '0' && trimmed[i] <= '9') || trimmed[i] == '_') {
		i++
	}
	rest := strings.TrimLeft(trimmed[i:], " \t")
	return strings.HasPrefix(rest, "=") && !strings.HasPrefix(rest, "==")
}

func isPSIfBlockLine(lines []string, idx int) bool {
	trimmed := strings.TrimSpace(lines[idx])
	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(lower, "if ") && !strings.HasPrefix(lower, "if(") {
		return false
	}
	if !strings.HasPrefix(strings.TrimLeft(lower[2:], " \t"), "(") {
		return false
	}
	if isBashIfThenLine(lines, idx) {
		return false
	}
	var st psScanState
	o, _, _ := st.scanLine(lines[idx])
	if o > 0 {
		return true
	}
	if !st.inMultiline() {
		for k := idx + 1; k < len(lines); k++ {
			if isPSLineBlankOrComment(lines[k], &st) {
				continue
			}
			return strings.HasPrefix(strings.TrimSpace(lines[k]), "{")
		}
	}
	return false
}

func isPowerShellLikeContent(lines []string) bool {
	for i, l := range lines {
		t := strings.TrimSpace(l)
		lower := strings.ToLower(t)
		if strings.Contains(lower, "get-command") || strings.Contains(lower, "write-host") ||
			strings.Contains(lower, "write-output") || strings.Contains(lower, "invoke-expression") ||
			strings.Contains(lower, "out-string") || strings.Contains(lower, "out-null") ||
			strings.Contains(lower, "import-module") || strings.Contains(lower, "set-location") ||
			strings.Contains(lower, "test-path") || strings.Contains(lower, "join-path") ||
			strings.Contains(lower, "split-path") || strings.Contains(lower, "new-object") ||
			strings.Contains(lower, "completion powershell") || strings.Contains(lower, "-erroraction") ||
			strings.Contains(lower, "$env:") || strings.Contains(lower, "$profile") ||
			strings.Contains(lower, "$psversiontable") || strings.Contains(lower, "$psscriptroot") ||
			strings.Contains(lower, "set-alias") ||
			strings.HasPrefix(t, "<#") || strings.HasSuffix(strings.TrimRight(l, " \t"), "@\"") ||
			strings.HasSuffix(strings.TrimRight(l, " \t"), "@'") ||
			isPSVarAssignmentLine(t) || isPSIfBlockLine(lines, i) {
			return true
		}
	}
	return false
}

// stripCompletionHooks removes marker-delimited, legacy multi-line/single-line (PowerShell/Bash/Zsh/Fish),
// and standalone minhthetus-cli completion hooks without leaving dangling braces or truncating user code.
func stripCompletionHooks(content string) (string, bool) {
	return stripCompletionHooksWithMode(content, nil)
}

func stripCompletionHooksForPath(content string, path string) (string, bool) {
	if path != "" {
		lower := strings.ToLower(path)
		if strings.HasSuffix(lower, ".ps1") {
			ps := true
			return stripCompletionHooksWithMode(content, &ps)
		}
		base := strings.ToLower(filepath.Base(path))
		if base == ".bashrc" || base == ".bash_profile" || base == ".bash_login" || base == ".profile" ||
			base == ".zshrc" || strings.HasSuffix(lower, ".fish") || strings.HasSuffix(lower, ".sh") || strings.HasSuffix(lower, ".zsh") {
			ps := false
			return stripCompletionHooksWithMode(content, &ps)
		}
	}
	return stripCompletionHooksWithMode(content, nil)
}

func stripCompletionHooksWithMode(content string, forcePS *bool) (string, bool) {
	if content == "" {
		return content, false
	}

	hasBOM := strings.HasPrefix(content, "\xef\xbb\xbf")
	body := strings.TrimPrefix(content, "\xef\xbb\xbf")
	if body == "" {
		return content, false
	}

	eol := "\n"
	if strings.Contains(body, "\r\n") {
		eol = "\r\n"
	}
	hadTrailingNewline := strings.HasSuffix(body, "\n")

	norm := strings.ReplaceAll(body, "\r\n", "\n")
	if hadTrailingNewline {
		norm = strings.TrimSuffix(norm, "\n")
	}
	lines := strings.Split(norm, "\n")

	// Pass 1: Remove marker-delimited blocks (# >>> minhthetus-cli completion >>> ... # <<< minhthetus-cli completion <<<)
	var pass1 []string
	changed := false
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == completionBlockStart {
			endIdx := -1
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) == completionBlockEnd {
					endIdx = j
					break
				}
			}
			if endIdx != -1 {
				changed = true
				if len(pass1) > 0 && strings.TrimSpace(pass1[len(pass1)-1]) == "" &&
					endIdx+1 < len(lines) && strings.TrimSpace(lines[endIdx+1]) == "" {
					pass1 = pass1[:len(pass1)-1]
				}
				i = endIdx
				continue
			}
		}
		pass1 = append(pass1, lines[i])
	}

	// Pass 2: Remove unmarked completion blocks (Bash/Zsh before PowerShell braces, then Fish) and standalone hooks
	psMode := isPowerShellLikeContent(pass1)
	if forcePS != nil {
		psMode = *forcePS
	}
	allowBackslashEscape := !psMode

	var psUserState psScanState
	var shUserState shellScanState
	userBraceDepth := 0
	userBashIfDepth := 0

	trackKeptLine := func(idx int, line string) {
		if psMode {
			o, c, _ := psUserState.scanLine(line)
			userBraceDepth += o - c
			if userBraceDepth < 0 {
				userBraceDepth = 0
			}
			return
		}
		// Even in non-PS files, check PS-specific here-string / block-comment only if not inside POSIX quotes
		var linePS psScanState
		linePS.inHereDouble = psUserState.inHereDouble
		linePS.inHereSingle = psUserState.inHereSingle
		linePS.inBlockComment = psUserState.inBlockComment
		if !shUserState.inSingle && !shUserState.inDouble && (linePS.inMultiline() || strings.HasPrefix(strings.TrimSpace(line), "<#") || strings.HasSuffix(strings.TrimRight(line, " \t"), "@\"") || strings.HasSuffix(strings.TrimRight(line, " \t"), "@'")) {
			o, c, _ := psUserState.scanLine(line)
			userBraceDepth += o - c
			if userBraceDepth < 0 {
				userBraceDepth = 0
			}
			return
		}

		// Track POSIX quotes, braces, and if/fi depth
		wasQuoted := shUserState.inSingle || shUserState.inDouble
		ifc, fic, hasThen, firstTok := shUserState.scanLine(line)
		if userBashIfDepth > 0 {
			userBashIfDepth += ifc - fic
		} else if firstTok == "if" && ifc > fic && (hasThen || isBashIfThenLine(pass1, idx)) {
			userBashIfDepth += ifc - fic
		}
		if userBashIfDepth < 0 {
			userBashIfDepth = 0
		}

		if !wasQuoted && !shUserState.inSingle && !shUserState.inDouble {
			code := extractExecutableCode(line, true)
			for k := 0; k < len(code); k++ {
				if code[k] == '{' {
					userBraceDepth++
				} else if code[k] == '}' && userBraceDepth > 0 {
					userBraceDepth--
				}
			}
		}
	}

	var pass2 []string
	for i := 0; i < len(pass1); i++ {
		line := pass1[i]
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if !psUserState.inMultiline() && !shUserState.inSingle && !shUserState.inDouble {
			isIfHeader := strings.HasPrefix(lower, "if ") || strings.HasPrefix(lower, "if(")
			headerHasCLI := strings.Contains(strings.ToLower(extractExecutableCode(line, allowBackslashEscape)), "minhthetus-cli")
			if isIfHeader && headerHasCLI {
				hasPrecedingComment := len(pass2) > 0 && isCLICommentLine(pass2[len(pass2)-1])
				endIdx := -1
				handled := false
				blockAllowBackslash := allowBackslashEscape

				// 1. Check Bash/Zsh if ... then ... fi FIRST so ${VAR} or { ...; } never enters PowerShell braces
				if !psMode && isBashIfThenLine(pass1, i) {
					handled = true
					blockAllowBackslash = true
					if e, ok := scanBashIfBlock(pass1, i); ok {
						endIdx = e
					}
				} else if psMode && isBashIfThenLine(pass1, i) {
					// Defensive fallback if forcePS wasn't set and content had mixed hints
					handled = true
					blockAllowBackslash = true
					if e, ok := scanBashIfBlock(pass1, i); ok {
						endIdx = e
					}
				} else if strings.HasPrefix(strings.TrimLeft(lower[2:], " \t"), "(") {
					// 2. PowerShell if (...) { ... } (requires '(' after 'if' and '{' on same or Allman line)
					var probePS psScanState
					o, c, _ := probePS.scanLine(line)
					isPSCandidate := o > 0
					if !isPSCandidate && c == 0 {
						for k := i + 1; k < len(pass1); k++ {
							if isPSLineBlankOrComment(pass1[k], &probePS) {
								continue
							}
							isPSCandidate = strings.HasPrefix(strings.TrimSpace(pass1[k]), "{")
							break
						}
					}
					if isPSCandidate {
						handled = true
						blockAllowBackslash = false
						if e, ok := scanPSIfBlock(pass1, i); ok {
							endIdx = e
						}
					}
				}

				// 3. Fish if ... end (only if not handled by Bash/Zsh or PowerShell)
				if !handled {
					if e, ok := scanFishIfBlock(pass1, i); ok {
						endIdx = e
					}
				}

				if endIdx != -1 && isCompletionIfBlock(pass1[i:endIdx+1], hasPrecedingComment, blockAllowBackslash) {
					if hasPrecedingComment {
						pass2 = pass2[:len(pass2)-1]
					}
					if len(pass2) > 0 && strings.TrimSpace(pass2[len(pass2)-1]) == "" &&
						endIdx+1 < len(pass1) && strings.TrimSpace(pass1[endIdx+1]) == "" {
						pass2 = pass2[:len(pass2)-1]
					}
					changed = true
					i = endIdx
					continue
				}
			}

			// Standalone CLI completion comment or invocation (only outside user blocks and when brace-balanced)
			if userBraceDepth == 0 && userBashIfDepth == 0 {
				balanced := false
				if psMode {
					var linePS psScanState
					o, c, _ := linePS.scanLine(line)
					balanced = o == c && !linePS.inMultiline()
				} else {
					var lineSH shellScanState
					ifc, fic, _, _ := lineSH.scanLine(line)
					code := extractExecutableCode(line, true)
					balanced = !lineSH.inSingle && !lineSH.inDouble && ifc == fic && strings.Count(code, "{") == strings.Count(code, "}")
				}
				if isCLICommentLine(line) || (balanced && hasCompletionCommand(line, allowBackslashEscape)) {
					if len(pass2) > 0 && strings.TrimSpace(pass2[len(pass2)-1]) == "" &&
						i+1 < len(pass1) && strings.TrimSpace(pass1[i+1]) == "" {
						pass2 = pass2[:len(pass2)-1]
					}
					changed = true
					continue
				}
			}
		}

		trackKeptLine(i, line)
		pass2 = append(pass2, line)
	}

	if !changed {
		return content, false
	}

	for len(pass2) > 0 && strings.TrimSpace(pass2[0]) == "" && len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		pass2 = pass2[1:]
	}
	for len(pass2) > 0 && strings.TrimSpace(pass2[len(pass2)-1]) == "" {
		pass2 = pass2[:len(pass2)-1]
	}

	if len(pass2) == 0 {
		return "", true
	}

	result := strings.Join(pass2, eol)
	if hadTrailingNewline {
		result += eol
	}
	if hasBOM {
		result = "\xef\xbb\xbf" + result
	}
	return result, true
}

func collectCleanupTargets(env completionEnv, home string) []cleanupTarget {
	caseInsensitive := env.goos == "windows" || env.goos == "darwin"
	seen := make(map[string]bool)
	var targets []cleanupTarget

	add := func(p string, deleteFile bool) {
		if strings.TrimSpace(p) == "" {
			return
		}
		cleaned := filepath.Clean(p)
		key := cleaned
		if caseInsensitive {
			key = strings.ToLower(cleaned)
		}
		if seen[key] {
			return
		}
		seen[key] = true
		targets = append(targets, cleanupTarget{Path: cleaned, DeleteFile: deleteFile})
	}

	add(filepath.Join(home, ".zshrc"), false)
	add(filepath.Join(home, ".bashrc"), false)
	add(filepath.Join(home, ".bash_profile"), false)

	xdgConfig := xdgConfigOrDefault(home, env.getenv)
	add(filepath.Join(xdgConfig, "fish", "config.fish"), false)
	add(filepath.Join(xdgConfig, "fish", "completions", "minhthetus-cli.fish"), true)

	// PowerShell 7+ & 5.1 profile locations
	add(filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), false)
	add(filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"), false)

	if env.winDocsDir != nil {
		if docsDir := strings.TrimSpace(env.winDocsDir(home)); docsDir != "" {
			add(filepath.Join(docsDir, "PowerShell", "Microsoft.PowerShell_profile.ps1"), false)
			add(filepath.Join(docsDir, "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"), false)
		}
	}

	if env.getenv != nil {
		for _, envName := range []string{"OneDrive", "OneDriveConsumer", "OneDriveCommercial"} {
			if od := strings.TrimSpace(env.getenv(envName)); od != "" {
				add(filepath.Join(od, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), false)
				add(filepath.Join(od, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"), false)
			}
		}
		if envProfile := strings.TrimSpace(env.getenv("PROFILE")); envProfile != "" &&
			strings.HasSuffix(strings.ToLower(envProfile), ".ps1") {
			add(envProfile, false)
		}
	}

	// Unix PowerShell 7+ location
	add(filepath.Join(xdgConfig, "powershell", "Microsoft.PowerShell_profile.ps1"), false)

	return targets
}

// cleanupShellConfigWithEnv removes autocomplete integrations across all supported shell profiles.
func cleanupShellConfigWithEnv(env completionEnv) []string {
	out := env.out
	if out == nil {
		out = os.Stdout
	}

	home, err := env.userHomeDir()
	if err != nil {
		fmt.Fprintf(out, "   %s\n", ui.WarningMessage(fmt.Sprintf("Could not resolve home directory: %v", err)))
		return nil
	}

	var cleanedPaths []string
	for _, target := range collectCleanupTargets(env, home) {
		info, err := os.Stat(target.Path)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(out, "   %s\n", ui.WarningMessage(fmt.Sprintf("Could not inspect %s: %v", shortenHomePath(target.Path, home), err)))
			}
			continue
		}

		displayPath := shortenHomePath(target.Path, home)
		if target.DeleteFile {
			if err := os.Remove(target.Path); err != nil {
				fmt.Fprintf(out, "   %s\n", ui.WarningMessage(fmt.Sprintf("Failed to remove %s: %v", displayPath, err)))
				continue
			}
			cleanedPaths = append(cleanedPaths, target.Path)
			fmt.Fprintf(out, "   %s Removed %s\n", ui.GreenStyle().Render(ui.BulletIcon), ui.CyanStyle().Render(displayPath))
			continue
		}

		data, err := os.ReadFile(target.Path)
		if err != nil {
			fmt.Fprintf(out, "   %s\n", ui.WarningMessage(fmt.Sprintf("Failed to read %s: %v", displayPath, err)))
			continue
		}

		cleaned, changed := stripCompletionHooksForPath(string(data), target.Path)
		if !changed {
			continue
		}

		perm := info.Mode().Perm()
		if perm == 0 {
			perm = 0644
		}
		if err := os.WriteFile(target.Path, []byte(cleaned), perm); err != nil {
			fmt.Fprintf(out, "   %s\n", ui.WarningMessage(fmt.Sprintf("Failed to update %s: %v", displayPath, err)))
			continue
		}

		cleanedPaths = append(cleanedPaths, target.Path)
		fmt.Fprintf(out, "   %s Cleaned up shell completion references in %s\n", ui.GreenStyle().Render(ui.BulletIcon), ui.CyanStyle().Render(displayPath))
	}

	return cleanedPaths
}

func refreshHintForShell(env completionEnv) string {
	shell, _ := detectShell(env)
	switch shell {
	case ShellPowerShell:
		return "Restart your PowerShell terminal (or run '. $PROFILE')"
	case ShellBash:
		if env.goos == "windows" {
			return "source ~/.bash_profile (or restart Git Bash)"
		}
		return "source ~/.bashrc (or restart your terminal)"
	case ShellZsh:
		return "source ~/.zshrc (or restart your terminal)"
	case ShellFish:
		return "Restart your Fish shell (or run 'exec fish')"
	default:
		return "Restart your terminal session to apply all changes."
	}
}

func runUninstall(deps uninstallDeps) error {
	out := deps.out
	if out == nil {
		out = os.Stdout
	}

	fmt.Fprintf(out, "%s Preparing uninstallation...\n", ui.CyanStyle().Render(ui.HourglassIcon))

	// 1. Detect executable path
	exePath, err := deps.exePath()
	if err != nil {
		fmt.Fprintln(out, ui.ErrorMessage(fmt.Sprintf("Failed to detect executable path: %v", err)))
		return err
	}

	// 2. Ask user for confirmation
	confirmMsg := fmt.Sprintf("Are you sure you want to completely uninstall minhthetus-cli (running from '%s')?", exePath)
	confirmed, err := deps.confirm(confirmMsg)
	if err != nil || !confirmed {
		fmt.Fprintln(out, ui.ErrorMessage("Uninstallation cancelled."))
		return nil
	}

	// 3. Perform shell profile cleanup
	fmt.Fprintf(out, "%s Performing cleanup...\n", ui.CyanStyle().Render(ui.CleanIcon))
	fmt.Fprintf(out, "   %s Cleaning shell completion integrations...\n", ui.GrayStyle().Render(ui.BulletIcon))
	cleanupShellConfigWithEnv(deps.env)

	// 4. Determine installation method and proceed
	method := detectInstallMethod(exePath)
	fmt.Fprintf(out, "%s Detected %s installation.\n", ui.CyanStyle().Render(ui.HourglassIcon), ui.BoldStyle.Render(string(method)))

	switch method {
	case Homebrew:
		fmt.Fprintf(out, "%s Triggering 'brew uninstall'...\n", ui.CyanStyle().Render(ui.HourglassIcon))
		if err := deps.runBrew(); err != nil {
			fmt.Fprintln(out, ui.ErrorMessage(fmt.Sprintf("Homebrew uninstallation failed: %v", err)))
			fmt.Fprintln(out, ui.HintMessage("Please run the command manually:"))
			fmt.Fprintf(out, "   %s\n", ui.CyanStyle().Render("brew uninstall minhthetus-cli"))
			return err
		}

	case WinGet:
		fmt.Fprintf(out, "%s Triggering 'winget uninstall'...\n", ui.CyanStyle().Render(ui.HourglassIcon))
		if err := deps.runWinget(); err != nil {
			fmt.Fprintln(out, ui.ErrorMessage(fmt.Sprintf("WinGet uninstallation failed: %v", err)))
			fmt.Fprintln(out, ui.HintMessage("If the running binary was locked, close this command and run manually:"))
			fmt.Fprintf(out, "   %s\n", ui.CyanStyle().Render("winget uninstall MinhTuLeHoang.minhthetus-cli"))
			return err
		}

	case GoInstall, ManualBuild:
		fmt.Fprintf(out, "%s Deleting binary file at '%s'...\n", ui.CyanStyle().Render(ui.HourglassIcon), ui.GrayStyle().Render(exePath))
		scheduled, err := deps.removeBinary(exePath)
		if err != nil {
			fmt.Fprintln(out, ui.ErrorMessage("Permission denied or failed to delete binary."))
			fmt.Fprintln(out, ui.HintMessage("Please remove the binary manually:"))
			if deps.env.goos == "windows" {
				fmt.Fprintf(out, "   PowerShell: %s\n", ui.CyanStyle().Render(fmt.Sprintf(`Remove-Item -Force "%s"`, exePath)))
				fmt.Fprintf(out, "   CMD:        %s\n", ui.CyanStyle().Render(fmt.Sprintf(`del /f "%s"`, exePath)))
			} else {
				fmt.Fprintf(out, "   %s\n", ui.CyanStyle().Render(fmt.Sprintf(`sudo rm "%s"`, exePath)))
			}
			return err
		}
		if scheduled {
			fmt.Fprintln(out, ui.InfoMessage("Scheduled automatic binary removal as soon as this process exits."))
			fmt.Fprintf(out, "   %s\n", ui.GrayStyle().Render(fmt.Sprintf(`Fallback if still present: del /f "%s"`, exePath)))
		}
	}

	fmt.Fprintln(out, ui.SuccessMessage("Successfully uninstalled minhthetus-cli! Goodbye!"))
	fmt.Fprintln(out, "\n"+ui.HintMessage("To apply all changes and refresh your active shell:"))
	fmt.Fprintf(out, "   %s\n", ui.CyanStyle().Render(refreshHintForShell(deps.env)))
	return nil
}

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Completely remove the CLI and all integrations",
	Long: `Remove shell autocomplete integrations across Zsh, Bash/Git Bash,
PowerShell 5.1/7+, and Fish, and uninstall the minhthetus-cli binary.`,
	Example: `minhthetus-cli uninstall`,
	Args:    cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Uninstall",
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = runUninstall(defaultUninstallDeps())
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
