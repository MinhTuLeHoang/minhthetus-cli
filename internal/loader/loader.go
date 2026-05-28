package loader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// DiscoverCommands walks the scripts directory and adds dynamic commands to the root.
func DiscoverCommands(root *cobra.Command, scriptsDir string) error {
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".sh") {
			return nil
		}

		// Calculate command path
		rel, _ := filepath.Rel(scriptsDir, path)
		parts := strings.Split(rel, string(filepath.Separator))
		
		// Remove .sh extension from the last part
		lastIdx := len(parts) - 1
		parts[lastIdx] = strings.TrimSuffix(parts[lastIdx], ".sh")

		// Register commands recursively
		registerCommand(root, parts, path)
		return nil
	})
}

func registerCommand(root *cobra.Command, parts []string, scriptPath string) {
	current := root
	for i, part := range parts {
		isLast := i == len(parts)-1
		
		// Check if subcommand already exists
		var found *cobra.Command
		for _, cmd := range current.Commands() {
			if cmd.Name() == part {
				found = cmd
				break
			}
		}

		if found == nil {
			newCmd := &cobra.Command{
				Use:   part,
				Short: getScriptDescription(scriptPath),
			}
			
			if isLast {
				newCmd.Run = func(cmd *cobra.Command, args []string) {
					executeScript(scriptPath, args)
				}
			} else {
				// If it's a directory, it's just a container
				newCmd.Run = func(cmd *cobra.Command, args []string) {
					cmd.Help()
				}
			}
			
			current.AddCommand(newCmd)
			current = newCmd
		} else {
			current = found
			if isLast {
				// If we found a command but it was just a directory placeholder, 
				// and now we found the actual .sh file (shouldn't happen with walk but good to handle)
				if current.Run == nil || reflectFuncEmpty(current.Run) {
					current.Run = func(cmd *cobra.Command, args []string) {
						executeScript(scriptPath, args)
					}
				}
			}
		}
	}
}

func reflectFuncEmpty(f interface{}) bool {
	// Simple check for placeholder Run functions
	return f == nil
}

func getScriptDescription(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# Description:") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "# Description:"))
		}
	}
	return "No description provided"
}

func executeScript(scriptPath string, args []string) {
	fmt.Printf("\n🚀 Executing script: %s\n", filepath.Base(scriptPath))
	
	// Setup environment (Phase 3 & 5)
	// We should find constants.sh
	cwd, _ := os.Getwd()
	preamblePath := filepath.Join(cwd, "src", "generalScripts", "constants.sh")
	
	execCmd := exec.Command("bash", append([]string{scriptPath}, args...)...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin
	
	// Inject BASH_ENV for template extraction/preamble (Phase 3)
	if _, err := os.Stat(preamblePath); err == nil {
		execCmd.Env = append(os.Environ(), "BASH_ENV="+preamblePath)
	} else {
		execCmd.Env = os.Environ()
	}

	if err := execCmd.Run(); err != nil {
		// Silent exit if it was a user interrupt or expected error
		os.Exit(1)
	}
	fmt.Println("") // Spacing end
}
