package sys

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// --- Test Fakes ---

type fakeFS struct {
	files map[string][]byte
	globs map[string][]string
	home  string
}

func (f *fakeFS) Stat(path string) (os.FileInfo, error) {
	if _, ok := f.files[path]; ok {
		return fakeFileInfo{name: filepath.Base(path)}, nil
	}
	return nil, os.ErrNotExist
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	if b, ok := f.files[path]; ok {
		return b, nil
	}
	return nil, os.ErrNotExist
}

func (f *fakeFS) Glob(pattern string) ([]string, error) {
	return f.globs[pattern], nil
}

func (f *fakeFS) UserHomeDir() (string, error) {
	return f.home, nil
}

func (f *fakeFS) WriteTemp(pattern string, content []byte) (string, error) {
	return filepath.Join(f.home, pattern), nil
}

func (f *fakeFS) Remove(path string) error {
	return nil
}

type fakeFileInfo struct {
	name string
}

func (fi fakeFileInfo) Name() string       { return fi.name }
func (fi fakeFileInfo) Size() int64        { return 100 }
func (fi fakeFileInfo) Mode() os.FileMode  { return 0644 }
func (fi fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (fi fakeFileInfo) IsDir() bool        { return false }
func (fi fakeFileInfo) Sys() any           { return nil }

type fakeBackend struct {
	status      FingerprintStatus
	enablePlan  ActionPlan
	disablePlan ActionPlan
	verifyErr   error
	verifyCalls int
}

func (b *fakeBackend) Name() string { return b.status.OS }
func (b *fakeBackend) Inspect() (FingerprintStatus, error) {
	return b.status, nil
}
func (b *fakeBackend) PlanEnable() (ActionPlan, error) {
	return b.enablePlan, nil
}
func (b *fakeBackend) PlanDisable() (ActionPlan, error) {
	return b.disablePlan, nil
}
func (b *fakeBackend) Verify() error {
	b.verifyCalls++
	return b.verifyErr
}

type fakePrompter struct {
	choices  []string
	confirms []bool
	cIdx     int
	yIdx     int
}

func (p *fakePrompter) Choose(title string, options []string) (string, error) {
	if p.cIdx >= len(p.choices) {
		return "", nil
	}
	c := p.choices[p.cIdx]
	p.cIdx++
	return c, nil
}

func (p *fakePrompter) Confirm(prompt string) (bool, error) {
	if p.yIdx >= len(p.confirms) {
		return false, nil
	}
	v := p.confirms[p.yIdx]
	p.yIdx++
	return v, nil
}

func (p *fakePrompter) ShowScroll(title, content string) error {
	return nil
}

// --- 1. Cobra Command Hierarchy & Flag Tests ---

func TestSysSetupFingerPrintCommandHierarchy(t *testing.T) {
	foundSetup := false
	for _, c := range Cmd.Commands() {
		if c.Name() == "setup" {
			foundSetup = true
			if c.Args == nil {
				t.Fatalf("SetupCmd must define Args (cobra.NoArgs) so unknown subcommands error cleanly")
			}
			break
		}
	}
	if !foundSetup {
		t.Fatalf("expected SetupCmd ('setup') to be registered on sys.Cmd")
	}

	if FingerPrintCmd.Use != "finger-print" {
		t.Fatalf("expected FingerPrintCmd.Use == 'finger-print', got %q", FingerPrintCmd.Use)
	}
	if FingerPrintCmd.Annotations["title"] != "Fingerprint Authentication Setup" {
		t.Fatalf("unexpected title annotation: %q", FingerPrintCmd.Annotations["title"])
	}

	foundFP := false
	for _, c := range SetupCmd.Commands() {
		if c.Name() == "finger-print" {
			foundFP = true
		}
		if c.Name() == "finger=print" {
			t.Fatalf("forbidden subcommand name 'finger=print' found")
		}
	}
	if !foundFP {
		t.Fatalf("expected FingerPrintCmd ('finger-print') to be registered on SetupCmd")
	}

	if err := FingerPrintCmd.Args(FingerPrintCmd, []string{"unexpected"}); err == nil {
		t.Fatalf("expected Cobra NoArgs validation to reject positional arguments")
	}

	sug, dir := FingerPrintCmd.ValidArgsFunction(FingerPrintCmd, nil, "")
	if dir != cobra.ShellCompDirectiveNoFileComp || len(sug) == 0 {
		t.Fatalf("unexpected ValidArgsFunction result: sug=%v dir=%v", sug, dir)
	}
}

func TestSysSetupRejectsFingerEqualsPrintAndMutuallyExclusiveFlags(t *testing.T) {
	root := &cobra.Command{Use: "minhthetus-cli"}
	root.AddCommand(Cmd)
	root.SilenceErrors = true
	root.SilenceUsage = true

	root.SetArgs([]string{"sys", "setup", "finger=print"})
	if err := root.Execute(); err == nil {
		t.Fatalf("expected 'sys setup finger=print' to fail as an unknown command")
	}

	// Reset flags after test
	defer func() {
		_ = FingerPrintCmd.Flags().Set("enable", "false")
		_ = FingerPrintCmd.Flags().Set("disable", "false")
	}()
	root.SetArgs([]string{"sys", "setup", "finger-print", "--enable", "--disable"})
	if err := root.Execute(); err == nil {
		t.Fatalf("expected mutually exclusive flags --enable and --disable to return an error")
	}
}

// --- 2. macOS PAM Logic Tests ---

func TestResolveDarwinPAMTarget(t *testing.T) {
	sonomaSudo := `# sudo: auth account password session
auth       include        sudo_local
auth       sufficient     pam_smartcard.so
auth       required       pam_opendirectory.so
account    required       pam_permit.so
password   required       pam_deny.so
session    required       pam_permit.so
`
	if target, useLocal := ResolveDarwinPAMTarget("14.5", sonomaSudo); target != DarwinSudoLocalPath || !useLocal {
		t.Fatalf("expected Sonoma with sudo_local include -> %s (true), got %s (%v)", DarwinSudoLocalPath, target, useLocal)
	}

	if target, useLocal := ResolveDarwinPAMTarget("15.1", "# customized\nauth required pam_opendirectory.so\n"); target != DarwinSudoPath || useLocal {
		t.Fatalf("expected customized sudo without sudo_local include -> %s (false), got %s (%v)", DarwinSudoPath, target, useLocal)
	}

	if target, useLocal := ResolveDarwinPAMTarget("13.6.7", sonomaSudo); target != DarwinSudoPath || useLocal {
		t.Fatalf("expected Ventura 13.x -> %s (false), got %s (%v)", DarwinSudoPath, target, useLocal)
	}
}

func TestPlanDarwinPAMEnableSonomaTemplateAndReattach(t *testing.T) {
	tmpl := `# sudo_local: local config file which survives system update and is included for sudo
# uncomment following line to enable Touch ID for sudo
#auth       sufficient     pam_tid.so
`
	res, err := PlanDarwinPAMEnable(DarwinPAMEnableInput{
		TargetPath:      DarwinSudoLocalPath,
		CurrentContent:  "",
		TemplateContent: tmpl,
		HasTemplate:     true,
		PAMReattachPath: DarwinPAMReattachArm64,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.AlreadyEnabled {
		t.Fatalf("expected fresh setup not to report AlreadyEnabled")
	}
	if !IsActivePAMModule(res.NewContent, DarwinPAMTidModule) {
		t.Fatalf("expected pam_tid.so to be active:\n%s", res.NewContent)
	}
	if !IsActivePAMModule(res.NewContent, DarwinPAMReattachModule) {
		t.Fatalf("expected pam_reattach.so to be active:\n%s", res.NewContent)
	}
	if strings.Contains(res.NewContent, "required       "+DarwinPAMReattachArm64) {
		t.Fatalf("pam_reattach.so must NEVER be configured with 'required'")
	}
	if !strings.Contains(res.NewContent, "auth       optional       /opt/homebrew/lib/pam/pam_reattach.so ignore_ssh") {
		t.Fatalf("expected optional pam_reattach.so with ignore_ssh:\n%s", res.NewContent)
	}

	rIdx := findActivePAMModuleLine(res.NewContent, DarwinPAMReattachModule)
	tIdx := findActivePAMModuleLine(res.NewContent, DarwinPAMTidModule)
	if rIdx < 0 || tIdx < 0 || rIdx >= tIdx {
		t.Fatalf("expected pam_reattach.so (line %d) before pam_tid.so (line %d)", rIdx, tIdx)
	}

	// Idempotency check
	idem, err := PlanDarwinPAMEnable(DarwinPAMEnableInput{
		TargetPath:      DarwinSudoLocalPath,
		CurrentContent:  res.NewContent,
		PAMReattachPath: DarwinPAMReattachArm64,
	})
	if err != nil || !idem.AlreadyEnabled {
		t.Fatalf("expected idempotent enable, got err=%v already=%v", err, idem.AlreadyEnabled)
	}

	// Disable check
	dis, err := PlanDarwinPAMDisable(DarwinSudoLocalPath, res.NewContent)
	if err != nil {
		t.Fatalf("unexpected disable error: %v", err)
	}
	if IsActivePAMModule(dis.NewContent, DarwinPAMTidModule) || IsActivePAMModule(dis.NewContent, DarwinPAMReattachModule) {
		t.Fatalf("expected Touch ID & pam_reattach removed:\n%s", dis.NewContent)
	}
}

func TestPlanDarwinPAMPreSonomaSudoAndSafety(t *testing.T) {
	legacySudo := `# sudo: auth account password session
auth       sufficient     pam_smartcard.so
auth       required       pam_opendirectory.so
account    required       pam_permit.so
`
	res, err := PlanDarwinPAMEnable(DarwinPAMEnableInput{
		TargetPath:     DarwinSudoPath,
		CurrentContent: legacySudo,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findActivePAMModuleLine(res.NewContent, DarwinPAMTidModule) >= findActivePAMModuleLine(res.NewContent, "pam_opendirectory.so") {
		t.Fatalf("pam_tid.so must appear before pam_opendirectory.so:\n%s", res.NewContent)
	}

	// Refuses if no existing auth rule exists
	if _, err := PlanDarwinPAMEnable(DarwinPAMEnableInput{
		TargetPath:     DarwinSudoPath,
		CurrentContent: "# empty file\n",
	}); err == nil {
		t.Fatalf("expected refusal when /etc/pam.d/sudo has no existing auth rule")
	}

	// Disable preserves pam_opendirectory.so
	dis, err := PlanDarwinPAMDisable(DarwinSudoPath, res.NewContent)
	if err != nil {
		t.Fatalf("disable error: %v", err)
	}
	if !IsActivePAMModule(dis.NewContent, "pam_opendirectory.so") {
		t.Fatalf("expected pam_opendirectory.so fallback preserved")
	}
}

// --- 3. Linux PAM Logic Tests ---

func TestFindLinuxPamFprintdAndFprintdListParsing(t *testing.T) {
	fs := &fakeFS{
		files: map[string][]byte{
			"/usr/lib/x86_64-linux-gnu/security/pam_fprintd.so": []byte("ELF"),
		},
	}
	if got := FindLinuxPamFprintd(fs); got == "" {
		t.Fatalf("expected pam_fprintd.so to be found")
	}

	emptyFS := &fakeFS{files: map[string][]byte{}}
	if got := FindLinuxPamFprintd(emptyFS); got != "" {
		t.Fatalf("expected empty string when pam_fprintd.so missing, got %q", got)
	}

	noDev := DetectFprintdEnrollment("GDBus.Error:net.reactivated.Fprint.Error.NoSuchDevice: No devices available")
	if noDev.HasDevice {
		t.Fatalf("expected NoSuchDevice -> HasDevice=false")
	}

	noPrints := DetectFprintdEnrollment("User alice has no fingers enrolled for Goodix Fingerprint Sensor.")
	if !noPrints.HasDevice || noPrints.EnrolledCount != 0 {
		t.Fatalf("expected HasDevice=true, EnrolledCount=0, got %+v", noPrints)
	}

	enrolled := DetectFprintdEnrollment("Fingerprints for user alice on Goodix Fingerprint Sensor:\n - #0: right-index-finger\n - #1: left-index-finger\n")
	if !enrolled.HasDevice || enrolled.EnrolledCount != 2 {
		t.Fatalf("expected 2 enrolled fingers, got %+v", enrolled)
	}
}

func TestPlanLinuxSudoPAMEnableDistrosAndSafety(t *testing.T) {
	soPath := "/usr/lib/x86_64-linux-gnu/security/pam_fprintd.so"

	// 1. Hard stop when pam_fprintd.so missing on disk
	if _, err := PlanLinuxSudoPAMEnable(LinuxPAMEnableInput{
		SudoPath:       LinuxSudoPath,
		SudoContent:    "@include common-auth\n",
		PamFprintdPath: "",
	}); err == nil {
		t.Fatalf("expected hard error when pam_fprintd.so is missing on disk")
	}

	// 2. Ubuntu layout: preserve pam_env.so before pam_fprintd.so before @include common-auth
	ubuntuSudo := `#%PAM-1.0
session    required   pam_env.so readenv=1 user_readenv=0
auth       required   pam_env.so readenv=1 envfile=/etc/default/locale
@include common-auth
@include common-account
`
	res, err := PlanLinuxSudoPAMEnable(LinuxPAMEnableInput{
		SudoPath:       LinuxSudoPath,
		SudoContent:    ubuntuSudo,
		PamFprintdPath: soPath,
	})
	if err != nil {
		t.Fatalf("ubuntu enable failed: %v", err)
	}
	lines := splitLinesPreserve(res.NewContent)
	envIdx, fpIdx, incIdx := -1, -1, -1
	for i, l := range lines {
		if strings.Contains(l, "pam_env.so") && strings.HasPrefix(strings.TrimSpace(l), "auth") {
			envIdx = i
		}
		if strings.Contains(l, "pam_fprintd.so") {
			fpIdx = i
		}
		if strings.TrimSpace(l) == "@include common-auth" {
			incIdx = i
		}
	}
	if !(envIdx < fpIdx && fpIdx < incIdx) {
		t.Fatalf("invalid Ubuntu PAM ordering (env=%d, fp=%d, include=%d):\n%s", envIdx, fpIdx, incIdx, res.NewContent)
	}

	// 3. Double-prompt prevention when included common-auth already contains pam_fprintd.so
	doubleRes, err := PlanLinuxSudoPAMEnable(LinuxPAMEnableInput{
		SudoPath:    LinuxSudoPath,
		SudoContent: ubuntuSudo,
		IncludedContents: map[string]string{
			"common-auth": "auth [success=1 default=ignore] pam_fprintd.so\nauth required pam_unix.so\n",
		},
		PamFprintdPath: soPath,
	})
	if err != nil || !doubleRes.AlreadyEnabled {
		t.Fatalf("expected double-prompt prevention via included common-auth, got err=%v already=%v", err, doubleRes.AlreadyEnabled)
	}

	// 4. Fedora / Arch system-auth layout
	fedoraSudo := `#%PAM-1.0
auth       include      system-auth
account    include      system-auth
`
	fedRes, err := PlanLinuxSudoPAMEnable(LinuxPAMEnableInput{
		SudoPath:       LinuxSudoPath,
		SudoContent:    fedoraSudo,
		PamFprintdPath: "/usr/lib64/security/pam_fprintd.so",
	})
	if err != nil {
		t.Fatalf("fedora enable failed: %v", err)
	}
	if !strings.Contains(fedRes.NewContent, "auth       sufficient     pam_fprintd.so\nauth       include      system-auth") {
		t.Fatalf("expected pam_fprintd.so directly before system-auth:\n%s", fedRes.NewContent)
	}

	// 5. Disable preserves fallback anchor
	dis, err := PlanLinuxSudoPAMDisable(LinuxSudoPath, res.NewContent)
	if err != nil {
		t.Fatalf("disable failed: %v", err)
	}
	if IsActivePAMModule(dis.NewContent, LinuxPamFprintdName) {
		t.Fatalf("pam_fprintd.so should be removed")
	}
	if !strings.Contains(dis.NewContent, "@include common-auth") {
		t.Fatalf("fallback @include common-auth must remain intact")
	}
}

func TestParseAuthselectCurrentAndBackupOutsideEtcPamD(t *testing.T) {
	managed, profile, fp := ParseAuthselectCurrent("Profile ID: sssd\nEnabled features:\n- with-fingerprint\n")
	if !managed || profile != "sssd" || !fp {
		t.Fatalf("unexpected authselect parse: managed=%v profile=%q fp=%v", managed, profile, fp)
	}

	managed2, _, _ := ParseAuthselectCurrent("No existing configuration detected.")
	if managed2 {
		t.Fatalf("expected unmanaged authselect when no existing configuration detected")
	}

	home := t.TempDir()
	backupDir := GetPAMBackupDir(home)
	if strings.HasPrefix(backupDir, "/etc/pam.d") {
		t.Fatalf("backups must NEVER live inside /etc/pam.d/")
	}
	if !strings.HasSuffix(backupDir, filepath.Join(".minhthetus-cli", "backups", "pam")) {
		t.Fatalf("unexpected backupDir: %s", backupDir)
	}

	p1, err := SavePAMBackup(backupDir, "/etc/pam.d/sudo_local", []byte("# old\n"), time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SavePAMBackup failed: %v", err)
	}
	p2, err := SavePAMBackup(backupDir, "/etc/pam.d/sudo_local", []byte("# newer\n"), time.Date(2026, 11, 20, 10, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SavePAMBackup 2 failed: %v", err)
	}
	if latest := FindLatestPAMBackup(backupDir, "/etc/pam.d/sudo_local"); latest != p2 || latest == p1 {
		t.Fatalf("expected latest backup %s, got %s", p2, latest)
	}
}

// --- 4. Windows Diagnostic & Planning Tests ---

func TestWindowsBuildDetectionIgnoresProductNameTrap(t *testing.T) {
	snap23H2 := WindowsProbeSnapshot{
		ProductName:                "Windows 10 Home Single Language",
		DisplayVersion:             "23H2",
		CurrentBuild:               22631,
		UBR:                        6199,
		WbioSrvcState:              "Stopped",
		WbioSrvcStartType:          "Manual",
		UserConsentAvailability:    "Available", // PIN/NGC only!
		WinBioAnyUnits:             []WinBioUnitInfo{{UnitID: 2, BiometricFactor: WinBioTypeVoice, Description: "Analog NUI Voice Virtual Sensor"}},
		WinBioFingerprintUnits:     nil,
		EnableLUA:                  1,
		HasEnableLUA:               true,
		ConsentPromptBehaviorAdmin: 5,
		HasConsentPromptBehavior:   true,
		PromptOnSecureDesktop:      1,
		HasPromptOnSecureDesktop:   true,
		HasNativeSudoExe:           false,
	}

	if !IsWindows11(snap23H2.CurrentBuild) {
		t.Fatalf("build 22631 must be detected as Windows 11")
	}
	if SupportsWindowsNativeSudo(snap23H2.CurrentBuild) {
		t.Fatalf("build 22631 must NOT claim native sudo.exe support")
	}
	label := FormatWindowsOSLabel(snap23H2)
	if !strings.HasPrefix(label, "Windows 11 23H2 Build 22631.6199") {
		t.Fatalf("unexpected Windows label (ProductName trap?): %q", label)
	}

	st := BuildWindowsStatus(snap23H2)
	if st.HardwareDetected || st.Enrolled || st.Enabled {
		t.Fatalf("voice virtual sensor + UserConsentVerifier PIN must NOT fake fingerprint hardware/enrollment: %+v", st)
	}
	if len(st.Limitations) == 0 {
		t.Fatalf("expected explicit Windows UAC/Secure Desktop limitations")
	}

	// Ensure PlanWindowsEnableActions does NOT flip ConsentPromptBehaviorAdmin=1 when no fingerprint sensor exists
	plan := PlanWindowsEnableActions(snap23H2)
	for _, s := range plan.Steps {
		if s.Key == "set_uac_cred_prompt" {
			t.Fatalf("must NOT set ConsentPromptBehaviorAdmin=1 when no physical fingerprint reader is present")
		}
	}
}

func TestWindows24H2WithFingerprintHardwareEnableAndDisable(t *testing.T) {
	snap24H2 := WindowsProbeSnapshot{
		CurrentBuild:               26100,
		DisplayVersion:             "24H2",
		WbioSrvcState:              "Stopped",
		WbioSrvcStartType:          "Manual",
		WinBioFingerprintUnits:     []WinBioUnitInfo{{UnitID: 1, BiometricFactor: WinBioTypeFingerprint, Description: "Synaptics WBDI"}},
		UserConsentAvailability:    "Available",
		AccountEnrolledFactors:     0,
		EnableLUA:                  1,
		HasEnableLUA:               true,
		ConsentPromptBehaviorAdmin: 5,
		HasConsentPromptBehavior:   true,
		HasNativeSudoExe:           true,
		NativeSudoMode:             "disabled",
	}

	if !SupportsWindowsNativeSudo(snap24H2.CurrentBuild) {
		t.Fatalf("build 26100 must support native sudo.exe")
	}

	plan := PlanWindowsEnableActions(snap24H2)
	if plan.AlreadySatisfied || len(plan.Steps) < 4 {
		t.Fatalf("expected start_wbiosrvc, enable_native_sudo, set_uac_cred_prompt, open_signin_settings; got %+v", plan.Steps)
	}

	tmpDir := t.TempDir()
	savedPath, err := SaveWindowsStateBackup(tmpDir, WindowsSavedState{
		ModifiedUACToggle:      true,
		PrevConsentPromptAdmin: 5,
		ModifiedNativeSudo:     true,
		PrevNativeSudoMode:     "disabled",
	})
	if err != nil || savedPath == "" {
		t.Fatalf("SaveWindowsStateBackup failed: %v", err)
	}

	loaded, _ := LoadWindowsStateBackup(tmpDir)
	snapAfter := snap24H2
	snapAfter.ConsentPromptBehaviorAdmin = 1
	snapAfter.NativeSudoMode = "normal"

	dis := PlanWindowsDisableActions(snapAfter, loaded)
	if dis.AlreadySatisfied || len(dis.Steps) != 2 {
		t.Fatalf("expected 2 Windows revert steps (UAC + sudo), got %+v", dis.Steps)
	}
}

func TestWindowsOutputParsers(t *testing.T) {
	state, start := ParseSCServiceOutput("        STATE              : 1  STOPPED", "        START_TYPE         : 3   DEMAND_START")
	if state != "Stopped" || start != "Manual" {
		t.Fatalf("unexpected SC parse: state=%q start=%q", state, start)
	}

	if devs := ParsePnPBiometricOutput("Microsoft PnP Utility\n\nNo devices were found on the system.\n"); len(devs) != 0 {
		t.Fatalf("expected 0 devices, got %v", devs)
	}
	if devs := ParsePnPBiometricOutput("Device Description: Goodix Fingerprint SPI Device\n"); len(devs) != 1 {
		t.Fatalf("expected 1 device, got %v", devs)
	}

	if mode := ParseSudoConfigState(3, true, ""); mode != "normal" {
		t.Fatalf("expected sudo reg 3 -> normal, got %q", mode)
	}
}

// --- 5. End-to-End CLI Flow & Security Invariant Tests ---

func TestRunFingerPrintWithDepsEnableConfirmCancelAndVerify(t *testing.T) {
	applied := false
	be := &fakeBackend{
		status: FingerprintStatus{
			OS:               "darwin",
			OSVersion:        "14.5",
			Supported:        true,
			Configurable:     true,
			Enabled:          true,
			HardwareDetected: true,
			Enrolled:         true,
			EnrolledCount:    2,
			Method:           "Touch ID (pam_tid.so)",
			ConfigPath:       DarwinSudoLocalPath,
		},
		enablePlan: ActionPlan{
			Title:       "Enable Touch ID for sudo",
			Description: "Enable Touch ID",
			TargetPath:  DarwinSudoLocalPath,
			PreviewDiff: "--- a/sudo_local\n+++ b/sudo_local\n+auth       sufficient     pam_tid.so",
			Apply: func() error {
				applied = true
				return nil
			},
		},
	}

	// 1. Declined confirmation does not apply
	var buf bytes.Buffer
	err := runFingerPrintWithDeps(fingerPrintOptions{Enable: true}, be, &fakePrompter{confirms: []bool{false}}, &buf)
	if err != nil || applied {
		t.Fatalf("declining confirmation must not apply changes (err=%v, applied=%v)", err, applied)
	}

	// 2. --enable --yes applies without prompting
	buf.Reset()
	err = runFingerPrintWithDeps(fingerPrintOptions{Enable: true, Yes: true}, be, &fakePrompter{}, &buf)
	if err != nil || !applied {
		t.Fatalf("--enable --yes should apply changes (err=%v, applied=%v)", err, applied)
	}

	// 3. --verify delegates to backend.Verify()
	buf.Reset()
	err = runFingerPrintWithDeps(fingerPrintOptions{Verify: true}, be, &fakePrompter{}, &buf)
	if err != nil || be.verifyCalls != 1 {
		t.Fatalf("expected Verify() call, err=%v calls=%d", err, be.verifyCalls)
	}

	// 4. Interactive menu exits cleanly on Exit
	buf.Reset()
	err = runFingerPrintWithDeps(fingerPrintOptions{}, be, &fakePrompter{
		choices: []string{"Check Status & Diagnostics", "Exit"},
	}, &buf)
	if err != nil {
		t.Fatalf("interactive menu failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Touch ID (pam_tid.so)") {
		t.Fatalf("expected status output in interactive mode")
	}
}

func TestNoPasswordCaptureOrStorageInSourceFiles(t *testing.T) {
	files := []string{
		"setup.go",
		"setup_fingerprint.go",
		"setup_fingerprint_pam.go",
		"setup_fingerprint_darwin.go",
		"setup_fingerprint_linux.go",
		"setup_fingerprint_windows.go",
		"setup_fingerprint_windows_logic.go",
		"setup_fingerprint_other.go",
	}
	forbidden := []string{
		"sudo -S",
		`"-S"`,
		"ReadPassword",
		"EchoPassword",
		"keychain",
		"SavePassword",
		"StorePassword",
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		src := string(data)
		for _, pat := range forbidden {
			if strings.Contains(src, pat) {
				t.Fatalf("forbidden password-handling pattern %q found in %s", pat, f)
			}
		}
	}
}
