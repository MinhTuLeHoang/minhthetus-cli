package sys

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

var (
	fpStatusFlag  bool
	fpEnableFlag  bool
	fpDisableFlag bool
	fpVerifyFlag  bool
	fpYesFlag     bool
)

// CommandRunner abstracts external OS command execution for testability.
// Interactive commands always attach os.Stdin/os.Stdout/os.Stderr directly so
// passwords are never captured, buffered, or stored by the CLI.
type CommandRunner interface {
	LookPath(file string) (string, error)
	Output(name string, args ...string) ([]byte, error)
	CombinedOutput(name string, args ...string) ([]byte, error)
	RunInteractive(name string, args ...string) error
}

type execCommandRunner struct{}

func (r execCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (r execCommandRunner) Output(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

func (r execCommandRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func (r execCommandRunner) RunInteractive(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// HostFS abstracts filesystem inspection and temp file creation for testability.
type HostFS interface {
	Stat(path string) (os.FileInfo, error)
	ReadFile(path string) ([]byte, error)
	Glob(pattern string) ([]string, error)
	UserHomeDir() (string, error)
	WriteTemp(pattern string, content []byte) (string, error)
	Remove(path string) error
}

type osHostFS struct{}

func (f osHostFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (f osHostFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f osHostFS) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func (f osHostFS) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (f osHostFS) WriteTemp(pattern string, content []byte) (string, error) {
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	name := tmp.Name()
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		os.Remove(name)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	return name, nil
}

func (f osHostFS) Remove(path string) error {
	return os.Remove(path)
}

// FingerprintStatus describes the inspected biometric-first sudo/admin state for the host OS.
type FingerprintStatus struct {
	OS               string
	OSVersion        string
	Supported        bool
	Configurable     bool
	Enabled          bool
	BiometricReady   bool
	HardwareDetected bool
	Enrolled         bool
	EnrolledCount    int
	Method           string
	ConfigPath       string
	BackupPath       string
	Details          []string
	Warnings         []string
	Notes            []string
	Limitations      []string
}

// ActionPlan describes a proposed configuration mutation with preview and elevation metadata.
type ActionPlan struct {
	Title             string
	Description       string
	TargetPath        string
	BackupPath        string
	PreviewDiff       string
	Commands          []string
	Warnings          []string
	RequiresElevation bool
	AlreadySatisfied  bool
	Apply             func() error
}

// PlatformBackend encapsulates OS-specific fingerprint authentication inspection and configuration.
type PlatformBackend interface {
	Name() string
	Inspect() (FingerprintStatus, error)
	PlanEnable() (ActionPlan, error)
	PlanDisable() (ActionPlan, error)
	Verify() error
}

// UIPrompter abstracts interactive TUI selections and confirmations for testability.
type UIPrompter interface {
	Choose(title string, options []string) (string, error)
	Confirm(prompt string) (bool, error)
	ShowScroll(title, content string) error
}

type bubbleTeaPrompter struct{}

func (p bubbleTeaPrompter) Choose(title string, options []string) (string, error) {
	return ui.Choose(title, options)
}

func (p bubbleTeaPrompter) Confirm(prompt string) (bool, error) {
	return ui.Confirm(prompt, 0, false)
}

func (p bubbleTeaPrompter) ShowScroll(title, content string) error {
	return ui.ScrollView(title, content)
}

type fingerPrintOptions struct {
	Status  bool
	Enable  bool
	Disable bool
	Verify  bool
	Yes     bool
}

// FingerPrintCmd represents the `sys setup finger-print` command.
var FingerPrintCmd = &cobra.Command{
	Use:   "finger-print",
	Short: "Configure biometric-first terminal authentication with traditional fallback",
	Long: `Configures terminal admin/root elevation to prefer fingerprint authentication first
and fall back to traditional authentication (password/PIN) on failure or cancellation.

Security guarantee: This command NEVER prompts for, captures, caches, or stores user
passwords. Privileged updates delegate directly to native OS authentication (sudo or UAC).

Platform behavior:
  - macOS:   Configures Touch ID (pam_tid.so) via /etc/pam.d/sudo_local (Sonoma+)
             or /etc/pam.d/sudo, with optional tmux/screen pam_reattach.so support.
  - Linux:   Verifies fprintd & pam_fprintd.so on disk, checks enrolled prints, and
             safely configures /etc/pam.d/sudo before fallback password auth.
  - Windows: Inspects WinBio fingerprint sensors, WbioSrvc, Windows Hello readiness,
             Windows 11 24H2+ sudo.exe, and UAC policies; configures feasible settings
             while transparently reporting Windows UAC/Secure Desktop boundaries.`,
	Example: `minhthetus-cli sys setup finger-print
minhthetus-cli sys setup finger-print --status
minhthetus-cli sys setup finger-print --enable --yes
minhthetus-cli sys setup finger-print --disable
minhthetus-cli sys setup finger-print --verify`,
	Args: cobra.NoArgs,
	Annotations: map[string]string{
		"title": "Fingerprint Authentication Setup",
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"-s", "--status",
			"-e", "--enable",
			"-d", "--disable",
			"--verify",
			"-y", "--yes",
			"-h", "--help",
		}, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		opts := fingerPrintOptions{
			Status:  fpStatusFlag,
			Enable:  fpEnableFlag,
			Disable: fpDisableFlag,
			Verify:  fpVerifyFlag,
			Yes:     fpYesFlag,
		}
		backend := newPlatformBackend(execCommandRunner{}, osHostFS{})
		if err := runFingerPrintWithDeps(opts, backend, bubbleTeaPrompter{}, os.Stdout); err != nil {
			fmt.Println(ui.ErrorMessage(err.Error()))
			os.Exit(1)
		}
	},
}

func init() {
	FingerPrintCmd.Flags().BoolVarP(&fpStatusFlag, "status", "s", false, "Show current fingerprint authentication status and diagnostics")
	FingerPrintCmd.Flags().BoolVarP(&fpEnableFlag, "enable", "e", false, "Enable fingerprint-first authentication with traditional fallback")
	FingerPrintCmd.Flags().BoolVarP(&fpDisableFlag, "disable", "d", false, "Disable fingerprint-first authentication and restore prior configuration")
	FingerPrintCmd.Flags().BoolVar(&fpVerifyFlag, "verify", false, "Run a live authentication verification test")
	FingerPrintCmd.Flags().BoolVarP(&fpYesFlag, "yes", "y", false, "Skip interactive confirmation prompts before applying changes")
	FingerPrintCmd.MarkFlagsMutuallyExclusive("status", "enable", "disable", "verify")
}

func runFingerPrintWithDeps(opts fingerPrintOptions, backend PlatformBackend, prompter UIPrompter, out io.Writer) error {
	printSecurityHeader(out)

	switch {
	case opts.Status:
		st, err := backend.Inspect()
		if err != nil {
			return fmt.Errorf("failed to inspect fingerprint status: %w", err)
		}
		fmt.Fprint(out, formatStatusReport(st))
		return nil
	case opts.Enable:
		return executeEnableFlow(opts.Yes, backend, prompter, out, false)
	case opts.Disable:
		return executeDisableFlow(opts.Yes, backend, prompter, out)
	case opts.Verify:
		return executeVerifyFlow(backend, out)
	default:
		return runInteractiveFingerPrintMenu(opts.Yes, backend, prompter, out)
	}
}

func printSecurityHeader(out io.Writer) {
	fmt.Fprintf(out, "\n%s %s\n", ui.ShieldIcon, ui.BoldStyle.Render("Fingerprint Authentication Setup"))
	fmt.Fprintf(out, "%s\n\n", ui.GrayStyle().Render(ui.LockIcon+" Biometric-first with traditional fallback • Zero password capture or storage"))
}

func runInteractiveFingerPrintMenu(autoYes bool, backend PlatformBackend, prompter UIPrompter, out io.Writer) error {
	const (
		optStatus  = "Check Status & Diagnostics"
		optEnable  = "Enable / Configure Fingerprint Auth"
		optDisable = "Disable / Revert Fingerprint Auth"
		optVerify  = "Verify Biometric / Admin Auth"
		optExit    = "Exit"
	)

	for {
		choice, err := prompter.Choose(ui.LockIcon+" Select Fingerprint Authentication Action", []string{
			optStatus,
			optEnable,
			optDisable,
			optVerify,
			optExit,
		})
		if err != nil {
			return fmt.Errorf("menu error: %w", err)
		}

		if choice == "" || choice == optExit {
			fmt.Fprintln(out, ui.InfoMessage("Exiting fingerprint setup."))
			return nil
		}

		switch choice {
		case optStatus:
			st, err := backend.Inspect()
			if err != nil {
				fmt.Fprintln(out, ui.ErrorMessage(fmt.Sprintf("Inspection failed: %v", err)))
			} else {
				report := formatStatusReport(st)
				if strings.Count(report, "\n") > 32 {
					if err := prompter.ShowScroll("Fingerprint Authentication Diagnostics", report); err != nil {
						fmt.Fprint(out, report)
					}
				} else {
					fmt.Fprint(out, report)
				}
			}
		case optEnable:
			if err := executeEnableFlow(autoYes, backend, prompter, out, true); err != nil {
				fmt.Fprintln(out, ui.ErrorMessage(err.Error()))
			}
		case optDisable:
			if err := executeDisableFlow(autoYes, backend, prompter, out); err != nil {
				fmt.Fprintln(out, ui.ErrorMessage(err.Error()))
			}
		case optVerify:
			if err := executeVerifyFlow(backend, out); err != nil {
				fmt.Fprintln(out, ui.ErrorMessage(err.Error()))
			}
		}
		fmt.Fprintln(out)
	}
}

func executeEnableFlow(autoYes bool, backend PlatformBackend, prompter UIPrompter, out io.Writer, interactive bool) error {
	plan, err := backend.PlanEnable()
	if err != nil {
		return err
	}
	if plan.AlreadySatisfied {
		fmt.Fprintln(out, ui.InfoMessage(plan.Description))
		return nil
	}

	printActionPlan(out, plan)

	if !autoYes {
		confirmed, err := prompter.Confirm(fmt.Sprintf("Apply '%s'?", plan.Title))
		if err != nil {
			return fmt.Errorf("confirmation error: %w", err)
		}
		if !confirmed {
			fmt.Fprintln(out, ui.WarningMessage("Operation cancelled. No system configuration was modified."))
			return nil
		}
	}

	if plan.RequiresElevation {
		fmt.Fprintln(out, ui.InfoMessage(privilegeMessage()))
	}

	if plan.Apply == nil {
		return fmt.Errorf("internal error: action plan has no apply handler")
	}
	if err := plan.Apply(); err != nil {
		return err
	}

	fmt.Fprintln(out, ui.SuccessMessage("Configuration applied successfully."))

	st, inspectErr := backend.Inspect()
	if inspectErr == nil {
		fmt.Fprintln(out)
		fmt.Fprint(out, formatStatusReport(st))
	}

	if interactive && !autoYes {
		verifyNow, err := prompter.Confirm("Run live authentication verification now?")
		if err == nil && verifyNow {
			if err := executeVerifyFlow(backend, out); err != nil {
				fmt.Fprintln(out, ui.WarningMessage(fmt.Sprintf("Verification did not complete: %v", err)))
			}
		}
	}
	return nil
}

func executeDisableFlow(autoYes bool, backend PlatformBackend, prompter UIPrompter, out io.Writer) error {
	plan, err := backend.PlanDisable()
	if err != nil {
		return err
	}
	if plan.AlreadySatisfied {
		fmt.Fprintln(out, ui.InfoMessage(plan.Description))
		return nil
	}

	printActionPlan(out, plan)

	if !autoYes {
		confirmed, err := prompter.Confirm(fmt.Sprintf("Apply '%s'?", plan.Title))
		if err != nil {
			return fmt.Errorf("confirmation error: %w", err)
		}
		if !confirmed {
			fmt.Fprintln(out, ui.WarningMessage("Operation cancelled. No system configuration was modified."))
			return nil
		}
	}

	if plan.RequiresElevation {
		fmt.Fprintln(out, ui.InfoMessage(privilegeMessage()))
	}

	if plan.Apply == nil {
		return fmt.Errorf("internal error: action plan has no apply handler")
	}
	if err := plan.Apply(); err != nil {
		return err
	}

	fmt.Fprintln(out, ui.SuccessMessage("Fingerprint configuration reverted successfully."))

	st, inspectErr := backend.Inspect()
	if inspectErr == nil {
		fmt.Fprintln(out)
		fmt.Fprint(out, formatStatusReport(st))
	}
	return nil
}

func executeVerifyFlow(backend PlatformBackend, out io.Writer) error {
	fmt.Fprintln(out, ui.InfoMessage("Starting live authentication verification (no credentials are captured or stored)..."))
	if err := backend.Verify(); err != nil {
		return fmt.Errorf("verification failed or cancelled: %w", err)
	}
	fmt.Fprintln(out, ui.SuccessMessage("Live authentication verification succeeded."))
	return nil
}

func printActionPlan(out io.Writer, plan ActionPlan) {
	fmt.Fprintf(out, "%s %s\n", ui.CyanStyle().Bold(true).Render("Planned Action:"), ui.BoldStyle.Render(plan.Title))
	if plan.Description != "" {
		fmt.Fprintf(out, "  %s\n", plan.Description)
	}
	if plan.TargetPath != "" {
		fmt.Fprintf(out, "  %s %s\n", ui.BoldStyle.Render("Target:"), ui.CyanStyle().Render(plan.TargetPath))
	}
	if plan.BackupPath != "" {
		fmt.Fprintf(out, "  %s %s\n", ui.BoldStyle.Render("Backup:"), ui.GrayStyle().Render(plan.BackupPath))
	}
	if len(plan.Commands) > 0 {
		fmt.Fprintf(out, "  %s\n", ui.BoldStyle.Render("Commands:"))
		for _, c := range plan.Commands {
			fmt.Fprintf(out, "    %s %s\n", ui.BulletIcon, ui.GrayStyle().Render(c))
		}
	}
	for _, w := range plan.Warnings {
		fmt.Fprintf(out, "  %s\n", ui.WarningMessage(w))
	}
	if plan.PreviewDiff != "" {
		fmt.Fprintf(out, "\n%s\n%s\n", ui.BoldStyle.Render("Proposed Changes:"), ColorizeDiff(plan.PreviewDiff))
	}
	fmt.Fprintln(out)
}

func formatStatusReport(st FingerprintStatus) string {
	var sb strings.Builder

	stateBadge := ui.YellowStyle().Bold(true).Render("NOT ENABLED")
	if !st.Supported {
		stateBadge = ui.RedStyle().Bold(true).Render("UNSUPPORTED")
	} else if !st.Configurable {
		stateBadge = ui.YellowStyle().Bold(true).Render("OS-BOUNDARY / READ-ONLY")
	} else if st.Enabled {
		stateBadge = ui.GreenStyle().Bold(true).Render("ENABLED")
	}

	osLabel := st.OS
	if st.OSVersion != "" {
		osLabel = fmt.Sprintf("%s (%s)", st.OS, st.OSVersion)
	}

	sb.WriteString(ui.BoldStyle.Render("System Overview") + "\n")
	sb.WriteString(fmt.Sprintf("  %s %-18s %s\n", ui.BulletIcon, "Platform:", ui.CyanStyle().Render(osLabel)))
	sb.WriteString(fmt.Sprintf("  %s %-18s %s\n", ui.BulletIcon, "Auth Status:", stateBadge))
	if st.Method != "" {
		sb.WriteString(fmt.Sprintf("  %s %-18s %s\n", ui.BulletIcon, "Method:", st.Method))
	}
	if st.ConfigPath != "" {
		sb.WriteString(fmt.Sprintf("  %s %-18s %s\n", ui.BulletIcon, "Config Target:", ui.GrayStyle().Render(st.ConfigPath)))
	}
	if st.BackupPath != "" {
		sb.WriteString(fmt.Sprintf("  %s %-18s %s\n", ui.BulletIcon, "Latest Backup:", ui.GrayStyle().Render(st.BackupPath)))
	}

	hwText := ui.YellowStyle().Render("Not detected")
	if st.HardwareDetected {
		hwText = ui.GreenStyle().Render("Detected")
	}
	enrollText := ui.YellowStyle().Render("No enrolled fingerprints")
	if st.Enrolled {
		if st.EnrolledCount > 0 {
			enrollText = ui.GreenStyle().Render(fmt.Sprintf("Yes (%d enrolled)", st.EnrolledCount))
		} else {
			enrollText = ui.GreenStyle().Render("Yes")
		}
	}
	sb.WriteString(fmt.Sprintf("  %s %-18s %s\n", ui.BulletIcon, "Sensor Hardware:", hwText))
	sb.WriteString(fmt.Sprintf("  %s %-18s %s\n", ui.BulletIcon, "Enrolled Prints:", enrollText))

	if len(st.Details) > 0 {
		sb.WriteString("\n" + ui.BoldStyle.Render("Diagnostics") + "\n")
		for _, d := range st.Details {
			sb.WriteString(fmt.Sprintf("  %s %s\n", ui.BulletIcon, d))
		}
	}

	if len(st.Warnings) > 0 {
		sb.WriteString("\n" + ui.YellowStyle().Bold(true).Render("Warnings") + "\n")
		for _, w := range st.Warnings {
			sb.WriteString(fmt.Sprintf("  %s\n", ui.WarningMessage(w)))
		}
	}

	if len(st.Notes) > 0 {
		sb.WriteString("\n" + ui.CyanStyle().Bold(true).Render("Guidance & Notes") + "\n")
		for _, n := range st.Notes {
			sb.WriteString(fmt.Sprintf("  %s %s\n", ui.InfoIcon, n))
		}
	}

	if len(st.Limitations) > 0 {
		sb.WriteString("\n" + ui.BoldStyle.Render("OS Security Boundaries") + "\n")
		for _, l := range st.Limitations {
			sb.WriteString(fmt.Sprintf("  %s %s\n", ui.LockIcon, ui.GrayStyle().Render(l)))
		}
	}

	return sb.String()
}
