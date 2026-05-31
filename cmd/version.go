package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
)

// getVersionString resolves the installation version dynamically from Go runtime metadata, falling back to build variables
func getVersionString() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return config.Version
}

// getBuildDate resolves the exact Git commit date dynamically from Go build metadata, falling back to build variables
func getBuildDate() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				// vcs.time format: 2026-05-29T10:00:00Z -> extract first 10 characters (YYYY-MM-DD)
				if len(setting.Value) >= 10 {
					return setting.Value[:10]
				}
				return setting.Value
			}
		}
	}
	return config.BuildDate
}

func init() {
	// Set the global version flag on the root command
	rootCmd.Version = getVersionString()

	// Build the dynamic version template string featuring the resolved commit/release date
	versionTemplate := fmt.Sprintf(`{{.Name}} version {{.Version}} (%s)
https://github.com/MinhTuLeHoang/minhthetus-cli/releases/tag/{{.Version}}
`, getBuildDate())

	// Apply the customized template to Cobra's version formatting
	rootCmd.SetVersionTemplate(versionTemplate)
}
