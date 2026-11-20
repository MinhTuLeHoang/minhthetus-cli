package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestReadFileWithDefault_CreatesDefaultObjectWhenMissing(t *testing.T) {
	tmpHome := t.TempDir()
	setTestHome(t, tmpHome)

	type sampleCfg struct {
		Aliases []string `json:"aliases"`
		Applied []string `json:"applied"`
	}

	var cfg sampleCfg
	defaultJSON := []byte(`{"aliases":[],"applied":["c"]}`)
	if err := ReadFileWithDefault("sample.json", defaultJSON, &cfg); err != nil {
		t.Fatalf("ReadFileWithDefault failed: %v", err)
	}

	if len(cfg.Aliases) != 0 {
		t.Fatalf("expected 0 aliases, got %d", len(cfg.Aliases))
	}
	if len(cfg.Applied) != 1 || cfg.Applied[0] != "c" {
		t.Fatalf("expected applied=['c'], got %#v", cfg.Applied)
	}

	writtenPath := filepath.Join(tmpHome, ".minhthetus-cli", "sample.json")
	if _, err := os.Stat(writtenPath); err != nil {
		t.Fatalf("expected default file to exist at %s: %v", writtenPath, err)
	}
}

func TestReadFileWithDefault_ReadsExistingWithoutOverwriting(t *testing.T) {
	tmpHome := t.TempDir()
	setTestHome(t, tmpHome)

	configDir := filepath.Join(tmpHome, ".minhthetus-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	existingPath := filepath.Join(configDir, "sample.json")
	if err := os.WriteFile(existingPath, []byte(`{"aliases":["gs"],"applied":["pull"]}`), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	type sampleCfg struct {
		Aliases []string `json:"aliases"`
		Applied []string `json:"applied"`
	}

	var cfg sampleCfg
	if err := ReadFileWithDefault("sample.json", []byte(`{"aliases":[],"applied":[]}`), &cfg); err != nil {
		t.Fatalf("ReadFileWithDefault failed: %v", err)
	}

	if len(cfg.Aliases) != 1 || cfg.Aliases[0] != "gs" {
		t.Fatalf("expected aliases=['gs'], got %#v", cfg.Aliases)
	}
	if len(cfg.Applied) != 1 || cfg.Applied[0] != "pull" {
		t.Fatalf("expected applied=['pull'], got %#v", cfg.Applied)
	}
}

func TestReadFile_BackwardsCompatibleSliceDefault(t *testing.T) {
	tmpHome := t.TempDir()
	setTestHome(t, tmpHome)

	type item struct {
		Name string `json:"name"`
	}

	var items []item
	if err := ReadFile("items.json", &items); err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %#v", items)
	}

	if err := WriteFile("items.json", []item{{Name: "work"}}); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var loaded []item
	if err := ReadFile("items.json", &loaded); err != nil {
		t.Fatalf("ReadFile after WriteFile failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Name != "work" {
		t.Fatalf("expected loaded=[{work}], got %#v", loaded)
	}
}
