package web

import (
	"os"
)

// DetectPackageManager checks for lock files and returns the detected package manager.
func DetectPackageManager() string {
	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat("yarn.lock"); err == nil {
		return "yarn"
	}
	if _, err := os.Stat("package-lock.json"); err == nil {
		return "npm"
	}
	
	// Check parent directories as fallback (optional, but good for monorepos)
	return ""
}

// HasNodeModules checks if node_modules directory exists.
func HasNodeModules() bool {
	if _, err := os.Stat("node_modules"); err == nil {
		return true
	}
	return false
}

// GetNodeVersion reads .nvmrc if it exists.
func GetNodeVersion() string {
	data, err := os.ReadFile(".nvmrc")
	if err != nil {
		return ""
	}
	return string(data)
}
