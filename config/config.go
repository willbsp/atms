package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	SearchDirs []string `json:"search_dirs"`
}

func Load() (Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to load config. Will create default configuration.\n")
		return saveDefaultConfig()
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse json %w", err)
	}

	return config, nil
}

func Save(config Config) error {
	path := configPath()
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory %w", err)
	}

	b, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialise json %w", err)
	}
	return os.WriteFile(path, b, 0644)
}

func saveDefaultConfig() (Config, error) {
	cfg, err := buildDefaultConfig()
	if err != nil {
		return Config{}, fmt.Errorf("building default configuration failed %w", err)
	}
	if err = Save(cfg); err != nil {
		return Config{}, fmt.Errorf("failed to save default configuration %w", err)
	}
	return cfg, nil
}

func buildDefaultConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("cannot find home directory %w", err)
	}
	return Config{
		SearchDirs: []string{filepath.Join(home, "repos"), filepath.Join(home, "Developer")},
	}, nil
}

func configPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "atns", "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "atns", "config.json")
}
