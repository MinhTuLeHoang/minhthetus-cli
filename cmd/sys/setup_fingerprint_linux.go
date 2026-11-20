//go:build linux

package sys

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type linuxBackend struct {
	runner CommandRunner
	fs     HostFS
}

func newPlatformBackend(runner CommandRunner, fs HostFS) PlatformBackend {
	return &linuxBackend{runner: runner, fs: fs}
}

func (b *linuxBackend) Name() string {
	return "linux"
}

func (b *linuxBackend) Inspect() (FingerprintStatus, error) {
	distro := "Linux"
	if osRel, err := b.fs.ReadFile("/etc/os-release"); err == nil {
		if name := ParseOSReleasePrettyName(string(osRel)); name != "" {
			distro = name
		}
	}

	pamSoPath := FindLinuxPamFprintd(b.fs)

	currentUser := os.Getenv("USER")
	if u, err := user.Current(); err == nil && u.Username != "" {
		currentUser = u.Username
	}

	var fpStatus FprintdStatus
	hasFprintdList := false
	if _, err := b.runner.LookPath("fprintd-list"); err == nil && currentUser != "" {
		hasFprintdList = true
		fpStatus = b.queryFprintdStatus(currentUser)
	}

	sudoBytes, sudoErr := b.fs.ReadFile(LinuxSudoPath)
	sudoContent := string(sudoBytes)
	includedMap := b.readIncludedPAMFiles(sudoContent)

	enabledInSudo := IsActivePAMModule(sudoContent, LinuxPamFprintdName)
	enabledViaInclude := ""
	if inc, ok := findActivePAMModuleInIncludes(sudoContent, includedMap, LinuxPamFprintdName); ok {
		enabledViaInclude = inc
	}

	hasPamAuthUpdate := false
	if _, err := b.fs.Stat(LinuxPamConfigsDir); err == nil {
		hasPamAuthUpdate = true
	}

	authselectInfo := ""
	if _, err := b.runner.LookPath("authselect"); err == nil {
		if out, err := b.runner.CombinedOutput("authselect", "current"); err == nil {
			managed, profile, fpFeat := ParseAuthselectCurrent(string(out))
			if managed {
				authselectInfo = fmt.Sprintf("profile=%s, with-fingerprint=%v", profile, fpFeat)
			}
		}
	}

	serviceState := "unknown"
	if _, err := b.runner.LookPath("systemctl"); err == nil {
		if out, err := b.runner.CombinedOutput("systemctl", "is-active", "fprintd.service"); err == nil {
			serviceState = strings.TrimSpace(string(out))
		} else if trimmed := strings.TrimSpace(string(out)); trimmed != "" {
			serviceState = trimmed
		}
	}

	backupPath := ""
	if home, err := b.fs.UserHomeDir(); err == nil {
		backupPath = FindLatestPAMBackup(GetPAMBackupDir(home), LinuxSudoPath)
	}

	enabled := enabledInSudo || enabledViaInclude != ""
	configTarget := LinuxSudoPath
	if !enabledInSudo && enabledViaInclude != "" {
		configTarget = fmt.Sprintf("%s (via %s)", LinuxSudoPath, enabledViaInclude)
	}

	st := FingerprintStatus{
		OS:               "linux",
		OSVersion:        distro,
		Supported:        true,
		Configurable:     sudoErr == nil,
		Enabled:          enabled,
		BiometricReady:   pamSoPath != "" && fpStatus.HasDevice && fpStatus.EnrolledCount > 0,
		HardwareDetected: fpStatus.HasDevice,
		Enrolled:         fpStatus.EnrolledCount > 0,
		EnrolledCount:    fpStatus.EnrolledCount,
		Method:           "fprintd (pam_fprintd.so)",
		ConfigPath:       configTarget,
		BackupPath:       backupPath,
	}

	if pamSoPath != "" {
		st.Details = append(st.Details, fmt.Sprintf("PAM module on disk: %s", pamSoPath))
	} else {
		st.Warnings = append(st.Warnings, "pam_fprintd.so not found on disk. Install fprintd PAM package before enabling.")
	}
	if hasFprintdList {
		st.Details = append(st.Details, fmt.Sprintf("fprintd-list (%s): %s", currentUser, fpStatus.Reason))
	} else {
		st.Warnings = append(st.Warnings, "fprintd-list CLI not found in PATH; install fprintd to inspect sensor & enrolled prints.")
	}
	st.Details = append(st.Details, fmt.Sprintf("fprintd.service status: %s", serviceState))
	if hasPamAuthUpdate {
		st.Details = append(st.Details, "Debian/Ubuntu pam-auth-update fprintd profile detected (/usr/share/pam-configs/fprintd)")
	}
	if authselectInfo != "" {
		st.Details = append(st.Details, fmt.Sprintf("authselect: %s", authselectInfo))
	}
	if sudoErr != nil {
		st.Warnings = append(st.Warnings, fmt.Sprintf("Cannot read %s: %v", LinuxSudoPath, sudoErr))
	}
	if fpStatus.HasDevice && fpStatus.EnrolledCount == 0 {
		st.Notes = append(st.Notes, "Run `fprintd-enroll` (or `--enable`, which can invoke it interactively) to enroll your fingerprint.")
	}
	if os.Getenv("SSH_CONNECTION") != "" {
		st.Warnings = append(st.Warnings, "SSH session detected: local fingerprint prompts over SSH wait for timeout before falling back to password.")
	}

	return st, nil
}

func (b *linuxBackend) PlanEnable() (ActionPlan, error) {
	pamSoPath := FindLinuxPamFprintd(b.fs)
	if pamSoPath == "" {
		return ActionPlan{}, fmt.Errorf("cannot enable fingerprint sudo: pam_fprintd.so was not found on disk (installing a missing PAM module would cause PAM_MODULE_UNKNOWN and lock out sudo). Install via: sudo apt install fprintd libpam-fprintd | sudo dnf install fprintd fprintd-pam | sudo pacman -S fprintd")
	}

	currentUser := os.Getenv("USER")
	if u, err := user.Current(); err == nil && u.Username != "" {
		currentUser = u.Username
	}
	if _, err := b.runner.LookPath("fprintd-list"); err != nil {
		return ActionPlan{}, fmt.Errorf("fprintd-list was not found in PATH; install fprintd so hardware and enrolled prints can be verified safely")
	}
	fpStatus := b.queryFprintdStatus(currentUser)
	if !fpStatus.HasDevice {
		return ActionPlan{}, fmt.Errorf("no supported fingerprint reader detected (%s). Check `lsusb` and `systemctl status fprintd.service`; aborting to avoid blocking sudo prompts", fpStatus.Reason)
	}

	sudoBytes, err := b.fs.ReadFile(LinuxSudoPath)
	if err != nil {
		return ActionPlan{}, fmt.Errorf("failed to read %s: %w", LinuxSudoPath, err)
	}
	sudoContent := string(sudoBytes)
	includedMap := b.readIncludedPAMFiles(sudoContent)

	res, err := PlanLinuxSudoPAMEnable(LinuxPAMEnableInput{
		SudoPath:         LinuxSudoPath,
		SudoContent:      sudoContent,
		IncludedContents: includedMap,
		PamFprintdPath:   pamSoPath,
	})
	if err != nil {
		return ActionPlan{}, err
	}
	if res.AlreadyEnabled {
		return ActionPlan{
			Title:            "Enable fprintd for sudo",
			Description:      res.Note,
			TargetPath:       res.TargetPath,
			AlreadySatisfied: true,
		}, nil
	}

	var backupPath string
	home, homeErr := b.fs.UserHomeDir()
	if homeErr == nil {
		backupPath = GetPAMBackupDir(home)
	}

	var cmds []string
	var warnings []string
	needEnroll := fpStatus.EnrolledCount == 0
	if needEnroll {
		if _, err := b.runner.LookPath("fprintd-enroll"); err == nil {
			cmds = append(cmds, "fprintd-enroll")
			warnings = append(warnings, "0 fingerprints currently enrolled; interactive `fprintd-enroll` will run before updating PAM.")
		} else {
			warnings = append(warnings, "0 fingerprints currently enrolled and `fprintd-enroll` is not in PATH; PAM will fall back to password until enrolled.")
		}
	}
	cmds = append(cmds, fmt.Sprintf("sudo install -m 0644 -o root -g root <tmp> %s", res.TargetPath))

	return ActionPlan{
		Title:             "Enable fprintd for sudo",
		Description:       fmt.Sprintf("Insert `auth sufficient pam_fprintd.so` into %s before password auth fallback.", res.TargetPath),
		TargetPath:        res.TargetPath,
		BackupPath:        backupPath,
		PreviewDiff:       res.Diff,
		Commands:          cmds,
		Warnings:          warnings,
		RequiresElevation: true,
		Apply: func() error {
			if needEnroll {
				if _, err := b.runner.LookPath("fprintd-enroll"); err == nil {
					if err := b.runner.RunInteractive("fprintd-enroll"); err != nil {
						return fmt.Errorf("fingerprint enrollment did not complete: %w", err)
					}
				}
			}
			if homeErr == nil {
				if _, err := SavePAMBackup(GetPAMBackupDir(home), res.TargetPath, []byte(sudoContent), time.Now()); err != nil {
					return err
				}
			}
			tmpPath, err := b.fs.WriteTemp("minhthetus-pam-*", []byte(res.NewContent))
			if err != nil {
				return fmt.Errorf("failed to create temp PAM file: %w", err)
			}
			defer b.fs.Remove(tmpPath)

			if err := b.runner.RunInteractive("sudo", "install", "-m", "0644", "-o", "root", "-g", "root", tmpPath, res.TargetPath); err != nil {
				return fmt.Errorf("failed to install %s: %w", res.TargetPath, err)
			}
			return nil
		},
	}, nil
}

func (b *linuxBackend) PlanDisable() (ActionPlan, error) {
	sudoBytes, err := b.fs.ReadFile(LinuxSudoPath)
	if err != nil {
		return ActionPlan{}, fmt.Errorf("failed to read %s: %w", LinuxSudoPath, err)
	}
	sudoContent := string(sudoBytes)

	if IsActivePAMModule(sudoContent, LinuxPamFprintdName) {
		res, err := PlanLinuxSudoPAMDisable(LinuxSudoPath, sudoContent)
		if err != nil {
			return ActionPlan{}, err
		}
		var backupPath string
		home, homeErr := b.fs.UserHomeDir()
		if homeErr == nil {
			backupPath = GetPAMBackupDir(home)
		}
		return ActionPlan{
			Title:             "Disable fprintd in /etc/pam.d/sudo",
			Description:       fmt.Sprintf("Remove %s from %s while preserving password authentication.", LinuxPamFprintdName, LinuxSudoPath),
			TargetPath:        res.TargetPath,
			BackupPath:        backupPath,
			PreviewDiff:       res.Diff,
			Commands:          []string{fmt.Sprintf("sudo install -m 0644 -o root -g root <tmp> %s", res.TargetPath)},
			RequiresElevation: true,
			Apply: func() error {
				if homeErr == nil {
					if _, err := SavePAMBackup(GetPAMBackupDir(home), res.TargetPath, []byte(sudoContent), time.Now()); err != nil {
						return err
					}
				}
				tmpPath, err := b.fs.WriteTemp("minhthetus-pam-*", []byte(res.NewContent))
				if err != nil {
					return fmt.Errorf("failed to create temp PAM file: %w", err)
				}
				defer b.fs.Remove(tmpPath)

				if err := b.runner.RunInteractive("sudo", "install", "-m", "0644", "-o", "root", "-g", "root", tmpPath, res.TargetPath); err != nil {
					return fmt.Errorf("failed to update %s: %w", res.TargetPath, err)
				}
				return nil
			},
		}, nil
	}

	includedMap := b.readIncludedPAMFiles(sudoContent)
	if IsActivePAMModule(includedMap["common-auth"], LinuxPamFprintdName) {
		if _, err := b.fs.Stat(LinuxPamConfigsDir); err == nil {
			return ActionPlan{
				Title:             "Disable fprintd via pam-auth-update",
				Description:       "Disable the system-wide fprintd PAM profile configured in common-auth.",
				TargetPath:        "/etc/pam.d/common-auth",
				Commands:          []string{"sudo pam-auth-update --disable fprintd"},
				RequiresElevation: true,
				Apply: func() error {
					return b.runner.RunInteractive("sudo", "pam-auth-update", "--disable", "fprintd")
				},
			}, nil
		}
	}

	if IsActivePAMModule(includedMap["system-auth"], LinuxPamFprintdName) {
		if _, err := b.runner.LookPath("authselect"); err == nil {
			if out, err := b.runner.CombinedOutput("authselect", "current"); err == nil {
				managed, _, fpFeat := ParseAuthselectCurrent(string(out))
				if managed && fpFeat {
					return ActionPlan{
						Title:             "Disable fprintd via authselect",
						Description:       "Disable the `with-fingerprint` feature in the active authselect profile.",
						TargetPath:        "/etc/pam.d/system-auth",
						Commands:          []string{"sudo authselect disable-feature with-fingerprint"},
						RequiresElevation: true,
						Apply: func() error {
							return b.runner.RunInteractive("sudo", "authselect", "disable-feature", "with-fingerprint")
						},
					}, nil
				}
			}
		}
	}

	if inc, ok := findActivePAMModuleInIncludes(sudoContent, includedMap, LinuxPamFprintdName); ok {
		incPath := filepath.Join("/etc/pam.d", inc)
		return ActionPlan{}, fmt.Errorf(
			"%s is active in included PAM file %s (not directly in %s and not managed by pam-auth-update/authselect); inspect and edit %s manually to avoid breaking system-wide auth",
			LinuxPamFprintdName, incPath, LinuxSudoPath, incPath,
		)
	}

	return ActionPlan{
		Title:            "Disable fprintd for sudo",
		Description:      fmt.Sprintf("%s is not enabled in %s or included auth stacks.", LinuxPamFprintdName, LinuxSudoPath),
		TargetPath:       LinuxSudoPath,
		AlreadySatisfied: true,
	}, nil
}

func (b *linuxBackend) Verify() error {
	return b.runner.RunInteractive("sh", "-c", "sudo -k && sudo -v")
}

func (b *linuxBackend) queryFprintdStatus(username string) FprintdStatus {
	out, err := b.runner.CombinedOutput("fprintd-list", username)
	st := DetectFprintdEnrollment(string(out))
	if err != nil && strings.TrimSpace(string(out)) == "" {
		st.HasDevice = false
		st.Reason = fmt.Sprintf("fprintd-list failed: %v", err)
	}
	return st
}

func (b *linuxBackend) readIncludedPAMFiles(sudoContent string) map[string]string {
	out := map[string]string{}
	queue := ParsePAMIncludes(sudoContent)
	visited := map[string]bool{}
	for len(queue) > 0 {
		inc := queue[0]
		queue = queue[1:]
		if visited[inc] {
			continue
		}
		visited[inc] = true
		path := filepath.Join("/etc/pam.d", inc)
		if data, err := b.fs.ReadFile(path); err == nil {
			content := string(data)
			out[inc] = content
			for _, child := range ParsePAMIncludes(content) {
				if !visited[child] {
					queue = append(queue, child)
				}
			}
		}
	}
	return out
}
