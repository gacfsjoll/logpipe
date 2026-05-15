package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SourceConfig describes a single log file source.
type SourceConfig struct {
	Path string `yaml:"path"`
	Tag  string `yaml:"tag"`
}

// SinkConfig describes a single log destination.
type SinkConfig struct {
	Type    string            `yaml:"type"`
	URL     string            `yaml:"url,omitempty"`
	Path    string            `yaml:"path,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// Config is the top-level configuration structure.
type Config struct {
	Sources []SourceConfig `yaml:"sources"`
	Sinks   []SinkConfig   `yaml:"sinks"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks that a Config has the required fields populated.
func Validate(cfg *Config) error {
	if len(cfg.Sources) == 0 {
		return fmt.Errorf("config must define at least one source")
	}
	for i, src := range cfg.Sources {
		if src.Path == "" {
			return fmt.Errorf("source[%d]: path is required", i)
		}
	}
	if len(cfg.Sinks) == 0 {
		return fmt.Errorf("config must define at least one sink")
	}
	for i, s := range cfg.Sinks {
		switch s.Type {
		case "http":
			if s.URL == "" {
				return fmt.Errorf("sink[%d]: http sink requires a url", i)
			}
		case "file":
			if s.Path == "" {
				return fmt.Errorf("sink[%d]: file sink requires a path", i)
			}
		case "stdout":
			// no additional fields required
		case "":
			return fmt.Errorf("sink[%d]: type is required", i)
		default:
			return fmt.Errorf("sink[%d]: unknown type %q", i, s.Type)
		}
	}
	return nil
}
