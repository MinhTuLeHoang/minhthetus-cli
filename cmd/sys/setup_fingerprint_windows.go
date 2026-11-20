//go:build windows

package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type windowsBackend struct {
	runner CommandRunner
	fs     HostFS
}

func newPlatformBackend(runner CommandRunner, fs HostFS) PlatformBackend {
	return &windowsBackend{runner: runner, fs: fs}
}

func (b *windowsBackend) Name() string {
	return "windows"
}

func (b *windowsBackend) Inspect() (FingerprintStatus, error) {
	snap := b.collectSnapshot()
	return BuildWindowsStatus(snap), nil
}

func (b *windowsBackend) PlanEnable() (ActionPlan, error) {
	snap := b.collectSnapshot()
	wPlan := PlanWindowsEnableActions(snap)
	if wPlan.AlreadySatisfied {
		return ActionPlan{
			Title:            wPlan.Title,
			Description:      wPlan.Description,
			Warnings:         wPlan.Warnings,
			AlreadySatisfied: true,
		}, nil
	}

	var cmds []string
	needsUAC := false
	for _, step := range wPlan.Steps {
		cmds = append(cmds, step.Command)
		if step.RequiresUAC {
			needsUAC = true
		}
	}

	backupDir := ""
	home, homeErr := b.fs.UserHomeDir()
	if homeErr == nil {
		backupDir = GetPAMBackupDir(home)
	}

	return ActionPlan{
		Title:             wPlan.Title,
		Description:       wPlan.Description,
		TargetPath:        `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`,
		BackupPath:        backupDir,
		Commands:          cmds,
		Warnings:          wPlan.Warnings,
		RequiresElevation: needsUAC && !snap.IsAdmin,
		Apply: func() error {
			return b.applyWindowsEnablePlan(wPlan, snap, backupDir, homeErr == nil)
		},
	}, nil
}

func (b *windowsBackend) applyWindowsEnablePlan(wPlan WindowsConfigPlan, snap WindowsProbeSnapshot, backupDir string, canPersist bool) error {
	saved := WindowsSavedState{}
	if canPersist && backupDir != "" {
		if prev, _ := LoadWindowsStateBackup(backupDir); prev != nil {
			saved = *prev
		}
	}
	persistBackup := func(stepKey string) error {
		if !canPersist || backupDir == "" {
			return nil
		}
		if _, err := SaveWindowsStateBackup(backupDir, saved); err != nil {
			return fmt.Errorf("step %s succeeded, but failed to persist rollback state: %w", stepKey, err)
		}
		return nil
	}
	for _, step := range wPlan.Steps {
		switch step.Key {
		case "enable_wbiosrvc":
			if err := b.runElevatedCmd(snap.IsAdmin, "sc.exe config WbioSrvc start= demand && sc.exe start WbioSrvc"); err != nil {
				return fmt.Errorf("failed to configure/start WbioSrvc: %w", err)
			}
			if !saved.ModifiedWbioSrvc {
				saved.ModifiedWbioSrvc = true
				saved.PrevWbioSrvcStartType = snap.WbioSrvcStartType
			}
			if err := persistBackup(step.Key); err != nil {
				return err
			}
		case "start_wbiosrvc":
			if err := b.runElevatedCmd(snap.IsAdmin, "sc.exe start WbioSrvc"); err != nil {
				return fmt.Errorf("failed to start WbioSrvc: %w", err)
			}
			if !saved.ModifiedWbioSrvc {
				saved.ModifiedWbioSrvc = true
				saved.PrevWbioSrvcStartType = snap.WbioSrvcStartType
			}
			if err := persistBackup(step.Key); err != nil {
				return err
			}
		case "enable_native_sudo":
			if err := b.runElevatedCmd(snap.IsAdmin, "sudo config --enable normal"); err != nil {
				return fmt.Errorf("failed to enable native sudo inline mode: %w", err)
			}
			if !saved.ModifiedNativeSudo {
				saved.ModifiedNativeSudo = true
				saved.PrevNativeSudoMode = snap.NativeSudoMode
			}
			if err := persistBackup(step.Key); err != nil {
				return err
			}
		case "set_uac_cred_prompt":
			if err := b.runElevatedCmd(snap.IsAdmin, `reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v ConsentPromptBehaviorAdmin /t REG_DWORD /d 1 /f`); err != nil {
				return fmt.Errorf("failed to update ConsentPromptBehaviorAdmin: %w", err)
			}
			if !saved.ModifiedUACToggle {
				saved.ModifiedUACToggle = true
				if snap.HasConsentPromptBehavior {
					saved.PrevConsentPromptAdmin = snap.ConsentPromptBehaviorAdmin
				} else {
					saved.PrevConsentPromptAdmin = 5
				}
			}
			if err := persistBackup(step.Key); err != nil {
				return err
			}
		case "open_signin_settings":
			if err := b.runner.RunInteractive("cmd", "/c", "start", WindowsSignInSettingsURI); err != nil {
				return fmt.Errorf("failed to open Windows Sign-in options (%s): %w", WindowsSignInSettingsURI, err)
			}
		}
	}
	return nil
}

func (b *windowsBackend) PlanDisable() (ActionPlan, error) {
	snap := b.collectSnapshot()
	var saved *WindowsSavedState
	backupDir := ""
	home, homeErr := b.fs.UserHomeDir()
	if homeErr == nil {
		backupDir = GetPAMBackupDir(home)
		saved, _ = LoadWindowsStateBackup(backupDir)
	}

	wPlan := PlanWindowsDisableActions(snap, saved)
	if wPlan.AlreadySatisfied {
		return ActionPlan{
			Title:            wPlan.Title,
			Description:      wPlan.Description,
			AlreadySatisfied: true,
		}, nil
	}

	var cmds []string
	needsUAC := false
	for _, step := range wPlan.Steps {
		cmds = append(cmds, step.Command)
		if step.RequiresUAC {
			needsUAC = true
		}
	}

	return ActionPlan{
		Title:             wPlan.Title,
		Description:       wPlan.Description,
		TargetPath:        `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`,
		BackupPath:        backupDir,
		Commands:          cmds,
		RequiresElevation: needsUAC && !snap.IsAdmin,
		Apply: func() error {
			return b.applyWindowsDisablePlan(wPlan, snap, saved, backupDir, homeErr == nil)
		},
	}, nil
}

func (b *windowsBackend) applyWindowsDisablePlan(wPlan WindowsConfigPlan, snap WindowsProbeSnapshot, saved *WindowsSavedState, backupDir string, canPersist bool) error {
	for _, step := range wPlan.Steps {
		if err := b.runElevatedCmd(snap.IsAdmin, step.Command); err != nil {
			return fmt.Errorf("failed to apply revert step (%s): %w", step.Key, err)
		}
		if canPersist && backupDir != "" && saved != nil {
			switch step.Key {
			case "restore_uac_prompt":
				saved.ModifiedUACToggle = false
			case "restore_native_sudo":
				saved.ModifiedNativeSudo = false
			}
			if _, err := SaveWindowsStateBackup(backupDir, *saved); err != nil {
				return fmt.Errorf("revert step %s succeeded, but failed to update backup state: %w", step.Key, err)
			}
		}
	}
	if canPersist && backupDir != "" && saved != nil {
		if err := SaveWindowsStateBackupReset(backupDir); err != nil {
			return fmt.Errorf("reverted system settings, but failed to clear backup state: %w", err)
		}
	}
	return nil
}

func (b *windowsBackend) Verify() error {
	avail := b.queryUserConsentAvailability()
	if !strings.EqualFold(avail, "Available") {
		return fmt.Errorf("Windows Hello UserConsentVerifier is %q (not Available); verify hardware/enrollment in Settings > Sign-in options", avail)
	}
	script := `
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$asTaskGeneric = ([System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object { $_.Name -eq 'AsTask' -and $_.GetParameters().Count -eq 1 -and $_.GetParameters()[0].ParameterType.Name -eq 'IAsyncOperation` + "`" + `1' })[0]
function Await($WinRtTask, $ResultType) {
  $asTask = $asTaskGeneric.MakeGenericMethod($ResultType)
  $netTask = $asTask.Invoke($null, @($WinRtTask))
  $netTask.Wait(-1) | Out-Null
  $netTask.Result
}
$null = [Windows.Security.Credentials.UI.UserConsentVerifier, Windows.Security.Credentials.UI, ContentType = WindowsRuntime]
$res = Await ([Windows.Security.Credentials.UI.UserConsentVerifier]::RequestVerificationAsync('minhthetus-cli biometric verification')) ([Windows.Security.Credentials.UI.UserConsentVerificationResult])
Write-Output $res.ToString()
if ($res.ToString() -ne 'Verified') { exit 1 }
`
	return b.runner.RunInteractive("powershell", "-NoProfile", "-Command", script)
}

func (b *windowsBackend) runElevatedCmd(isAdmin bool, cmdLine string) error {
	if isAdmin {
		return b.runner.RunInteractive("cmd", "/c", cmdLine)
	}
	psArgs := BuildWindowsElevationPSCommand(cmdLine)
	return b.runner.RunInteractive("powershell", "-NoProfile", "-NonInteractive", "-Command", psArgs)
}

func (b *windowsBackend) collectSnapshot() WindowsProbeSnapshot {
	snap := WindowsProbeSnapshot{
		IsAdmin: isCurrentProcessElevated(),
	}

	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE); err == nil {
		if v, _, err := k.GetStringValue("CurrentBuild"); err == nil {
			snap.CurrentBuild = ParseWindowsBuildNumber(v)
		}
		if v, _, err := k.GetStringValue("DisplayVersion"); err == nil {
			snap.DisplayVersion = v
		}
		if v, _, err := k.GetIntegerValue("UBR"); err == nil {
			snap.UBR = int(v)
		}
		if v, _, err := k.GetStringValue("ProductName"); err == nil {
			snap.ProductName = v
		}
		k.Close()
	}

	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`, registry.QUERY_VALUE); err == nil {
		if v, _, err := k.GetIntegerValue("EnableLUA"); err == nil {
			snap.EnableLUA = uint32(v)
			snap.HasEnableLUA = true
		}
		if v, _, err := k.GetIntegerValue("ConsentPromptBehaviorAdmin"); err == nil {
			snap.ConsentPromptBehaviorAdmin = uint32(v)
			snap.HasConsentPromptBehavior = true
		}
		if v, _, err := k.GetIntegerValue("PromptOnSecureDesktop"); err == nil {
			snap.PromptOnSecureDesktop = uint32(v)
			snap.HasPromptOnSecureDesktop = true
		}
		k.Close()
	}

	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\Biometrics`, registry.QUERY_VALUE); err == nil {
		if v, _, err := k.GetIntegerValue("Enabled"); err == nil && v == 0 {
			snap.BiometricsPolicyBlocked = true
		}
		k.Close()
	}

	if sidStr := currentUserSIDString(); sidStr != "" {
		subkey := `SOFTWARE\Microsoft\Windows\CurrentVersion\WinBio\AccountInfo\` + sidStr
		if k, err := registry.OpenKey(registry.LOCAL_MACHINE, subkey, registry.QUERY_VALUE); err == nil {
			snap.HasAccountInfoKey = true
			if v, _, err := k.GetIntegerValue("EnrolledFactors"); err == nil {
				snap.AccountEnrolledFactors = uint32(v)
			}
			k.Close()
		}
	}

	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = `C:\Windows`
	}
	sudoExePath := filepath.Join(systemRoot, "System32", "sudo.exe")
	if info, err := b.fs.Stat(sudoExePath); err == nil && !info.IsDir() {
		snap.HasNativeSudoExe = true
	}
	if _, err := b.runner.LookPath("gsudo"); err == nil {
		snap.HasGsudo = true
	}

	if snap.HasNativeSudoExe {
		var regVal uint32
		hasRegVal := false
		if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Sudo`, registry.QUERY_VALUE); err == nil {
			if v, _, err := k.GetIntegerValue("Enabled"); err == nil {
				regVal = uint32(v)
				hasRegVal = true
			}
			k.Close()
		}
		cliOut := ""
		if !hasRegVal {
			if out, err := b.runner.CombinedOutput(sudoExePath, "config"); err == nil {
				cliOut = string(out)
			}
		}
		snap.NativeSudoMode = ParseSudoConfigState(regVal, hasRegVal, cliOut)
	}

	qOut, _ := b.runner.CombinedOutput("sc.exe", "query", "WbioSrvc")
	qcOut, _ := b.runner.CombinedOutput("sc.exe", "qc", "WbioSrvc")
	snap.WbioSrvcState, snap.WbioSrvcStartType = ParseSCServiceOutput(string(qOut), string(qcOut))
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\WbioSrvc`, registry.QUERY_VALUE); err == nil {
		if v, _, err := k.GetIntegerValue("Start"); err == nil {
			if regStart := ParseWindowsServiceStartDWORD(uint32(v)); regStart != "Unknown" {
				snap.WbioSrvcStartType = regStart
			}
		}
		k.Close()
	}

	if pnpOut, err := b.runner.CombinedOutput("pnputil.exe", "/enum-devices", "/class", "Biometric"); err == nil {
		snap.PnPBiometricDevices = ParsePnPBiometricOutput(string(pnpOut))
	}

	snap.WinBioFingerprintUnits = enumWinBioUnits(WinBioTypeFingerprint)
	snap.WinBioAnyUnits = enumWinBioUnits(WinBioTypeAny)
	snap.UserConsentAvailability = b.queryUserConsentAvailability()

	if home, err := b.fs.UserHomeDir(); err == nil {
		_, savedPath := LoadWindowsStateBackup(GetPAMBackupDir(home))
		snap.BackupStatePath = savedPath
	}

	return snap
}

func (b *windowsBackend) queryUserConsentAvailability() string {
	script := `
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$asTaskGeneric = ([System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object { $_.Name -eq 'AsTask' -and $_.GetParameters().Count -eq 1 -and $_.GetParameters()[0].ParameterType.Name -eq 'IAsyncOperation` + "`" + `1' })[0]
function Await($WinRtTask, $ResultType) {
  $asTask = $asTaskGeneric.MakeGenericMethod($ResultType)
  $netTask = $asTask.Invoke($null, @($WinRtTask))
  $netTask.Wait(-1) | Out-Null
  $netTask.Result
}
$null = [Windows.Security.Credentials.UI.UserConsentVerifier, Windows.Security.Credentials.UI, ContentType = WindowsRuntime]
(Await ([Windows.Security.Credentials.UI.UserConsentVerifier]::CheckAvailabilityAsync()) ([Windows.Security.Credentials.UI.UserConsentVerifierAvailability])).ToString()
`
	out, err := b.runner.Output("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	if err != nil {
		return "Unknown"
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return "Unknown"
	}
	return trimmed
}

func isCurrentProcessElevated() bool {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false
	}
	defer token.Close()
	return token.IsElevated()
}

func currentUserSIDString() string {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return ""
	}
	defer token.Close()
	tu, err := token.GetTokenUser()
	if err != nil || tu == nil || tu.User.Sid == nil {
		return ""
	}
	return tu.User.Sid.String()
}

type winBioUnitSchemaRaw struct {
	UnitID           uint32
	PoolType         uint32
	BiometricFactor  uint32
	SensorSubType    uint32
	Capabilities     uint32
	DeviceInstanceID [256]uint16
	Description      [256]uint16
	Manufacturer     [256]uint16
	Model            [256]uint16
	SerialNumber     [256]uint16
	FirmwareMajor    uint32
	FirmwareMinor    uint32
}

func enumWinBioUnits(factor uint32) []WinBioUnitInfo {
	dll := windows.NewLazySystemDLL("winbio.dll")
	if err := dll.Load(); err != nil {
		return nil
	}
	enumProc := dll.NewProc("WinBioEnumBiometricUnits")
	freeProc := dll.NewProc("WinBioFree")
	if err := enumProc.Find(); err != nil {
		return nil
	}

	var unitSchema *winBioUnitSchemaRaw
	var unitCount uintptr
	hr, _, _ := enumProc.Call(
		uintptr(factor),
		uintptr(unsafe.Pointer(&unitSchema)),
		uintptr(unsafe.Pointer(&unitCount)),
	)
	if int32(hr) != 0 || unitSchema == nil || unitCount == 0 {
		if unitSchema != nil && freeProc.Find() == nil {
			_, _, _ = freeProc.Call(uintptr(unsafe.Pointer(unitSchema)))
		}
		return nil
	}
	if freeProc.Find() == nil {
		defer freeProc.Call(uintptr(unsafe.Pointer(unitSchema)))
	}

	n := int(unitCount)
	if n > 64 {
		n = 64
	}
	schemas := unsafe.Slice(unitSchema, n)
	units := make([]WinBioUnitInfo, 0, n)
	for _, s := range schemas {
		units = append(units, WinBioUnitInfo{
			UnitID:           s.UnitID,
			PoolType:         s.PoolType,
			BiometricFactor:  s.BiometricFactor,
			DeviceInstanceID: windows.UTF16ToString(s.DeviceInstanceID[:]),
			Description:      windows.UTF16ToString(s.Description[:]),
			Manufacturer:     windows.UTF16ToString(s.Manufacturer[:]),
			Model:            windows.UTF16ToString(s.Model[:]),
		})
	}
	return units
}

// SaveWindowsStateBackupReset removes or clears the saved state after revert.
func SaveWindowsStateBackupReset(backupDir string) error {
	path := filepath.Join(backupDir, WindowsBackupStateFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
