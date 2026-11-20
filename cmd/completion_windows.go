//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// detectParentShellProcess walks ancestor processes in-process via Win32 Toolhelp32Snapshot
// (sub-millisecond, zero subprocesses) to identify the active Windows shell.
func detectParentShellProcess() (ShellType, string) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return ShellUnknown, ""
	}
	defer windows.CloseHandle(snapshot)

	type procNode struct {
		ppid uint32
		exe  string
	}

	procs := make(map[uint32]procNode, 128)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	if err := windows.Process32First(snapshot, &pe); err != nil {
		return ShellUnknown, ""
	}
	for {
		procs[pe.ProcessID] = procNode{
			ppid: pe.ParentProcessID,
			exe:  strings.ToLower(windows.UTF16ToString(pe.ExeFile[:])),
		}
		if err := windows.Process32Next(snapshot, &pe); err != nil {
			break
		}
	}

	curr := uint32(os.Getppid())
	visited := make(map[uint32]bool, 8)
	for hop := 0; hop < 8 && curr != 0 && !visited[curr]; hop++ {
		visited[curr] = true
		node, ok := procs[curr]
		if !ok {
			break
		}

		switch node.exe {
		case "pwsh.exe":
			return ShellPowerShell, "pwsh"
		case "powershell.exe":
			return ShellPowerShell, "winps"
		case "bash.exe":
			return ShellBash, ""
		case "zsh.exe":
			return ShellZsh, ""
		case "fish.exe":
			return ShellFish, ""
		case "cmd.exe":
			return ShellCmd, ""
		case "sh.exe", "make.exe", "winpty-agent.exe", "dash.exe", "env.exe", "timeout.exe", "go.exe":
			// Transparent wrapper/runner: keep walking to parent shell
			if node.ppid == curr {
				return ShellUnknown, ""
			}
			curr = node.ppid
		default:
			// Stop at GUI/terminal host boundaries so distant ancestors are not misattributed
			return ShellUnknown, ""
		}
	}

	return ShellUnknown, ""
}

// windowsDocumentsDir resolves the user's Windows Documents directory in-process
// from the User Shell Folders registry key, falling back to OneDrive/Documents or ~/Documents.
func windowsDocumentsDir(home string) string {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		defer key.Close()
		if raw, _, err := key.GetStringValue("Personal"); err == nil && strings.TrimSpace(raw) != "" {
			if expanded, err := registry.ExpandString(raw); err == nil && strings.TrimSpace(expanded) != "" {
				return filepath.Clean(expanded)
			}
			return filepath.Clean(os.ExpandEnv(raw))
		}
	}

	defaultDocs := filepath.Join(home, "Documents")
	if _, err := os.Stat(defaultDocs); err == nil {
		return defaultDocs
	}
	for _, envName := range []string{"OneDrive", "OneDriveConsumer", "OneDriveCommercial"} {
		if base := strings.TrimSpace(os.Getenv(envName)); base != "" {
			candidate := filepath.Join(base, "Documents")
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Clean(candidate)
			}
		}
	}
	return defaultDocs
}

// scheduleWindowsSelfDelete verifies directory write access and launches a detached
// cmd.exe delayed delete so a running Windows .exe can be removed cleanly upon exit.
func scheduleWindowsSelfDelete(exePath string) error {
	probe, err := os.CreateTemp(filepath.Dir(exePath), ".minhthetus-del-check-*")
	if err != nil {
		return err
	}
	probeName := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probeName)

	delCmd := exec.Command("cmd.exe")
	delCmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NO_WINDOW,
		CmdLine:       fmt.Sprintf(`cmd.exe /C "ping -n 2 127.0.0.1 >NUL & del /F /Q "%s""`, exePath),
	}
	return delCmd.Start()
}
