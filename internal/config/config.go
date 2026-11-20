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
		if err := os.WriteFile(filePath, defaultContent, 0644); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// ensureConfig makes sure the config directory and default JSON file exist before any I/O
func ensureConfig(filename string) (string, error) {
	return ensureConfigWithDefault(filename, []byte("[]"))
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

// WriteFile marshals and writes JSON to a file in ~/.minhthetus-cli/
func WriteFile(filename string, v interface{}) error {
	filePath, err := ensureConfig(filename)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
