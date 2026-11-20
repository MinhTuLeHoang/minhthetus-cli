package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/ui"
)

const (
	DarwinSudoPath              = "/etc/pam.d/sudo"
	DarwinSudoLocalPath         = "/etc/pam.d/sudo_local"
	DarwinSudoLocalTemplatePath = "/etc/pam.d/sudo_local.template"
	DarwinPAMTidModule          = "pam_tid.so"
	DarwinPAMReattachModule     = "pam_reattach.so"
	DarwinPAMReattachArm64      = "/opt/homebrew/lib/pam/pam_reattach.so"
	DarwinPAMReattachAmd64      = "/usr/local/lib/pam/pam_reattach.so"

	LinuxSudoPath       = "/etc/pam.d/sudo"
	LinuxPamFprintdName = "pam_fprintd.so"
	LinuxPamConfigsDir  = "/usr/share/pam-configs/fprintd"
)

var linuxPamFprintdCandidates = []string{
	"/usr/lib/x86_64-linux-gnu/security/pam_fprintd.so",
	"/usr/lib/aarch64-linux-gnu/security/pam_fprintd.so",
	"/lib/x86_64-linux-gnu/security/pam_fprintd.so",
	"/lib/aarch64-linux-gnu/security/pam_fprintd.so",
	"/usr/lib64/security/pam_fprintd.so",
	"/lib64/security/pam_fprintd.so",
	"/usr/lib/security/pam_fprintd.so",
	"/lib/security/pam_fprintd.so",
	"/usr/lib/*/security/pam_fprintd.so",
	"/lib/*/security/pam_fprintd.so",
}

// PAMPlanResult holds the output of a pure PAM planning computation.
type PAMPlanResult struct {
	TargetPath     string
	AlreadyEnabled bool
	AlreadyAbsent  bool
	NewContent     string
	Diff           string
	Note           string
	Warnings       []string
}

// DarwinPAMEnableInput provides inputs for PlanDarwinPAMEnable.
type DarwinPAMEnableInput struct {
	TargetPath      string
	CurrentContent  string
	TemplateContent string
	HasTemplate     bool
	PAMReattachPath string
}

// LinuxPAMEnableInput provides inputs for PlanLinuxSudoPAMEnable.
type LinuxPAMEnableInput struct {
	SudoPath         string
	SudoContent      string
	IncludedContents map[string]string
	PamFprintdPath   string
}

// FprintdStatus represents parsed `fprintd-list` output.
type FprintdStatus struct {
	HasDevice       bool
	EnrolledCount   int
	EnrolledFingers []string
	Reason          string
}

type pamRule struct {
	pamType string
	control string
	module  string
	args    []string
}

// parsePAMRule parses a non-comment PAM line (supporting bracketed controls like [success=end default=ignore]).
func parsePAMRule(line string) (pamRule, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "@include") {
		return pamRule{}, false
	}

	fields := strings.Fields(trimmed)
	if len(fields) < 3 {
		return pamRule{}, false
	}

	pamType := strings.TrimPrefix(strings.ToLower(fields[0]), "-")
	idx := 1
	var control string

	if strings.HasPrefix(fields[idx], "[") {
		var parts []string
		closed := false
		for idx < len(fields) {
			parts = append(parts, fields[idx])
			if strings.HasSuffix(fields[idx], "]") {
				closed = true
				idx++
				break
			}
			idx++
		}
		if !closed || idx >= len(fields) {
			return pamRule{}, false
		}
		control = strings.Join(parts, " ")
	} else {
		control = strings.ToLower(fields[idx])
		idx++
	}

	if idx >= len(fields) {
		return pamRule{}, false
	}

	return pamRule{
		pamType: pamType,
		control: control,
		module:  fields[idx],
		args:    fields[idx+1:],
	}, true
}

func pamModuleMatches(moduleToken, targetName string) bool {
	base := filepath.Base(moduleToken)
	return base == targetName || base == targetName+".2"
}

// IsActivePAMModule reports whether moduleName is configured as an active (uncommented) auth rule.
func IsActivePAMModule(content, moduleName string) bool {
	return findActivePAMModuleLine(content, moduleName) >= 0
}

func findActivePAMModuleLine(content, moduleName string) int {
	lines := splitLinesPreserve(content)
	for i, line := range lines {
		rule, ok := parsePAMRule(line)
		if !ok {
			continue
		}
		if rule.pamType == "auth" && pamModuleMatches(rule.module, moduleName) {
			return i
		}
	}
	return -1
}

// ParsePAMIncludes extracts included PAM filenames from @include and include/substack directives.
func ParsePAMIncludes(content string) []string {
	var includes []string
	seen := map[string]bool{}

	for _, line := range splitLinesPreserve(content) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && fields[0] == "@include" {
			target := filepath.Base(fields[1])
			if !seen[target] {
				seen[target] = true
				includes = append(includes, target)
			}
			continue
		}
		rule, ok := parsePAMRule(line)
		if ok && (rule.control == "include" || rule.control == "substack") {
			target := filepath.Base(rule.module)
			if !seen[target] {
				seen[target] = true
				includes = append(includes, target)
			}
		}
	}
	return includes
}

// ParseMacOSMajorVersion extracts the major version number from `sw_vers -productVersion` output.
func ParseMacOSMajorVersion(version string) (int, bool) {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return 0, false
	}
	parts := strings.Split(trimmed, ".")
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return major, true
}

// ResolveDarwinPAMTarget selects /etc/pam.d/sudo_local on Sonoma (14+) when /etc/pam.d/sudo includes it,
// falling back to /etc/pam.d/sudo otherwise.
func ResolveDarwinPAMTarget(macosVersion, sudoContent string) (targetPath string, usesSudoLocal bool) {
	major, ok := ParseMacOSMajorVersion(macosVersion)
	if ok && major >= 14 {
		for _, inc := range ParsePAMIncludes(sudoContent) {
			if inc == "sudo_local" {
				return DarwinSudoLocalPath, true
			}
		}
	}
	return DarwinSudoPath, false
}

// PlanDarwinPAMEnable computes an idempotent, safe Touch ID PAM edit for macOS.
func PlanDarwinPAMEnable(in DarwinPAMEnableInput) (PAMPlanResult, error) {
	target := in.TargetPath
	if target == "" {
		target = DarwinSudoLocalPath
	}

	tidIdx := findActivePAMModuleLine(in.CurrentContent, DarwinPAMTidModule)
	reattachIdx := findActivePAMModuleLine(in.CurrentContent, DarwinPAMReattachModule)

	if tidIdx >= 0 && (in.PAMReattachPath == "" || (reattachIdx >= 0 && reattachIdx < tidIdx)) {
		return PAMPlanResult{
			TargetPath:     target,
			AlreadyEnabled: true,
			NewContent:     in.CurrentContent,
			Note:           fmt.Sprintf("Touch ID (%s) is already enabled in %s.", DarwinPAMTidModule, target),
		}, nil
	}

	reattachLine := ""
	if in.PAMReattachPath != "" {
		reattachLine = fmt.Sprintf("auth       optional       %s ignore_ssh", in.PAMReattachPath)
	}
	tidLine := "auth       sufficient     pam_tid.so"

	var baseContent string
	if strings.TrimSpace(in.CurrentContent) == "" && target == DarwinSudoLocalPath {
		if in.HasTemplate && strings.TrimSpace(in.TemplateContent) != "" {
			baseContent = in.TemplateContent
		} else {
			baseContent = "# sudo_local: local config file which survives system update and is included for sudo\n"
		}
	} else {
		baseContent = in.CurrentContent
	}

	lines := splitLinesPreserve(baseContent)
	var updated []string

	if tidIdx >= 0 {
		// pam_tid.so is active, but pam_reattach.so needs to be inserted before it.
		for i, line := range lines {
			if i == tidIdx && reattachLine != "" {
				updated = append(updated, reattachLine)
			}
			// Drop misplaced pam_reattach after tidIdx if any.
			if r, ok := parsePAMRule(line); ok && pamModuleMatches(r.module, DarwinPAMReattachModule) && i > tidIdx {
				continue
			}
			updated = append(updated, line)
		}
	} else if target == DarwinSudoLocalPath {
		inserted := false
		for _, line := range lines {
			if reattachLine != "" {
				if r, ok := parsePAMRule(line); ok && r.pamType == "auth" && pamModuleMatches(r.module, DarwinPAMReattachModule) {
					continue
				}
			}
			trimmed := strings.TrimSpace(line)
			isCommentedTidRule := false
			if strings.HasPrefix(trimmed, "#") {
				candidate := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
				if r, ok := parsePAMRule(candidate); ok && r.pamType == "auth" && pamModuleMatches(r.module, DarwinPAMTidModule) {
					isCommentedTidRule = true
				}
			}
			if !inserted && isCommentedTidRule {
				if reattachLine != "" {
					updated = append(updated, reattachLine)
				}
				updated = append(updated, tidLine)
				inserted = true
				continue
			}
			updated = append(updated, line)
		}
		if !inserted {
			if reattachLine != "" {
				updated = append(updated, reattachLine)
			}
			updated = append(updated, tidLine)
		}
	} else {
		// Pre-Sonoma /etc/pam.d/sudo: insert before the first active auth directive.
		firstAuthIdx := -1
		firstNonReattachAuthIdx := -1
		for i, line := range lines {
			if r, ok := parsePAMRule(line); ok && r.pamType == "auth" {
				if firstAuthIdx < 0 {
					firstAuthIdx = i
				}
				if !pamModuleMatches(r.module, DarwinPAMReattachModule) && firstNonReattachAuthIdx < 0 {
					firstNonReattachAuthIdx = i
				}
			}
		}
		insertIdx := firstAuthIdx
		if reattachLine == "" && firstNonReattachAuthIdx >= 0 {
			insertIdx = firstNonReattachAuthIdx
		}
		if insertIdx < 0 {
			return PAMPlanResult{}, fmt.Errorf("refusing to modify %s: no existing auth directive found to preserve fallback authentication", target)
		}
		for i, line := range lines {
			if i == insertIdx {
				if reattachLine != "" {
					updated = append(updated, reattachLine)
				}
				updated = append(updated, tidLine)
			}
			if reattachLine != "" {
				if r, ok := parsePAMRule(line); ok && r.pamType == "auth" && pamModuleMatches(r.module, DarwinPAMReattachModule) {
					continue
				}
			}
			updated = append(updated, line)
		}
	}

	newContent := joinLinesWithTrailingNewline(updated)
	if !IsActivePAMModule(newContent, DarwinPAMTidModule) {
		return PAMPlanResult{}, fmt.Errorf("validation failed: %s is not active in generated %s", DarwinPAMTidModule, target)
	}
	if reattachLine != "" {
		rIdx := findActivePAMModuleLine(newContent, DarwinPAMReattachModule)
		tIdx := findActivePAMModuleLine(newContent, DarwinPAMTidModule)
		if rIdx < 0 || rIdx > tIdx {
			return PAMPlanResult{}, fmt.Errorf("validation failed: %s must appear before %s", DarwinPAMReattachModule, DarwinPAMTidModule)
		}
	}
	if target == DarwinSudoPath && !hasAnyAuthAfter(newContent, DarwinPAMTidModule) {
		return PAMPlanResult{}, fmt.Errorf("validation failed: fallback auth directive missing after %s in %s", DarwinPAMTidModule, target)
	}

	label := filepath.Base(target)
	return PAMPlanResult{
		TargetPath: target,
		NewContent: newContent,
		Diff:       GenerateSimpleDiff(label, in.CurrentContent, newContent),
	}, nil
}

// PlanDarwinPAMDisable removes active pam_tid.so and pam_reattach.so rules while keeping fallback auth intact.
func PlanDarwinPAMDisable(targetPath, currentContent string) (PAMPlanResult, error) {
	if targetPath == "" {
		targetPath = DarwinSudoLocalPath
	}
	hasTid := IsActivePAMModule(currentContent, DarwinPAMTidModule)
	hasReattach := IsActivePAMModule(currentContent, DarwinPAMReattachModule)

	if !hasTid && !hasReattach {
		return PAMPlanResult{
			TargetPath:    targetPath,
			AlreadyAbsent: true,
			NewContent:    currentContent,
			Note:          fmt.Sprintf("Touch ID (%s) is already disabled in %s.", DarwinPAMTidModule, targetPath),
		}, nil
	}

	var updated []string
	for _, line := range splitLinesPreserve(currentContent) {
		if r, ok := parsePAMRule(line); ok && r.pamType == "auth" {
			if pamModuleMatches(r.module, DarwinPAMTidModule) || pamModuleMatches(r.module, DarwinPAMReattachModule) {
				continue
			}
		}
		updated = append(updated, line)
	}

	newContent := joinLinesWithTrailingNewline(updated)
	if targetPath == DarwinSudoPath && !hasAnyActiveAuth(newContent) {
		return PAMPlanResult{}, fmt.Errorf("refusing to write %s: disabling Touch ID would leave no active auth rule", targetPath)
	}

	return PAMPlanResult{
		TargetPath: targetPath,
		NewContent: newContent,
		Diff:       GenerateSimpleDiff(filepath.Base(targetPath), currentContent, newContent),
	}, nil
}

// FindLinuxPamFprintd locates pam_fprintd.so across common Linux security directories.
func FindLinuxPamFprintd(fs HostFS) string {
	for _, pattern := range linuxPamFprintdCandidates {
		if strings.Contains(pattern, "*") {
			matches, err := fs.Glob(pattern)
			if err == nil {
				sort.Strings(matches)
				for _, m := range matches {
					if info, err := fs.Stat(m); err == nil && !info.IsDir() {
						return m
					}
				}
			}
			continue
		}
		if info, err := fs.Stat(pattern); err == nil && !info.IsDir() {
			return pattern
		}
	}
	return ""
}

// DetectFprintdEnrollment parses `fprintd-list <user>` output into device and finger count status.
func DetectFprintdEnrollment(output string) FprintdStatus {
	trimmed := strings.TrimSpace(output)
	lower := strings.ToLower(trimmed)

	if trimmed == "" {
		return FprintdStatus{HasDevice: false, Reason: "fprintd-list returned no output"}
	}
	if strings.Contains(lower, "nosuchdevice") || strings.Contains(lower, "no devices available") {
		return FprintdStatus{HasDevice: false, Reason: "no supported fingerprint device detected by fprintd"}
	}
	if strings.Contains(lower, "has no fingers enrolled") {
		return FprintdStatus{HasDevice: true, EnrolledCount: 0, Reason: "fingerprint device found, but no fingers are enrolled"}
	}

	var fingers []string
	seenHeader := false
	for _, line := range strings.Split(trimmed, "\n") {
		t := strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(t), "fingerprints for user") {
			seenHeader = true
			continue
		}
		if isNumberedFprintdLine(t) {
			if idx := strings.Index(t, ":"); idx >= 0 && idx+1 < len(t) {
				fingers = append(fingers, strings.TrimSpace(t[idx+1:]))
			} else {
				fingers = append(fingers, t)
			}
		} else if seenHeader && strings.HasPrefix(t, "-") && !strings.HasPrefix(t, "---") {
			fingers = append(fingers, strings.TrimSpace(strings.TrimPrefix(t, "-")))
		}
	}
	if len(fingers) > 0 {
		return FprintdStatus{
			HasDevice:       true,
			EnrolledCount:   len(fingers),
			EnrolledFingers: fingers,
			Reason:          fmt.Sprintf("%d enrolled fingerprint(s) detected", len(fingers)),
		}
	}

	if strings.Contains(lower, "fingerprints for user") {
		return FprintdStatus{HasDevice: true, EnrolledCount: 0, Reason: "fingerprint device responded with 0 parsed fingers"}
	}

	return FprintdStatus{HasDevice: false, Reason: trimmed}
}

func isNumberedFprintdLine(t string) bool {
	if !strings.HasPrefix(t, "- #") {
		return false
	}
	rest := strings.TrimPrefix(t, "- #")
	colon := strings.Index(rest, ":")
	if colon <= 0 {
		return false
	}
	_, err := strconv.Atoi(rest[:colon])
	return err == nil
}

// findActivePAMModuleInIncludes traverses PAM includes recursively via BFS with a visited-set cycle guard.
func findActivePAMModuleInIncludes(rootContent string, includedContents map[string]string, moduleName string) (string, bool) {
	queue := ParsePAMIncludes(rootContent)
	visited := map[string]bool{}
	for len(queue) > 0 {
		inc := queue[0]
		queue = queue[1:]
		if visited[inc] {
			continue
		}
		visited[inc] = true
		incContent, ok := includedContents[inc]
		if !ok {
			continue
		}
		if IsActivePAMModule(incContent, moduleName) {
			return inc, true
		}
		for _, child := range ParsePAMIncludes(incContent) {
			if !visited[child] {
				queue = append(queue, child)
			}
		}
	}
	return "", false
}

// PlanLinuxSudoPAMEnable plans safe insertion of `auth sufficient pam_fprintd.so` into /etc/pam.d/sudo
// before fallback password authentication, refusing if pam_fprintd.so is missing on disk.
func PlanLinuxSudoPAMEnable(in LinuxPAMEnableInput) (PAMPlanResult, error) {
	target := in.SudoPath
	if target == "" {
		target = LinuxSudoPath
	}

	if strings.TrimSpace(in.PamFprintdPath) == "" {
		return PAMPlanResult{}, fmt.Errorf("pam_fprintd.so was not found on disk; refusing to modify %s (missing PAM modules cause PAM_MODULE_UNKNOWN and lock out sudo). Install via: sudo apt install fprintd libpam-fprintd (Debian/Ubuntu), sudo dnf install fprintd fprintd-pam (Fedora/RHEL), or sudo pacman -S fprintd (Arch)", target)
	}

	if IsActivePAMModule(in.SudoContent, LinuxPamFprintdName) {
		return PAMPlanResult{
			TargetPath:     target,
			AlreadyEnabled: true,
			NewContent:     in.SudoContent,
			Note:           fmt.Sprintf("Fingerprint authentication (%s) is already enabled in %s.", LinuxPamFprintdName, target),
		}, nil
	}

	// Prevent double fingerprint prompts if any direct or transitive included stack file already enables pam_fprintd.so.
	if inc, ok := findActivePAMModuleInIncludes(in.SudoContent, in.IncludedContents, LinuxPamFprintdName); ok {
		return PAMPlanResult{
			TargetPath:     target,
			AlreadyEnabled: true,
			NewContent:     in.SudoContent,
			Note:           fmt.Sprintf("Fingerprint authentication (%s) is already active via included PAM file %q; skipping duplicate entry in %s.", LinuxPamFprintdName, inc, target),
		}, nil
	}

	lines := splitLinesPreserve(in.SudoContent)
	anchorIdx := findLinuxFallbackAnchor(lines)
	if anchorIdx < 0 {
		return PAMPlanResult{}, fmt.Errorf("refusing to modify %s: could not locate fallback password/include anchor (common-auth, system-auth, password-auth, or pam_unix.so)", target)
	}

	directive := "auth       sufficient     pam_fprintd.so"
	var updated []string
	for i, line := range lines {
		if i == anchorIdx {
			updated = append(updated, directive)
		}
		updated = append(updated, line)
	}

	newContent := joinLinesWithTrailingNewline(updated)
	if !IsActivePAMModule(newContent, LinuxPamFprintdName) {
		return PAMPlanResult{}, fmt.Errorf("validation failed: %s missing in generated %s", LinuxPamFprintdName, target)
	}
	if findLinuxFallbackAnchor(splitLinesPreserve(newContent)) <= findActivePAMModuleLine(newContent, LinuxPamFprintdName) {
		return PAMPlanResult{}, fmt.Errorf("validation failed: fallback password auth must remain after %s in %s", LinuxPamFprintdName, target)
	}

	return PAMPlanResult{
		TargetPath: target,
		NewContent: newContent,
		Diff:       GenerateSimpleDiff(filepath.Base(target), in.SudoContent, newContent),
	}, nil
}

// PlanLinuxSudoPAMDisable removes pam_fprintd.so from /etc/pam.d/sudo while ensuring fallback auth remains intact.
func PlanLinuxSudoPAMDisable(sudoPath, sudoContent string) (PAMPlanResult, error) {
	if sudoPath == "" {
		sudoPath = LinuxSudoPath
	}
	if !IsActivePAMModule(sudoContent, LinuxPamFprintdName) {
		return PAMPlanResult{
			TargetPath:    sudoPath,
			AlreadyAbsent: true,
			NewContent:    sudoContent,
			Note:          fmt.Sprintf("%s is not directly enabled in %s.", LinuxPamFprintdName, sudoPath),
		}, nil
	}

	var updated []string
	for _, line := range splitLinesPreserve(sudoContent) {
		if r, ok := parsePAMRule(line); ok && r.pamType == "auth" && pamModuleMatches(r.module, LinuxPamFprintdName) {
			continue
		}
		updated = append(updated, line)
	}

	newContent := joinLinesWithTrailingNewline(updated)
	if findLinuxFallbackAnchor(splitLinesPreserve(newContent)) < 0 {
		return PAMPlanResult{}, fmt.Errorf("refusing to modify %s: removing %s would leave no fallback auth anchor", sudoPath, LinuxPamFprintdName)
	}

	return PAMPlanResult{
		TargetPath: sudoPath,
		NewContent: newContent,
		Diff:       GenerateSimpleDiff(filepath.Base(sudoPath), sudoContent, newContent),
	}, nil
}

func findLinuxFallbackAnchor(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && fields[0] == "@include" {
			target := filepath.Base(fields[1])
			if target == "common-auth" || target == "system-auth" || target == "password-auth" {
				return i
			}
		}
		rule, ok := parsePAMRule(line)
		if !ok || rule.pamType != "auth" {
			continue
		}
		if rule.control == "include" || rule.control == "substack" {
			target := filepath.Base(rule.module)
			if target == "common-auth" || target == "system-auth" || target == "password-auth" {
				return i
			}
		}
		if pamModuleMatches(rule.module, "pam_unix.so") {
			return i
		}
	}
	return -1
}

// ParseAuthselectCurrent parses `authselect current` output for profile and fingerprint feature state.
func ParseAuthselectCurrent(output string) (managed bool, profile string, fingerprintEnabled bool) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" || strings.Contains(trimmed, "No existing configuration detected") {
		return false, "", false
	}
	for _, line := range strings.Split(trimmed, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "Profile ID:") {
			profile = strings.TrimSpace(strings.TrimPrefix(t, "Profile ID:"))
			managed = profile != ""
		}
		if strings.HasPrefix(t, "Enabled features:") && strings.Contains(t, "with-fingerprint") {
			fingerprintEnabled = true
		}
		if t == "- with-fingerprint" || t == "with-fingerprint" {
			fingerprintEnabled = true
		}
	}
	return managed, profile, fingerprintEnabled
}

// ParseOSReleasePrettyName extracts PRETTY_NAME or ID from /etc/os-release content.
func ParseOSReleasePrettyName(content string) string {
	var id string
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(t, "PRETTY_NAME="), `"'`)
		}
		if strings.HasPrefix(t, "ID=") {
			id = strings.Trim(strings.TrimPrefix(t, "ID="), `"'`)
		}
	}
	return id
}

// GetPAMBackupDir returns the external backup directory outside /etc/pam.d/.
func GetPAMBackupDir(homeDir string) string {
	return filepath.Join(homeDir, ".minhthetus-cli", "backups", "pam")
}

// SavePAMBackup persists a pre-edit backup outside /etc/pam.d/ with 0600 permissions.
func SavePAMBackup(backupDir, targetPath string, content []byte, now time.Time) (string, error) {
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create PAM backup directory %s: %w", backupDir, err)
	}
	base := filepath.Base(targetPath)
	if base == "" || base == "." {
		base = "pam"
	}
	filename := fmt.Sprintf("%s.%s.bak", base, now.UTC().Format("20060102-150405"))
	backupPath := filepath.Join(backupDir, filename)
	if err := os.WriteFile(backupPath, content, 0600); err != nil {
		return "", fmt.Errorf("failed to write PAM backup %s: %w", backupPath, err)
	}
	return backupPath, nil
}

// FindLatestPAMBackup returns the newest backup path for targetPath inside backupDir, or "".
func FindLatestPAMBackup(backupDir, targetPath string) string {
	base := filepath.Base(targetPath)
	pattern := filepath.Join(backupDir, base+".*.bak")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}
	sort.Strings(matches)
	return matches[len(matches)-1]
}

// GenerateSimpleDiff creates a readable unified-style diff between oldContent and newContent.
func GenerateSimpleDiff(label, oldContent, newContent string) string {
	if oldContent == newContent {
		return ""
	}
	oldLines := splitLinesPreserve(oldContent)
	newLines := splitLinesPreserve(newContent)

	// Simple LCS-based line diff so small PAM files render cleanly.
	m, n := len(oldLines), len(newLines)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- a/%s\n", label))
	sb.WriteString(fmt.Sprintf("+++ b/%s\n", label))

	i, j := 0, 0
	for i < m || j < n {
		if i < m && j < n && oldLines[i] == newLines[j] {
			sb.WriteString(" " + oldLines[i] + "\n")
			i++
			j++
		} else if j < n && (i == m || dp[i][j+1] >= dp[i+1][j]) {
			sb.WriteString("+" + newLines[j] + "\n")
			j++
		} else {
			sb.WriteString("-" + oldLines[i] + "\n")
			i++
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// ColorizeDiff applies internal/ui lipgloss colors to diff lines.
func ColorizeDiff(diff string) string {
	if diff == "" {
		return ""
	}
	var out []string
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			out = append(out, ui.CyanStyle().Bold(true).Render(line))
		case strings.HasPrefix(line, "+"):
			out = append(out, ui.GreenStyle().Render(line))
		case strings.HasPrefix(line, "-"):
			out = append(out, ui.RedStyle().Render(line))
		default:
			out = append(out, ui.GrayStyle().Render(line))
		}
	}
	return strings.Join(out, "\n")
}

func hasAnyActiveAuth(content string) bool {
	for _, line := range splitLinesPreserve(content) {
		if r, ok := parsePAMRule(line); ok && r.pamType == "auth" {
			return true
		}
	}
	return false
}

func hasAnyAuthAfter(content, moduleName string) bool {
	lines := splitLinesPreserve(content)
	idx := findActivePAMModuleLine(content, moduleName)
	if idx < 0 {
		return false
	}
	for i := idx + 1; i < len(lines); i++ {
		if r, ok := parsePAMRule(lines[i]); ok && r.pamType == "auth" {
			return true
		}
	}
	return false
}

func splitLinesPreserve(s string) []string {
	if s == "" {
		return nil
	}
	normalized := strings.ReplaceAll(s, "\r\n", "\n")
	normalized = strings.TrimSuffix(normalized, "\n")
	if normalized == "" {
		return nil
	}
	return strings.Split(normalized, "\n")
}

func joinLinesWithTrailingNewline(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
