package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SinkType enumerates the supported output sink kinds.
type SinkType string

const (
	SinkHTTP   SinkType = "http"
	SinkStdout SinkType = "stdout"
	SinkFile   SinkType = "file"
)

// Source describes a log file to tail.
type Source struct {
	Path string `yaml:"path"`
}

// Sink describes an output destination for parsed log entries.
type Sink struct {
	Type    SinkType          `yaml:"type"`
	URL     string            `yaml:"url,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	// FilePath is used when Type == SinkFile.
	FilePath string `yaml:"file_path,omitempty"`
}

// Config is the top-level application configuration.
type Config struct {
	Sources []Source `yaml:"sources"`
	Sinks   []Sink   `yaml:"sinks"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks that the configuration is semantically valid.
func (c *Config) Validate() error {
	if len(c.Sources) == 0 {
		return errors.New("config: at least one source is required")
	}
	for i, src := range c.Sources {
		if src.Path == "" {
			return fmt.Errorf("config: source[%d]: path is required", i)
		}
	}
	for i, snk := range c.Sinks {
		switch snk.Type {
		case SinkHTTP:
			if snk.URL == "" {
				return fmt.Errorf("config: sink[%d]: url is required for http sink", i)
			}
		case SinkFile:
			if snk.FilePath == "" {
				return fmt.Errorf("config: sink[%d]: file_path is required for file sink", i)
			}
		case SinkStdout:
			// no extra fields required
		default:
			return fmt.Errorf("config: sink[%d]: unknown type %q", i, snk.Type)
		}
	}
	return nil
}
