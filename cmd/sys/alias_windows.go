//go:build windows

package sys

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func resolveCurrentProfileTargets() []ProfileTarget {
	home, _ := os.UserHomeDir()
	docsDir, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0)
	if err != nil || docsDir == "" {
		if home != "" {
			docsDir = filepath.Join(home, "Documents")
		}
	}

	var altDocsDir string
	if home != "" {
		altDocsDir = filepath.Join(home, "Documents")
	}

	return resolveWindowsProfileTargets(
		home,
		docsDir,
		altDocsDir,
		os.Getenv("PSModulePath"),
		os.Getenv("MSYSTEM"),
		os.Getenv("SHELL"),
		exec.LookPath,
		func(p string) bool {
			_, err := os.Stat(p)
			return err == nil
		},
	)
}
