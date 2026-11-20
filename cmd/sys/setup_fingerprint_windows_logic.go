package sys

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	WinBioTypeFingerprint = 0x00000008
	WinBioTypeFacial      = 0x00000002
	WinBioTypeVoice       = 0x00000004
	WinBioTypeAny         = 0xFFFFFFFF

	Windows11MinBuild        = 22000
	Windows1124H2SudoBuild   = 26100
	WindowsBackupStateFile   = "windows-fingerprint-state.json"
	WindowsSignInSettingsURI = "ms-settings:signinoptions"
)

// WinBioUnitInfo represents a biometric unit returned by winbio.dll WinBioEnumBiometricUnits.
type WinBioUnitInfo struct {
	UnitID           uint32
	PoolType         uint32
	BiometricFactor  uint32
	DeviceInstanceID string
	Description      string
	Manufacturer     string
	Model            string
}

// WindowsProbeSnapshot captures all pure diagnostic values collected on a Windows host.
type WindowsProbeSnapshot struct {
	CurrentBuild               int
	DisplayVersion             string
	UBR                        int
	ProductName                string
	IsAdmin                    bool
	WinBioFingerprintUnits     []WinBioUnitInfo
	WinBioAnyUnits             []WinBioUnitInfo
	PnPBiometricDevices        []string
	WbioSrvcState              string
	WbioSrvcStartType          string
	BiometricsPolicyBlocked    bool
	AccountEnrolledFactors     uint32
	HasAccountInfoKey          bool
	UserConsentAvailability  string
	EnableLUA                  uint32
	HasEnableLUA               bool
	ConsentPromptBehaviorAdmin uint32
	HasConsentPromptBehavior   bool
	PromptOnSecureDesktop      uint32
	HasPromptOnSecureDesktop   bool
	HasNativeSudoExe           bool
	NativeSudoMode             string
	HasGsudo                   bool
	BackupStatePath            string
}

// WindowsSavedState stores pre-modification Windows configuration for clean rollback.
type WindowsSavedState struct {
	ModifiedUACToggle      bool   `json:"modified_uac_toggle,omitempty"`
	PrevConsentPromptAdmin uint32 `json:"prev_consent_prompt_admin,omitempty"`
	ModifiedWbioSrvc       bool   `json:"modified_wbio_srvc,omitempty"`
	PrevWbioSrvcStartType  string `json:"prev_wbio_srvc_start_type,omitempty"`
	ModifiedNativeSudo     bool   `json:"modified_native_sudo,omitempty"`
	PrevNativeSudoMode     string `json:"prev_native_sudo_mode,omitempty"`
}

// WindowsActionStep describes a concrete Windows remediation/setup action.
type WindowsActionStep struct {
	Key         string
	Description string
	Command     string
	RequiresUAC bool
}

// WindowsConfigPlan captures planned Windows setup or revert steps.
type WindowsConfigPlan struct {
	Title            string
	Description      string
	Steps            []WindowsActionStep
	Warnings         []string
	AlreadySatisfied bool
	OpenSettingsURI  string
}

// ParseWindowsBuildNumber safely parses the Windows CurrentBuild string into an integer.
func ParseWindowsBuildNumber(raw string) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// IsWindows11 reports whether the Windows build number belongs to Windows 11 (>= 22000).
// Never rely on registry ProductName, which frequently still reads "Windows 10" on Windows 11.
func IsWindows11(build int) bool {
	return build >= Windows11MinBuild
}

// SupportsWindowsNativeSudo reports whether the build supports native Windows 11 24H2+ sudo.exe (>= 26100).
func SupportsWindowsNativeSudo(build int) bool {
	return build >= Windows1124H2SudoBuild
}

// FormatWindowsOSLabel formats an accurate OS version string from CurrentBuild and DisplayVersion.
func FormatWindowsOSLabel(snap WindowsProbeSnapshot) string {
	family := "Windows"
	if snap.CurrentBuild >= Windows11MinBuild {
		family = "Windows 11"
	} else if snap.CurrentBuild > 0 {
		family = "Windows 10 / Legacy"
	}
	parts := []string{family}
	if snap.DisplayVersion != "" {
		parts = append(parts, snap.DisplayVersion)
	}
	if snap.CurrentBuild > 0 {
		if snap.UBR > 0 {
			parts = append(parts, fmt.Sprintf("Build %d.%d", snap.CurrentBuild, snap.UBR))
		} else {
			parts = append(parts, fmt.Sprintf("Build %d", snap.CurrentBuild))
		}
	}
	return strings.Join(parts, " ")
}

// BuildWindowsElevationPSCommand constructs a PowerShell command that runs cmdLine elevated
// via UAC and guarantees a non-zero exit code on UAC cancellation or Start-Process failure.
func BuildWindowsElevationPSCommand(cmdLine string) string {
	escaped := strings.ReplaceAll(cmdLine, "'", "''")
	return fmt.Sprintf(
		`$ErrorActionPreference = 'Stop'; try { $p = Start-Process cmd -ArgumentList @('/c', '%s') -Verb RunAs -Wait -PassThru; if (-not $p) { exit 1 }; exit $p.ExitCode } catch { Write-Error -ErrorAction Continue $_; exit 1 }`,
		escaped,
	)
}

// ParseWindowsServiceStartDWORD maps HKLM\SYSTEM\CurrentControlSet\Services\<svc>\Start DWORD values.
func ParseWindowsServiceStartDWORD(startVal uint32) string {
	switch startVal {
	case 2:
		return "Automatic"
	case 3:
		return "Manual"
	case 4:
		return "Disabled"
	default:
		return "Unknown"
	}
}

// ParseSCServiceOutput parses `sc query WbioSrvc` and `sc qc WbioSrvc` outputs,
// supporting both English keywords and numeric codes on localized Windows systems.
func ParseSCServiceOutput(queryOut, qcOut string) (state string, startType string) {
	state = parseSCQueryState(queryOut)
	startType = parseSCStartType(qcOut)
	return state, startType
}

func parseSCQueryState(queryOut string) string {
	lines := strings.Split(queryOut, "\n")
	// Pass 1: explicit STATE line or unambiguous state keyword.
	for _, line := range lines {
		upper := strings.ToUpper(strings.TrimSpace(line))
		if !strings.Contains(upper, ":") {
			continue
		}
		if strings.Contains(upper, "STATE") {
			if st := matchSCStateLine(upper); st != "Unknown" {
				return st
			}
		}
	}
	// Pass 2: localized STATE label — match unambiguous state keywords or numeric code on the line after TYPE.
	afterType := false
	for _, line := range lines {
		upper := strings.ToUpper(strings.TrimSpace(line))
		if !strings.Contains(upper, ":") {
			continue
		}
		parts := strings.SplitN(upper, ":", 2)
		val := strings.TrimSpace(parts[1])
		if strings.Contains(val, "WIN32") || strings.Contains(val, "KERNEL") || strings.Contains(val, "DRIVER") {
			afterType = true
			continue
		}
		if afterType {
			return matchSCStateLine(upper)
		}
		if st := matchSCStateKeywordOnly(upper); st != "Unknown" {
			return st
		}
	}
	return "Unknown"
}

func matchSCStateKeywordOnly(upperLine string) string {
	switch {
	case strings.Contains(upperLine, "START_PENDING"),
		strings.Contains(upperLine, "STOP_PENDING"),
		strings.Contains(upperLine, "CONTINUE_PENDING"),
		strings.Contains(upperLine, "PAUSE_PENDING"):
		return "Unknown"
	case strings.Contains(upperLine, "RUNNING"):
		return "Running"
	case strings.Contains(upperLine, "STOPPED"):
		return "Stopped"
	case strings.Contains(upperLine, "PAUSED"):
		return "Paused"
	default:
		return "Unknown"
	}
}

func matchSCStateLine(upperLine string) string {
	if st := matchSCStateKeywordOnly(upperLine); st != "Unknown" {
		return st
	}
	if code, ok := parseSCLeadingInt(upperLine); ok {
		switch code {
		case 1:
			return "Stopped"
		case 4:
			return "Running"
		case 7:
			return "Paused"
		}
	}
	return "Unknown"
}

func parseSCStartType(qcOut string) string {
	lines := strings.Split(qcOut, "\n")
	// Pass 1: explicit START_TYPE label or unambiguous start-type keyword.
	for _, line := range lines {
		upper := strings.ToUpper(strings.TrimSpace(line))
		if !strings.Contains(upper, ":") {
			continue
		}
		if strings.Contains(upper, "START_TYPE") {
			if st := matchSCStartTypeLine(upper); st != "Unknown" {
				return st
			}
		}
		if st := matchSCStartTypeKeywordOnly(upper); st != "Unknown" {
			return st
		}
	}
	// Pass 2: localized START_TYPE label — parse numeric code on the line immediately after TYPE.
	afterType := false
	for _, line := range lines {
		upper := strings.ToUpper(strings.TrimSpace(line))
		if !strings.Contains(upper, ":") {
			continue
		}
		parts := strings.SplitN(upper, ":", 2)
		val := strings.TrimSpace(parts[1])
		if strings.Contains(val, "WIN32") || strings.Contains(val, "KERNEL") || strings.Contains(val, "DRIVER") {
			afterType = true
			continue
		}
		if afterType {
			return matchSCStartTypeLine(upper)
		}
	}
	return "Unknown"
}

func matchSCStartTypeKeywordOnly(upperLine string) string {
	switch {
	case strings.Contains(upperLine, "AUTO_START"):
		return "Automatic"
	case strings.Contains(upperLine, "DEMAND_START"):
		return "Manual"
	case strings.Contains(upperLine, "DISABLED"):
		return "Disabled"
	default:
		return "Unknown"
	}
}

func matchSCStartTypeLine(upperLine string) string {
	if st := matchSCStartTypeKeywordOnly(upperLine); st != "Unknown" {
		return st
	}
	if code, ok := parseSCLeadingInt(upperLine); ok {
		switch code {
		case 2:
			return "Automatic"
		case 3:
			return "Manual"
		case 4:
			return "Disabled"
		}
	}
	return "Unknown"
}

func parseSCLeadingInt(line string) (int, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return 0, false
	}
	fields := strings.Fields(strings.TrimSpace(parts[1]))
	if len(fields) == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, false
	}
	return n, true
}

// ParsePnPBiometricOutput extracts device names/IDs from `pnputil /enum-devices /class Biometric` output.
func ParsePnPBiometricOutput(out string) []string {
	lower := strings.ToLower(out)
	if strings.TrimSpace(out) == "" || strings.Contains(lower, "no devices were found") {
		return nil
	}
	var devices []string
	var curDesc string
	for _, line := range strings.Split(out, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(t), "device description:") {
			curDesc = strings.TrimSpace(strings.SplitN(t, ":", 2)[1])
			if curDesc != "" {
				devices = append(devices, curDesc)
			}
		}
	}
	return devices
}

// ParseSudoConfigState interprets Windows 11 24H2+ Sudo registry Enabled value or `sudo config` CLI output.
func ParseSudoConfigState(regVal uint32, hasRegVal bool, cliOut string) string {
	if hasRegVal {
		switch regVal {
		case 0:
			return "disabled"
		case 1:
			return "forceNewWindow"
		case 2:
			return "disableInput"
		case 3:
			return "normal"
		}
	}

	foundModes := map[string]bool{}
	for _, line := range strings.Split(cliOut, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		// Skip CLI usage/help syntax lines such as `sudo config --enable <disabled|forceNewWindow|disableInput|normal>`.
		if strings.Contains(lower, "usage:") ||
			strings.Contains(lower, "--enable") ||
			strings.Contains(lower, "--disable") ||
			strings.Contains(lower, "<") ||
			strings.Contains(lower, ">") ||
			strings.Contains(lower, "|") {
			continue
		}
		tokens := strings.FieldsFunc(lower, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
		})
		for _, tok := range tokens {
			switch tok {
			case "normal":
				foundModes["normal"] = true
			case "disableinput":
				foundModes["disableInput"] = true
			case "forcenewwindow":
				foundModes["forceNewWindow"] = true
			case "disabled":
				foundModes["disabled"] = true
			}
		}
	}

	if len(foundModes) == 1 {
		for mode := range foundModes {
			return mode
		}
	}
	return "unknown"
}

// BuildWindowsStatus translates a WindowsProbeSnapshot into a user-facing FingerprintStatus.
func BuildWindowsStatus(snap WindowsProbeSnapshot) FingerprintStatus {
	hwDetected := len(snap.WinBioFingerprintUnits) > 0 || len(snap.PnPBiometricDevices) > 0
	enrolled := hwDetected && (snap.AccountEnrolledFactors&WinBioTypeFingerprint != 0)

	uacEnabled := !snap.HasEnableLUA || snap.EnableLUA != 0
	uacCredPrompt := uacEnabled && snap.HasConsentPromptBehavior && snap.ConsentPromptBehaviorAdmin == 1
	wbioRunning := snap.WbioSrvcState == "Running"
	sudoInlineReady := !snap.HasNativeSudoExe || snap.NativeSudoMode == "normal"

	// Enabled means Windows fingerprint hardware is present, WbioSrvc is active (or hardware ready),
	// and UAC/sudo readiness is configured to allow Windows Hello credential verification.
	enabled := hwDetected && enrolled && wbioRunning && uacCredPrompt && sudoInlineReady

	st := FingerprintStatus{
		OS:               "windows",
		OSVersion:        FormatWindowsOSLabel(snap),
		Supported:        true,
		Configurable:     true,
		Enabled:          enabled,
		BiometricReady:   hwDetected && enrolled && !snap.BiometricsPolicyBlocked,
		HardwareDetected: hwDetected,
		Enrolled:         enrolled,
		Method:           "Windows Hello + UAC (consent.exe)" + sudoMethodSuffix(snap),
		ConfigPath:       `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`,
		BackupPath:       snap.BackupStatePath,
	}

	st.Details = append(st.Details,
		fmt.Sprintf("Build & Sudo: %s | System32\\sudo.exe: %v (mode: %s) | gsudo: %v",
			FormatWindowsOSLabel(snap), snap.HasNativeSudoExe, nonEmptyOr(snap.NativeSudoMode, "not available"), snap.HasGsudo),
		fmt.Sprintf("WbioSrvc Service: State=%s, StartType=%s",
			nonEmptyOr(snap.WbioSrvcState, "Unknown"), nonEmptyOr(snap.WbioSrvcStartType, "Unknown")),
		fmt.Sprintf("WinBio Fingerprint Units (WINBIO_TYPE_FINGERPRINT=0x8): %d | PnP Biometric Devices: %d",
			len(snap.WinBioFingerprintUnits), len(snap.PnPBiometricDevices)),
		fmt.Sprintf("Windows Hello UserConsentVerifier: %s | WinBio AccountInfo EnrolledFactors=0x%x",
			nonEmptyOr(snap.UserConsentAvailability, "Unknown"), snap.AccountEnrolledFactors),
	)

	if snap.HasEnableLUA || snap.HasConsentPromptBehavior || snap.HasPromptOnSecureDesktop {
		st.Details = append(st.Details, fmt.Sprintf(
			"UAC Policies: EnableLUA=%d, ConsentPromptBehaviorAdmin=%d (%s), PromptOnSecureDesktop=%d",
			snap.EnableLUA, snap.ConsentPromptBehaviorAdmin, describeUACAdminMode(snap.ConsentPromptBehaviorAdmin), snap.PromptOnSecureDesktop,
		))
	}

	// Highlight virtual non-fingerprint units (e.g., Analog NUI Voice Virtual Sensor) so diagnostics are crystal clear.
	for _, u := range snap.WinBioAnyUnits {
		if u.BiometricFactor != WinBioTypeFingerprint {
			st.Details = append(st.Details, fmt.Sprintf(
				"Non-fingerprint WinBio unit ignored: %s (Factor=0x%x)",
				nonEmptyOr(u.Description, u.DeviceInstanceID), u.BiometricFactor,
			))
		}
	}

	if snap.HasEnableLUA && snap.EnableLUA == 0 {
		st.Warnings = append(st.Warnings, "UAC is disabled (EnableLUA=0); consent.exe credential prompts and Windows Hello UAC verification are inactive.")
	}
	if snap.BiometricsPolicyBlocked {
		st.Warnings = append(st.Warnings, "Windows Biometrics is restricted by Group Policy (HKLM\\SOFTWARE\\Policies\\Microsoft\\Biometrics).")
	}
	if !hwDetected {
		st.Warnings = append(st.Warnings, "No physical fingerprint reader (WINBIO_TYPE_FINGERPRINT) was detected on this Windows machine.")
		if strings.EqualFold(snap.UserConsentAvailability, "Available") {
			st.Notes = append(st.Notes, "UserConsentVerifier reports 'Available' due to Windows Hello PIN/NGC — this does NOT indicate a physical fingerprint sensor.")
		}
	} else if !enrolled {
		st.Warnings = append(st.Warnings, "Fingerprint hardware detected, but no fingerprint appears enrolled for the current user. Use Settings > Sign-in options (ms-settings:signinoptions) to enroll.")
	}

	if !snap.HasNativeSudoExe {
		if snap.CurrentBuild > 0 && snap.CurrentBuild < Windows1124H2SudoBuild {
			st.Notes = append(st.Notes, fmt.Sprintf("Native Windows sudo.exe requires Windows 11 24H2 (Build >= %d); current build is %d.", Windows1124H2SudoBuild, snap.CurrentBuild))
		}
	} else if snap.NativeSudoMode != "normal" {
		st.Notes = append(st.Notes, fmt.Sprintf("Native sudo.exe is installed (mode=%s); running `--enable` can configure `sudo config --enable normal` for inline execution.", snap.NativeSudoMode))
	}

	if snap.HasConsentPromptBehavior && snap.ConsentPromptBehaviorAdmin == 5 && hwDetected {
		st.Notes = append(st.Notes, "UAC ConsentPromptBehaviorAdmin is currently 5 (Yes/No consent). Setting it to 1 prompts for Windows Hello/credentials on Secure Desktop.")
	}

	st.Limitations = []string{
		"Windows has no PAM equivalent for inline terminal sudo authentication; both native sudo.exe (Win11 24H2+) and UAC always elevate via consent.exe on the Secure Desktop.",
		"A CLI cannot bypass consent.exe, intercept Secure Desktop, or force UAC Credential Provider ordering via any supported registry/GPO setting.",
		"Enrolled Windows Hello fingerprint/PIN works on UAC credential prompts with password fallback; passwords are NEVER captured or stored by minhthetus-cli.",
	}

	return st
}

// PlanWindowsEnableActions determines feasible, safe setup/repair actions on Windows without capturing passwords.
func PlanWindowsEnableActions(snap WindowsProbeSnapshot) WindowsConfigPlan {
	hwDetected := len(snap.WinBioFingerprintUnits) > 0 || len(snap.PnPBiometricDevices) > 0
	var steps []WindowsActionStep
	var warnings []string

	if snap.BiometricsPolicyBlocked {
		warnings = append(warnings, "Biometrics Group Policy appears disabled/blocked; administrator policy changes may be required.")
	}
	if hwDetected && snap.HasEnableLUA && snap.EnableLUA == 0 {
		warnings = append(warnings, "UAC is disabled (EnableLUA=0); ConsentPromptBehaviorAdmin=1 will not trigger Windows Hello prompts until EnableLUA=1 is enabled and Windows reboots.")
	}

	if hwDetected && (snap.WbioSrvcStartType == "Disabled" || snap.WbioSrvcState == "Stopped") {
		if snap.WbioSrvcStartType == "Disabled" {
			steps = append(steps, WindowsActionStep{
				Key:         "enable_wbiosrvc",
				Description: "Set Windows Biometric Service (WbioSrvc) startup type to Manual and start it",
				Command:     "sc.exe config WbioSrvc start= demand && sc.exe start WbioSrvc",
				RequiresUAC: true,
			})
		} else {
			steps = append(steps, WindowsActionStep{
				Key:         "start_wbiosrvc",
				Description: "Start Windows Biometric Service (WbioSrvc)",
				Command:     "sc.exe start WbioSrvc",
				RequiresUAC: true,
			})
		}
	}

	if hwDetected && snap.HasNativeSudoExe && snap.NativeSudoMode != "normal" {
		steps = append(steps, WindowsActionStep{
			Key:         "enable_native_sudo",
			Description: "Enable Windows 11 24H2+ native sudo in inline ('normal') mode",
			Command:     "sudo config --enable normal",
			RequiresUAC: true,
		})
	}

	// Only propose changing ConsentPromptBehaviorAdmin from 5 -> 1 when a fingerprint reader is actually present;
	// otherwise switching UAC to credential-prompt mode would force password typing on every UAC prompt!
	if hwDetected && (!snap.HasConsentPromptBehavior || snap.ConsentPromptBehaviorAdmin != 1) {
		steps = append(steps, WindowsActionStep{
			Key:         "set_uac_cred_prompt",
			Description: "Set UAC ConsentPromptBehaviorAdmin=1 so UAC prompts for Windows Hello fingerprint/credentials on Secure Desktop (reversible via --disable)",
			Command:     `reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v ConsentPromptBehaviorAdmin /t REG_DWORD /d 1 /f`,
			RequiresUAC: true,
		})
	} else if !hwDetected {
		warnings = append(warnings, "Skipping WbioSrvc, native sudo.exe, and UAC ConsentPromptBehaviorAdmin=1 changes because no physical fingerprint sensor was detected.")
	}

	openSettings := ""
	if hwDetected && (snap.AccountEnrolledFactors&WinBioTypeFingerprint == 0) {
		openSettings = WindowsSignInSettingsURI
		steps = append(steps, WindowsActionStep{
			Key:         "open_signin_settings",
			Description: "Open Windows Settings > Sign-in options to enroll your fingerprint with Windows Hello",
			Command:     "cmd /c start ms-settings:signinoptions",
			RequiresUAC: false,
		})
	}

	if len(steps) == 0 {
		msg := "All feasible Windows biometric/sudo settings are already configured."
		if !hwDetected {
			if snap.HasNativeSudoExe {
				msg = "No Windows changes are applicable: no physical fingerprint reader was detected, so WbioSrvc, native sudo.exe, and UAC credential prompt settings were left untouched."
			} else {
				msg = fmt.Sprintf("No additional Windows changes are applicable: no physical fingerprint reader detected, native sudo.exe requires Build >= %d (current: %d), and Windows UAC/Secure Desktop cannot be bypassed.", Windows1124H2SudoBuild, snap.CurrentBuild)
			}
		}
		return WindowsConfigPlan{
			Title:            "Windows Fingerprint & UAC Readiness",
			Description:      msg,
			Warnings:         warnings,
			AlreadySatisfied: true,
		}
	}

	return WindowsConfigPlan{
		Title:           "Configure Windows Biometric & Sudo Readiness",
		Description:     "Apply feasible Windows Biometric Service, Windows 11 sudo.exe, and UAC Windows Hello settings.",
		Steps:           steps,
		Warnings:        warnings,
		OpenSettingsURI: openSettings,
	}
}

// PlanWindowsDisableActions plans reverting changes previously applied by `sys setup finger-print`.
func PlanWindowsDisableActions(snap WindowsProbeSnapshot, saved *WindowsSavedState) WindowsConfigPlan {
	var steps []WindowsActionStep

	if saved != nil && saved.ModifiedUACToggle {
		prev := saved.PrevConsentPromptAdmin
		steps = append(steps, WindowsActionStep{
			Key:         "restore_uac_prompt",
			Description: fmt.Sprintf("Restore UAC ConsentPromptBehaviorAdmin back to %d (%s)", prev, describeUACAdminMode(prev)),
			Command:     fmt.Sprintf(`reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v ConsentPromptBehaviorAdmin /t REG_DWORD /d %d /f`, prev),
			RequiresUAC: true,
		})
	} else if snap.HasConsentPromptBehavior && snap.ConsentPromptBehaviorAdmin == 1 {
		steps = append(steps, WindowsActionStep{
			Key:         "restore_uac_default",
			Description: "Restore UAC ConsentPromptBehaviorAdmin back to Windows default 5 (consent prompt for non-Windows binaries)",
			Command:     `reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v ConsentPromptBehaviorAdmin /t REG_DWORD /d 5 /f`,
			RequiresUAC: true,
		})
	}

	if saved != nil && saved.ModifiedNativeSudo && snap.HasNativeSudoExe && saved.PrevNativeSudoMode != "" && saved.PrevNativeSudoMode != snap.NativeSudoMode {
		cmdStr := "sudo config --disable"
		targetDesc := "disabled"
		switch saved.PrevNativeSudoMode {
		case "forceNewWindow", "disableInput", "normal":
			cmdStr = "sudo config --enable " + saved.PrevNativeSudoMode
			targetDesc = saved.PrevNativeSudoMode
		case "disabled", "unknown":
			cmdStr = "sudo config --disable"
			targetDesc = "disabled"
		}
		if !(cmdStr == "sudo config --disable" && snap.NativeSudoMode == "disabled") {
			steps = append(steps, WindowsActionStep{
				Key:         "restore_native_sudo",
				Description: fmt.Sprintf("Restore Windows sudo.exe mode to %s", targetDesc),
				Command:     cmdStr,
				RequiresUAC: true,
			})
		}
	}

	if len(steps) == 0 {
		return WindowsConfigPlan{
			Title:            "Revert Windows UAC / Sudo Fingerprint Settings",
			Description:      "Windows UAC and biometric settings are already at default / non-modified state.",
			AlreadySatisfied: true,
		}
	}

	return WindowsConfigPlan{
		Title:       "Revert Windows UAC / Sudo Fingerprint Settings",
		Description: "Restore prior UAC / Windows sudo configuration.",
		Steps:       steps,
	}
}

// SaveWindowsStateBackup persists WindowsSavedState under ~/.minhthetus-cli/backups/pam/.
func SaveWindowsStateBackup(backupDir string, state WindowsSavedState) (string, error) {
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(backupDir, WindowsBackupStateFile)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return path, nil
}

// LoadWindowsStateBackup loads WindowsSavedState if present.
func LoadWindowsStateBackup(backupDir string) (*WindowsSavedState, string) {
	path := filepath.Join(backupDir, WindowsBackupStateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ""
	}
	var state WindowsSavedState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, ""
	}
	return &state, path
}

func sudoMethodSuffix(snap WindowsProbeSnapshot) string {
	if snap.HasNativeSudoExe {
		return " + Win11 sudo.exe"
	}
	return ""
}

func describeUACAdminMode(val uint32) string {
	switch val {
	case 0:
		return "Elevate without prompting"
	case 1:
		return "Prompt for credentials on secure desktop"
	case 2:
		return "Prompt for consent on secure desktop"
	case 3:
		return "Prompt for credentials"
	case 4:
		return "Prompt for consent"
	case 5:
		return "Prompt for consent for non-Windows binaries"
	default:
		return "Custom"
	}
}

func nonEmptyOr(val, fallback string) string {
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}
