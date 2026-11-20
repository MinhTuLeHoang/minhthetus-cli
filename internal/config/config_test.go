package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func setTempHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	return tmp
}

func TestReadFilePreservesSliceDefaultCompatibility(t *testing.T) {
	home := setTempHome(t)

	type dummyAccount struct {
		Title string `json:"title"`
	}
	var accounts []dummyAccount
	if err := ReadFile("git-accounts.json", &accounts); err != nil {
		t.Fatalf("ReadFile failed on missing file: %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("expected empty slice, got %v", accounts)
	}

	raw, err := os.ReadFile(filepath.Join(home, ".minhthetus-cli", "git-accounts.json"))
	if err != nil {
		t.Fatalf("expected git-accounts.json to be created: %v", err)
	}
	if string(raw) != "[]" {
		t.Fatalf("expected default content '[]', got %q", string(raw))
	}
}

func TestLoadTagPresetsSeedsObjectDefaults(t *testing.T) {
	home := setTempHome(t)

	presets, err := LoadTagPresets()
	if err != nil {
		t.Fatalf("LoadTagPresets failed: %v", err)
	}

	want := DefaultTagPresets()
	if !reflect.DeepEqual(presets, want) {
		t.Fatalf("got %+v, want %+v", presets, want)
	}

	filePath := filepath.Join(home, ".minhthetus-cli", TagPresetsFilename)
	if _, statErr := os.Stat(filePath); statErr != nil {
		t.Fatalf("expected %s to exist: %v", filePath, statErr)
	}
}

func TestLoadTagPresetsCorruptedJSONFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, TagPresetsFilename), []byte("{broken-json"), 0644); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	presets, err := LoadTagPresetsFromDir(dir)
	if !errors.Is(err, ErrTagPresetsCorrupted) {
		t.Fatalf("expected ErrTagPresetsCorrupted, got %v", err)
	}
	if !reflect.DeepEqual(presets, DefaultTagPresets()) {
		t.Fatalf("expected default presets on corrupt JSON, got %+v", presets)
	}
}

func TestLearnTagPresetsPersistence(t *testing.T) {
	setTempHome(t)

	// Learn new sandbox prefix (including trailing '*' and duplicate)
	updated, added, err := LearnTagPresets(false, " uat-v* ", "qc-v", "uat-v")
	if err != nil {
		t.Fatalf("LearnTagPresets sandbox failed: %v", err)
	}
	if !added {
		t.Fatalf("expected added=true")
	}
	wantSandbox := []string{"qc-v", "stg-v", "qc-cicd-v", "uat-v"}
	if !reflect.DeepEqual(updated.Sandbox, wantSandbox) {
		t.Fatalf("sandbox presets = %v, want %v", updated.Sandbox, wantSandbox)
	}

	// Learning an already-present prefix should report added=false
	_, addedAgain, err := LearnTagPresets(false, "uat-v")
	if err != nil {
		t.Fatalf("LearnTagPresets duplicate failed: %v", err)
	}
	if addedAgain {
		t.Fatalf("expected added=false for already-present prefix")
	}

	// Learn new prod prefix and reload from disk to verify persistence
	updatedProd, addedProd, err := LearnTagPresets(true, "hotfix-v")
	if err != nil || !addedProd {
		t.Fatalf("LearnTagPresets prod failed: added=%v err=%v", addedProd, err)
	}
	wantProd := []string{"release-v", "v", "hotfix-v"}
	if !reflect.DeepEqual(updatedProd.Prod, wantProd) {
		t.Fatalf("prod presets = %v, want %v", updatedProd.Prod, wantProd)
	}

	reloaded, err := LoadTagPresets()
	if err != nil {
		t.Fatalf("LoadTagPresets reload failed: %v", err)
	}
	if !reflect.DeepEqual(reloaded.Sandbox, wantSandbox) || !reflect.DeepEqual(reloaded.Prod, wantProd) {
		t.Fatalf("reloaded presets = %+v, want sandbox=%v prod=%v", reloaded, wantSandbox, wantProd)
	}
}

func TestLearnTagPresetsRecoversFromCorruptedJSONAndStripsMultiStars(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, TagPresetsFilename), []byte("{broken-json"), 0644); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	updated, added, err := LearnTagPresetsInDir(dir, false, "uat-v**")
	if err != nil {
		t.Fatalf("expected nil error after recovering and saving presets, got: %v", err)
	}
	if !added {
		t.Fatalf("expected added=true")
	}
	wantSandbox := []string{"qc-v", "stg-v", "qc-cicd-v", "uat-v"}
	if !reflect.DeepEqual(updated.Sandbox, wantSandbox) {
		t.Fatalf("sandbox presets = %v, want %v", updated.Sandbox, wantSandbox)
	}

	reloaded, err := LoadTagPresetsFromDir(dir)
	if err != nil {
		t.Fatalf("expected valid JSON on disk after LearnTagPresetsInDir recovery, got: %v", err)
	}
	if !reflect.DeepEqual(reloaded.Sandbox, wantSandbox) {
		t.Fatalf("reloaded sandbox = %v, want %v", reloaded.Sandbox, wantSandbox)
	}
}
