//go:build darwin

package sys

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type darwinBackend struct {
	runner CommandRunner
	fs     HostFS
}

func newPlatformBackend(runner CommandRunner, fs HostFS) PlatformBackend {
	return &darwinBackend{runner: runner, fs: fs}
}

func (b *darwinBackend) Name() string {
	return "darwin"
}

func (b *darwinBackend) Inspect() (FingerprintStatus, error) {
	osVer := ""
	if out, err := b.runner.Output("sw_vers", "-productVersion"); err == nil {
		osVer = strings.TrimSpace(string(out))
	}

	sudoBytes, sudoErr := b.fs.ReadFile(DarwinSudoPath)
	sudoContent := string(sudoBytes)
	targetPath, usesSudoLocal := ResolveDarwinPAMTarget(osVer, sudoContent)

	targetContent := sudoContent
	if usesSudoLocal {
		if bLocal, err := b.fs.ReadFile(DarwinSudoLocalPath); err == nil {
			targetContent = string(bLocal)
		} else {
			targetContent = ""
		}
	}

	pamTidPath := b.findPamTidPath()
	reattachPath := b.findPamReattachPath()
	enabled := IsActivePAMModule(targetContent, DarwinPAMTidModule)
	reattachActive := IsActivePAMModule(targetContent, DarwinPAMReattachModule)

	enrolledCount := 0
	enrolled := false
	hwDetected := pamTidPath != ""
	if out, err := b.runner.Output("bioutil", "-c"); err == nil {
		if n, convErr := strconv.Atoi(strings.TrimSpace(string(out))); convErr == nil {
			enrolledCount = n
			enrolled = n > 0
			if n > 0 {
				hwDetected = true
			}
		}
	}

	touchIDUnlock := ""
	if out, err := b.runner.Output("bioutil", "-r"); err == nil {
		touchIDUnlock = strings.TrimSpace(strings.ReplaceAll(string(out), "\n", " "))
	}

	backupPath := ""
	if home, err := b.fs.UserHomeDir(); err == nil {
		backupPath = FindLatestPAMBackup(GetPAMBackupDir(home), targetPath)
	}

	st := FingerprintStatus{
		OS:               "darwin",
		OSVersion:        osVer,
		Supported:        true,
		Configurable:     true,
		Enabled:          enabled,
		BiometricReady:   pamTidPath != "" && enrolled,
		HardwareDetected: hwDetected,
		Enrolled:         enrolled,
		EnrolledCount:    enrolledCount,
		Method:           "Touch ID (pam_tid.so)",
		ConfigPath:       targetPath,
		BackupPath:       backupPath,
	}

	if sudoErr != nil {
		st.Warnings = append(st.Warnings, fmt.Sprintf("Could not read %s: %v", DarwinSudoPath, sudoErr))
	}
	if pamTidPath != "" {
		st.Details = append(st.Details, fmt.Sprintf("PAM Touch ID module: %s", pamTidPath))
	} else {
		st.Warnings = append(st.Warnings, "Built-in Touch ID PAM module (/usr/lib/pam/pam_tid.so or pam_tid.so.2) not found.")
	}
	if usesSudoLocal {
		st.Details = append(st.Details, "Target mode: Sonoma+ persistent /etc/pam.d/sudo_local (survives macOS updates)")
	} else {
		st.Details = append(st.Details, "Target mode: /etc/pam.d/sudo (legacy or sudo_local not included)")
		st.Warnings = append(st.Warnings, "Editing /etc/pam.d/sudo directly may be reset by future macOS system updates.")
	}
	if touchIDUnlock != "" {
		st.Details = append(st.Details, fmt.Sprintf("bioutil -r: %s", touchIDUnlock))
	}
	if reattachPath != "" {
		st.Details = append(st.Details, fmt.Sprintf("pam_reattach.so detected: %s (active in PAM: %v)", reattachPath, reattachActive))
	} else {
		st.Details = append(st.Details, "pam_reattach.so: not installed")
		if os.Getenv("TMUX") != "" || os.Getenv("STY") != "" || os.Getenv("TERM_PROGRAM") == "iTerm.app" {
			st.Warnings = append(st.Warnings, "Multiplexer/iTerm session detected without pam_reattach.so; Touch ID may fall back to password inside tmux/screen. Install via: brew install pam-reattach")
		}
	}

	if !enrolled {
		st.Notes = append(st.Notes, "Enroll at least one fingerprint in System Settings > Touch ID & Password for Touch ID sudo prompts.")
	}
	return st, nil
}

func (b *darwinBackend) PlanEnable() (ActionPlan, error) {
	if b.findPamTidPath() == "" {
		return ActionPlan{}, fmt.Errorf("cannot enable Touch ID: neither /usr/lib/pam/pam_tid.so nor /usr/lib/pam/pam_tid.so.2 was found")
	}

	osVer := ""
	if out, err := b.runner.Output("sw_vers", "-productVersion"); err == nil {
		osVer = strings.TrimSpace(string(out))
	}
	sudoBytes, err := b.fs.ReadFile(DarwinSudoPath)
	if err != nil {
		return ActionPlan{}, fmt.Errorf("failed to read %s: %w", DarwinSudoPath, err)
	}
	sudoContent := string(sudoBytes)
	targetPath, usesSudoLocal := ResolveDarwinPAMTarget(osVer, sudoContent)

	currentContent := sudoContent
	targetExists := true
	if usesSudoLocal {
		if localBytes, err := b.fs.ReadFile(DarwinSudoLocalPath); err == nil {
			currentContent = string(localBytes)
		} else {
			currentContent = ""
			targetExists = false
		}
	}

	templateContent := ""
	hasTemplate := false
	if usesSudoLocal {
		if tmplBytes, err := b.fs.ReadFile(DarwinSudoLocalTemplatePath); err == nil {
			templateContent = string(tmplBytes)
			hasTemplate = true
		}
	}

	reattachPath := b.findPamReattachPath()
	res, err := PlanDarwinPAMEnable(DarwinPAMEnableInput{
		TargetPath:      targetPath,
		CurrentContent:  currentContent,
		TemplateContent: templateContent,
		HasTemplate:     hasTemplate,
		PAMReattachPath: reattachPath,
	})
	if err != nil {
		return ActionPlan{}, err
	}
	if res.AlreadyEnabled {
		return ActionPlan{
			Title:            "Enable Touch ID for sudo",
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

	var warnings []string
	if !usesSudoLocal {
		warnings = append(warnings, "Target is /etc/pam.d/sudo; macOS updates may overwrite this file.")
	}
	if out, err := b.runner.Output("bioutil", "-c"); err == nil {
		if n, convErr := strconv.Atoi(strings.TrimSpace(string(out))); convErr == nil && n == 0 {
			warnings = append(warnings, "bioutil reports 0 enrolled fingerprints. Enroll a finger in System Settings > Touch ID & Password.")
		}
	}

	return ActionPlan{
		Title:             "Enable Touch ID for sudo",
		Description:       fmt.Sprintf("Configure %s with Touch ID (%s) first and password fallback.", res.TargetPath, DarwinPAMTidModule),
		TargetPath:        res.TargetPath,
		BackupPath:        backupPath,
		PreviewDiff:       res.Diff,
		Commands:          []string{fmt.Sprintf("sudo install -m 0444 -o root -g wheel <tmp> %s", res.TargetPath)},
		Warnings:          warnings,
		RequiresElevation: true,
		Apply: func() error {
			if targetExists && homeErr == nil {
				if _, err := SavePAMBackup(GetPAMBackupDir(home), res.TargetPath, []byte(currentContent), time.Now()); err != nil {
					return err
				}
			}
			tmpPath, err := b.fs.WriteTemp("minhthetus-pam-*", []byte(res.NewContent))
			if err != nil {
				return fmt.Errorf("failed to create temp PAM file: %w", err)
			}
			defer b.fs.Remove(tmpPath)

			if err := b.runner.RunInteractive("sudo", "install", "-m", "0444", "-o", "root", "-g", "wheel", tmpPath, res.TargetPath); err != nil {
				return fmt.Errorf("failed to install %s: %w", res.TargetPath, err)
			}
			return nil
		},
	}, nil
}

func (b *darwinBackend) PlanDisable() (ActionPlan, error) {
	osVer := ""
	if out, err := b.runner.Output("sw_vers", "-productVersion"); err == nil {
		osVer = strings.TrimSpace(string(out))
	}
	sudoBytes, err := b.fs.ReadFile(DarwinSudoPath)
	if err != nil {
		return ActionPlan{}, fmt.Errorf("failed to read %s: %w", DarwinSudoPath, err)
	}
	sudoContent := string(sudoBytes)

	targetPath, usesSudoLocal := ResolveDarwinPAMTarget(osVer, sudoContent)
	currentContent := sudoContent
	if usesSudoLocal {
		if localBytes, err := b.fs.ReadFile(DarwinSudoLocalPath); err == nil {
			currentContent = string(localBytes)
		} else if !IsActivePAMModule(sudoContent, DarwinPAMTidModule) {
			return ActionPlan{
				Title:            "Disable Touch ID for sudo",
				Description:      fmt.Sprintf("%s does not exist and Touch ID is not active in %s.", DarwinSudoLocalPath, DarwinSudoPath),
				TargetPath:       DarwinSudoLocalPath,
				AlreadySatisfied: true,
			}, nil
		} else {
			targetPath = DarwinSudoPath
		}
	}

	res, err := PlanDarwinPAMDisable(targetPath, currentContent)
	if err != nil {
		return ActionPlan{}, err
	}
	if res.AlreadyAbsent {
		return ActionPlan{
			Title:            "Disable Touch ID for sudo",
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

	return ActionPlan{
		Title:             "Disable Touch ID for sudo",
		Description:       fmt.Sprintf("Remove Touch ID PAM directives from %s while preserving password auth.", res.TargetPath),
		TargetPath:        res.TargetPath,
		BackupPath:        backupPath,
		PreviewDiff:       res.Diff,
		Commands:          []string{fmt.Sprintf("sudo install -m 0444 -o root -g wheel <tmp> %s", res.TargetPath)},
		RequiresElevation: true,
		Apply: func() error {
			if homeErr == nil {
				if _, err := SavePAMBackup(GetPAMBackupDir(home), res.TargetPath, []byte(currentContent), time.Now()); err != nil {
					return err
				}
			}
			tmpPath, err := b.fs.WriteTemp("minhthetus-pam-*", []byte(res.NewContent))
			if err != nil {
				return fmt.Errorf("failed to create temp PAM file: %w", err)
			}
			defer b.fs.Remove(tmpPath)

			if err := b.runner.RunInteractive("sudo", "install", "-m", "0444", "-o", "root", "-g", "wheel", tmpPath, res.TargetPath); err != nil {
				return fmt.Errorf("failed to update %s: %w", res.TargetPath, err)
			}
			return nil
		},
	}, nil
}

func (b *darwinBackend) Verify() error {
	return b.runner.RunInteractive("sh", "-c", "sudo -k && sudo -v")
}

func (b *darwinBackend) findPamTidPath() string {
	for _, p := range []string{"/usr/lib/pam/pam_tid.so", "/usr/lib/pam/pam_tid.so.2"} {
		if info, err := b.fs.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func (b *darwinBackend) findPamReattachPath() string {
	candidates := []string{DarwinPAMReattachArm64, DarwinPAMReattachAmd64}
	if runtime.GOARCH == "amd64" {
		candidates = []string{DarwinPAMReattachAmd64, DarwinPAMReattachArm64}
	}
	for _, p := range candidates {
		if info, err := b.fs.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}
