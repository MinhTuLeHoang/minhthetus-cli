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

// EnsureConfigDir creates ~/.minhthetus-cli/ if it does not exist and returns its path.
func EnsureConfigDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", err
		}
	}

	return configDir, nil
}

// ensureConfig makes sure the config directory and default JSON file exist before any I/O
func ensureConfig(filename string) (string, error) {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(configDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, []byte("[]"), 0644); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// ReadFile reads and unmarshals JSON from a file in ~/.minhthetus-cli/
func ReadFile(filename string, v interface{}) error {
	filePath, err := ensureConfig(filename)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
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

// ReadFileWithDefault reads and unmarshals JSON from ~/.minhthetus-cli/<filename>,
// seeding the file with defaultJSON if it does not exist yet.
func ReadFileWithDefault(filename string, v interface{}, defaultJSON []byte) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	return ReadFileInDir(configDir, filename, v, defaultJSON)
}

// ReadFileInDir reads and unmarshals JSON from dir/<filename>,
// creating dir and seeding the file with defaultJSON (or "[]" if empty) when missing.
func ReadFileInDir(dir, filename string, v interface{}, defaultJSON []byte) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(dir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		seed := defaultJSON
		if len(seed) == 0 {
			seed = []byte("[]")
		}
		if err := os.WriteFile(filePath, seed, 0644); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// WriteFileInDir marshals and writes JSON to dir/<filename>, creating dir if necessary.
func WriteFileInDir(dir, filename string, v interface{}) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, filename)
	return os.WriteFile(filePath, data, 0644)
}
