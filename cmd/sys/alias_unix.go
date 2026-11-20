//go:build !windows

package sys

import (
	"os"
	"runtime"
)

func resolveCurrentProfileTargets() []ProfileTarget {
	home, _ := os.UserHomeDir()
	return resolveUnixProfileTargets(
		home,
		os.Getenv("SHELL"),
		runtime.GOOS,
		func(p string) bool {
			_, err := os.Stat(p)
			return err == nil
		},
	)
}
