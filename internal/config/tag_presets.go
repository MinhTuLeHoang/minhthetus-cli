package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// TagPresetsFilename is the filename under ~/.minhthetus-cli/ storing global tag prefix presets.
const TagPresetsFilename = "git-tag-presets.json"

// ErrTagPresetsCorrupted indicates that git-tag-presets.json could not be parsed and defaults were returned.
var ErrTagPresetsCorrupted = errors.New("corrupted tag presets configuration")

// TagPresets stores default and user-learned tag prefixes grouped by environment profile.
type TagPresets struct {
	Sandbox []string `json:"sandbox"`
	Prod    []string `json:"prod"`
}

// DefaultTagPresets returns the built-in preset prefixes for sandbox and production deployments.
func DefaultTagPresets() TagPresets {
	return TagPresets{
		Sandbox: []string{"qc-v", "stg-v", "qc-cicd-v"},
		Prod:    []string{"release-v", "v"},
	}
}

func normalizePresetsList(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, raw := range items {
		cleaned := strings.TrimSpace(raw)
		cleaned = strings.TrimRight(cleaned, "*")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" || seen[cleaned] {
			continue
		}
		seen[cleaned] = true
		result = append(result, cleaned)
	}
	return result
}

func normalizeTagPresets(p TagPresets) TagPresets {
	defaults := DefaultTagPresets()
	sandbox := normalizePresetsList(p.Sandbox)
	if len(sandbox) == 0 {
		sandbox = defaults.Sandbox
	}
	prod := normalizePresetsList(p.Prod)
	if len(prod) == 0 {
		prod = defaults.Prod
	}
	return TagPresets{
		Sandbox: sandbox,
		Prod:    prod,
	}
}

// LoadTagPresets loads ~/.minhthetus-cli/git-tag-presets.json, seeding defaults when missing.
// If the file is corrupted JSON, it returns DefaultTagPresets() along with a wrapped ErrTagPresetsCorrupted.
func LoadTagPresets() (TagPresets, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return DefaultTagPresets(), err
	}
	return LoadTagPresetsFromDir(dir)
}

// LoadTagPresetsFromDir loads git-tag-presets.json from dir, seeding defaults when missing.
func LoadTagPresetsFromDir(dir string) (TagPresets, error) {
	defaults := DefaultTagPresets()
	defaultBytes, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return defaults, err
	}

	var loaded TagPresets
	if err := ReadFileInDir(dir, TagPresetsFilename, &loaded, defaultBytes); err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
			return defaults, fmt.Errorf("%w: %v", ErrTagPresetsCorrupted, err)
		}
		return defaults, err
	}

	return normalizeTagPresets(loaded), nil
}

// SaveTagPresets writes presets to ~/.minhthetus-cli/git-tag-presets.json.
func SaveTagPresets(presets TagPresets) error {
	dir, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	return SaveTagPresetsToDir(dir, presets)
}

// SaveTagPresetsToDir writes normalized presets to dir/git-tag-presets.json.
func SaveTagPresetsToDir(dir string, presets TagPresets) error {
	return WriteFileInDir(dir, TagPresetsFilename, normalizeTagPresets(presets))
}

// LearnTagPresets appends any new prefixes to the sandbox or prod preset list in ~/.minhthetus-cli/git-tag-presets.json.
// Returns the updated presets, whether any new prefix was added, and any I/O or parse warning error.
func LearnTagPresets(isProd bool, prefixes ...string) (TagPresets, bool, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return DefaultTagPresets(), false, err
	}
	return LearnTagPresetsInDir(dir, isProd, prefixes...)
}

// LearnTagPresetsInDir appends any new prefixes to the target preset slice in dir/git-tag-presets.json,
// normalizing, deduplicating, and preserving order.
func LearnTagPresetsInDir(dir string, isProd bool, prefixes ...string) (TagPresets, bool, error) {
	current, loadErr := LoadTagPresetsFromDir(dir)
	if loadErr != nil && !errors.Is(loadErr, ErrTagPresetsCorrupted) {
		return current, false, loadErr
	}

	candidates := normalizePresetsList(prefixes)
	if len(candidates) == 0 {
		return current, false, loadErr
	}

	var target *[]string
	if isProd {
		target = &current.Prod
	} else {
		target = &current.Sandbox
	}

	existing := make(map[string]bool, len(*target))
	for _, p := range *target {
		existing[p] = true
	}

	added := false
	for _, c := range candidates {
		if !existing[c] {
			*target = append(*target, c)
			existing[c] = true
			added = true
		}
	}

	if !added {
		return current, false, loadErr
	}

	if err := SaveTagPresetsToDir(dir, current); err != nil {
		return current, false, err
	}

	return current, true, nil
}
