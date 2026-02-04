package core

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configDirName = "tuido"
const configFileName = "author"
const dataFileName = ".tuido"

// LoadEntries reads the .tuido file from the current directory
func LoadEntries(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if len(data) == 0 {
		return []Entry{}, nil
	}

	err = yaml.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// SaveEntries writes the entries to the .tuido file
func SaveEntries(path string, entries []Entry) error {
	data, err := yaml.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadConfig reads the author name from the config file
func LoadConfig() (Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}

	path := filepath.Join(configDir, configDirName, configFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Config{}, nil // Return empty config if not found
	}
	if err != nil {
		return Config{}, err
	}

	// The config file specifically just contains the author name in YAML format or plain text?
	// Specs say: "Check for file: ~/.config/tuido/author ... Save this to the config file."
	// It implies it might just be the name or a struct. Let's assume struct for extensibility,
	// but strictly following the spec "Save this to the config file" for "Author Name".
	// Let's stick to YAML for consistency.
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		// Fallback: try reading as plain string if YAML fails (legacy/simplicity support)
		return Config{Author: string(data)}, nil
	}
	return cfg, nil
}

// SaveConfig writes the author name to the config file
func SaveConfig(cfg Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dirPath := filepath.Join(configDir, configDirName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	path := filepath.Join(dirPath, configFileName)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
