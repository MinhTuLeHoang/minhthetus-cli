//go:build windows

package sys

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stubWindowsRunner struct {
	lookPaths map[string]string
	outputs   map[string][]byte
	errors    map[string]error
	calls     []string
	onCall    func(name string, args []string) error
}

func (r *stubWindowsRunner) LookPath(file string) (string, error) {
	if p, ok := r.lookPaths[file]; ok {
		return p, nil
	}
	return "", os.ErrNotExist
}

func (r *stubWindowsRunner) Output(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, key)
	if r.onCall != nil {
		if err := r.onCall(name, args); err != nil {
			return nil, err
		}
	}
	if err, ok := r.errors[key]; ok {
		return nil, err
	}
	return r.outputs[key], nil
}

func (r *stubWindowsRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return r.Output(name, args...)
}

func (r *stubWindowsRunner) RunInteractive(name string, args ...string) error {
	key := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, key)
	if r.onCall != nil {
		return r.onCall(name, args)
	}
	if err, ok := r.errors[key]; ok {
		return err
	}
	return nil
}

func TestWindowsBackendIncrementalBackupAndErrorSurfacing(t *testing.T) {
	home := t.TempDir()
	backupDir := GetPAMBackupDir(home)
	fs := &fakeFS{files: map[string][]byte{}, home: home}

	snap := WindowsProbeSnapshot{
		CurrentBuild:               26100,
		WinBioFingerprintUnits:     []WinBioUnitInfo{{UnitID: 1, BiometricFactor: WinBioTypeFingerprint, Description: "Synaptics"}},
		WbioSrvcState:              "Stopped",
		WbioSrvcStartType:          "Disabled",
		HasNativeSudoExe:           true,
		NativeSudoMode:             "disabled",
		EnableLUA:                  1,
		HasEnableLUA:               true,
		ConsentPromptBehaviorAdmin: 5,
		HasConsentPromptBehavior:   true,
	}
	wPlan := PlanWindowsEnableActions(snap)

	// 1. Step 1 (enable_wbiosrvc) succeeds, Step 2 (enable_native_sudo) UAC cancelled -> Step 1 rollback state MUST be persisted
	runner := &stubWindowsRunner{}
	psCalls := 0
	runner.onCall = func(name string, args []string) error {
		if name == "powershell" && len(args) >= 4 {
			psCmd := args[3]
			if !strings.Contains(psCmd, "$ErrorActionPreference = 'Stop'") || !strings.Contains(psCmd, "if (-not $p) { exit 1 }") {
				t.Fatalf("expected BuildWindowsElevationPSCommand wrapper, got: %s", psCmd)
			}
			psCalls++
			if psCalls == 1 {
				return nil
			}
			return errors.New("UAC cancelled")
		}
		return nil
	}

	b := &windowsBackend{runner: runner, fs: fs}
	err := b.applyWindowsEnablePlan(wPlan, snap, backupDir, true)
	if err == nil {
		t.Fatalf("expected error when step 2 fails")
	}

	loaded, path := LoadWindowsStateBackup(backupDir)
	if loaded == nil || !loaded.ModifiedWbioSrvc || loaded.ModifiedNativeSudo || loaded.ModifiedUACToggle || path != filepath.Join(backupDir, WindowsBackupStateFile) {
		t.Fatalf("expected incremental backup with ONLY ModifiedWbioSrvc=true, got %+v (path=%q)", loaded, path)
	}

	// 2. open_signin_settings error must be surfaced, not discarded
	runner2 := &stubWindowsRunner{
		onCall: func(name string, args []string) error {
			if name == "cmd" && len(args) >= 3 && args[2] == WindowsSignInSettingsURI {
				return errors.New("shell launch failed")
			}
			return nil
		},
	}
	b2 := &windowsBackend{runner: runner2, fs: fs}
	snapSettings := WindowsProbeSnapshot{
		CurrentBuild:               22631,
		WinBioFingerprintUnits:     []WinBioUnitInfo{{UnitID: 1, BiometricFactor: WinBioTypeFingerprint}},
		WbioSrvcState:              "Running",
		WbioSrvcStartType:          "Manual",
		EnableLUA:                  1,
		HasEnableLUA:               true,
		ConsentPromptBehaviorAdmin: 1,
		HasConsentPromptBehavior:   true,
	}
	settingsPlan := PlanWindowsEnableActions(snapSettings)
	if err := b2.applyWindowsEnablePlan(settingsPlan, snapSettings, backupDir, true); err == nil || !strings.Contains(err.Error(), WindowsSignInSettingsURI) {
		t.Fatalf("expected surfaced error from open_signin_settings, got %v", err)
	}

	// 3. Disable plan partial failure preserves remaining Modified* flag, and full success clears backup file
	savedForDisable := &WindowsSavedState{
		ModifiedUACToggle:      true,
		PrevConsentPromptAdmin: 5,
		ModifiedNativeSudo:     true,
		PrevNativeSudoMode:     "disabled",
	}
	_, _ = SaveWindowsStateBackup(backupDir, *savedForDisable)
	snapAfter := snap
	snapAfter.ConsentPromptBehaviorAdmin = 1
	snapAfter.NativeSudoMode = "normal"
	disPlan := PlanWindowsDisableActions(snapAfter, savedForDisable)

	disCalls := 0
	runner3 := &stubWindowsRunner{
		onCall: func(name string, args []string) error {
			disCalls++
			if disCalls == 1 {
				return nil
			}
			return errors.New("sudo revert failed")
		},
	}
	b3 := &windowsBackend{runner: runner3, fs: fs}
	if err := b3.applyWindowsDisablePlan(disPlan, snapAfter, savedForDisable, backupDir, true); err == nil {
		t.Fatalf("expected disable step 2 failure")
	}
	afterPartialDisable, _ := LoadWindowsStateBackup(backupDir)
	if afterPartialDisable == nil || afterPartialDisable.ModifiedUACToggle || !afterPartialDisable.ModifiedNativeSudo {
		t.Fatalf("expected ModifiedUACToggle=false, ModifiedNativeSudo=true after partial disable, got %+v", afterPartialDisable)
	}

	// Full disable success removes the backup file cleanly
	b4 := &windowsBackend{runner: &stubWindowsRunner{}, fs: fs}
	disPlan2 := PlanWindowsDisableActions(snapAfter, afterPartialDisable)
	if err := b4.applyWindowsDisablePlan(disPlan2, snapAfter, afterPartialDisable, backupDir, true); err != nil {
		t.Fatalf("expected clean full disable, got %v", err)
	}
	if rem, _ := LoadWindowsStateBackup(backupDir); rem != nil {
		t.Fatalf("expected backup file removed after full disable, got %+v", rem)
	}
}
