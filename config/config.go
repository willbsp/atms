package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	SearchDirs []string `json:"search_dirs"`
}

func DefaultConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("cannot find home directory %w", err)
	}
	return Config{
		SearchDirs: []string{filepath.Join(home, "repos"), filepath.Join(home, "Developer")},
	}, nil
}

func Load() Config {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to load config. Will create default configuration.\n")
		return createDefaultConfig()
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}

	return config
}

func Save(config Config) error {
	path := configPath()
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		panic(err)
	}

	b, _ := json.Marshal(config)
	return os.WriteFile(path, b, 0644)
}

func createDefaultConfig() Config {
	cfg, err := DefaultConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err = Save(cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}

func configPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "atns", "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "atns", "config.json")
}
