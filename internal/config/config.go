package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level application configuration.
type Config struct {
	StateFile string       `yaml:"state_file"`
	Images    []ImageEntry `yaml:"images"`
}

// ImageEntry is a single image to monitor.
type ImageEntry struct {
	Ref string `yaml:"ref"`
}

// Load reads and parses a YAML config file from the given path.
// StateFile defaults to "state.json" if not set.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	if cfg.StateFile == "" {
		cfg.StateFile = "state.json"
	}

	return &cfg, nil
}
