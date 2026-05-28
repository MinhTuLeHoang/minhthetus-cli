package project

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PackageJSON struct {
	Version string `json:"version"`
}

// GetVersion returns the version from package.json in the current directory.
func GetVersion() (string, error) {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return "", err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", err
	}

	return pkg.Version, nil
}

// BumpVersion calculates the next version based on the increment type.
func BumpVersion(current string, incrementType string) (string, error) {
	v := strings.Split(current, ".")
	if len(v) != 3 {
		return "", fmt.Errorf("invalid version format: %s", current)
	}

	major, _ := strconv.Atoi(v[0])
	minor, _ := strconv.Atoi(v[1])
	patch, _ := strconv.Atoi(v[2])

	switch incrementType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("invalid increment type: %s", incrementType)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

// UpdateVersion updates the version in package.json.
func UpdateVersion(newVersion string) error {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return err
	}

	// Simple string replacement would be fragile, so we use json.MarshalIndent
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}
	pkg["version"] = newVersion
	
	newData, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile("package.json", newData, 0644)
}
