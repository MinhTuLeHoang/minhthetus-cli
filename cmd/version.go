package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
	"github.com/spf13/cobra"
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

// getCliBinarySize returns the size of the current CLI binary in bytes
func getCliBinarySize() int64 {
	exePath, err := os.Executable()
	if err != nil {
		return 0
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}
	fi, err := os.Stat(exePath)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// getConfigFolderSize returns the total size in bytes of the config folder (~/.minhthetus-cli/)
func getConfigFolderSize() int64 {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return 0
	}
	fi, err := os.Stat(configDir)
	if err != nil || !fi.IsDir() {
		return 0
	}

	var total int64
	_ = filepath.WalkDir(configDir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

// formatSize formats a byte count into a human-readable string (e.g. 15.5 MB)
func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// getVersionSizeString calculates and formats the total size (CLI binary + config folder)
func getVersionSizeString() string {
	cliSize := getCliBinarySize()
	configSize := getConfigFolderSize()
	totalSize := cliSize + configSize

	return fmt.Sprintf("Total size: %s (CLI: %s, Config: %s)",
		formatSize(totalSize),
		formatSize(cliSize),
		formatSize(configSize),
	)
}

func printVersion() {
	ver := getVersionString()
	date := getBuildDate()
	fmt.Printf("minhthetus-cli version %s (%s)\n", ver, date)
	fmt.Printf("https://github.com/MinhTuLeHoang/minhthetus-cli/releases/tag/%s\n", ver)
	fmt.Println(getVersionSizeString())
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display minhthetus-cli version and total installation size",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func init() {
	// Set the global version flag on the root command
	rootCmd.Version = getVersionString()

	// Build the dynamic version template string featuring the resolved commit/release date and total size
	versionTemplate := fmt.Sprintf(`{{.Name}} version {{.Version}} (%s)
https://github.com/MinhTuLeHoang/minhthetus-cli/releases/tag/{{.Version}}
%s
`, getBuildDate(), getVersionSizeString())

	// Apply the customized template to Cobra's version formatting
	rootCmd.SetVersionTemplate(versionTemplate)

	// Register version subcommand
	rootCmd.AddCommand(versionCmd)
}
