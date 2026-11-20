//go:build !windows

package cmd

import (
	"fmt"
	"path/filepath"
)

func detectParentShellProcess() (ShellType, string) {
	return ShellUnknown, ""
}

func windowsDocumentsDir(home string) string {
	return filepath.Join(home, "Documents")
}

func scheduleWindowsSelfDelete(exePath string) error {
	return fmt.Errorf("windows self-delete is not applicable on this OS")
}
