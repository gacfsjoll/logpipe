package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SinkType represents the type of log sink.
type SinkType string

const (
	SinkTypeStdout SinkType = "stdout"
	SinkTypeHTTP   SinkType = "http"
)

// SinkConfig holds configuration for a single output sink.
type SinkConfig struct {
	Name    string   `yaml:"name"`
	Type    SinkType `yaml:"type"`
	URL     string   `yaml:"url,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// SourceConfig holds configuration for a single log file source.
type SourceConfig struct {
	Path  string `yaml:"path"`
	Label string `yaml:"label,omitempty"`
}

// Config is the top-level application configuration.
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

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration is semantically valid.
func (c *Config) Validate() error {
	if len(c.Sources) == 0 {
		return fmt.Errorf("at least one source must be defined")
	}
	if len(c.Sinks) == 0 {
		return fmt.Errorf("at least one sink must be defined")
	}
	for i, s := range c.Sinks {
		if s.Name == "" {
			return fmt.Errorf("sink[%d]: name is required", i)
		}
		if s.Type != SinkTypeStdout && s.Type != SinkTypeHTTP {
			return fmt.Errorf("sink[%d]: unsupported type %q", i, s.Type)
		}
		if s.Type == SinkTypeHTTP && s.URL == "" {
			return fmt.Errorf("sink[%d]: url is required for http sink", i)
		}
	}
	for i, src := range c.Sources {
		if src.Path == "" {
			return fmt.Errorf("source[%d]: path is required", i)
		}
	}
	return nil
}
