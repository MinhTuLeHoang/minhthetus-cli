package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GetConfigDir returns the absolute path to the configuration directory: ~/.minhthetus-cli/
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".minhthetus-cli"), nil
}

// GetConfigFilePath returns the absolute path to a file inside ~/.minhthetus-cli/.
func GetConfigFilePath(filename string) (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, filename), nil
}

func writeAtomicFile(filePath string, data []byte, defaultMode os.FileMode) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	mode := defaultMode
	if info, err := os.Stat(filePath); err == nil {
		mode = info.Mode()
	}

	tmpFile, err := os.CreateTemp(dir, ".minhthetus-tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		return err
	}
	committed = true
	return nil
}

// ensureConfigWithDefault makes sure the config directory and JSON file exist before any I/O,
// initializing missing files with defaultContent (falling back to "[]" when empty).
func ensureConfigWithDefault(filename string, defaultContent []byte) (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", err
		}
	}

	if len(defaultContent) == 0 {
		defaultContent = []byte("[]")
	}

	filePath := filepath.Join(configDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := writeAtomicFile(filePath, defaultContent, 0644); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// ReadFileWithDefault reads and unmarshals JSON from a file in ~/.minhthetus-cli/,
// creating the file with defaultContent if it does not yet exist.
func ReadFileWithDefault(filename string, defaultContent []byte, v interface{}) error {
	filePath, err := ensureConfigWithDefault(filename, defaultContent)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// ReadFile reads and unmarshals JSON from a file in ~/.minhthetus-cli/
func ReadFile(filename string, v interface{}) error {
	return ReadFileWithDefault(filename, []byte("[]"), v)
}

// WriteFile marshals and atomically writes JSON to a file in ~/.minhthetus-cli/
func WriteFile(filename string, v interface{}) error {
	filePath, err := GetConfigFilePath(filename)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return writeAtomicFile(filePath, data, 0644)
}
