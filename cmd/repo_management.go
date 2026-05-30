package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/MinhTuLeHoang/minhthetus-cli/internal/config"
	"github.com/spf13/cobra"
)

type Repository struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

var repoTrackCmd = &cobra.Command{
	Use:    "repo-track [path]",
	Short:  "Track a local repository",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}

		// Read package.json for metadata if present
		name := filepath.Base(absPath)
		desc := "Local tracked repository"
		pkgJSONPath := filepath.Join(absPath, "package.json")
		if data, err := os.ReadFile(pkgJSONPath); err == nil {
			var pkg struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if pkg.Name != "" {
					name = pkg.Name
				}
				if pkg.Description != "" {
					desc = pkg.Description
				}
			}
		}

		var repos []Repository
		_ = config.ReadFile("list-repo.json", &repos)

		// Check if already exists
		exists := false
		for i, r := range repos {
			if r.Path == absPath {
				// Update metadata
				repos[i].Name = name
				repos[i].Description = desc
				exists = true
				break
			}
		}

		if !exists {
			repos = append(repos, Repository{
				Name:        name,
				Path:        absPath,
				Description: desc,
			})
		}

		_ = config.WriteFile("list-repo.json", repos)
	},
}

var repoUntrackCmd = &cobra.Command{
	Use:    "repo-untrack [path]",
	Short:  "Untrack a local repository",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}

		var repos []Repository
		_ = config.ReadFile("list-repo.json", &repos)

		var newRepos []Repository
		for _, r := range repos {
			if r.Path != absPath {
				newRepos = append(newRepos, r)
			}
		}

		_ = config.WriteFile("list-repo.json", newRepos)
	},
}

func init() {
	rootCmd.AddCommand(repoTrackCmd)
	rootCmd.AddCommand(repoUntrackCmd)
}
