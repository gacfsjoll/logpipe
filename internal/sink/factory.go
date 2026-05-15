package sink

import (
	"fmt"

	"logpipe/internal/config"
)

// Sink is the interface that all log sinks must implement.
type Sink interface {
	Write(entry map[string]any) error
}

// FromConfig constructs a Sink from a sink configuration block.
// Supported types: "http", "stdout", "file".
func FromConfig(cfg config.SinkConfig) (Sink, error) {
	switch cfg.Type {
	case "http":
		if cfg.URL == "" {
			return nil, fmt.Errorf("http sink requires a url")
		}
		return NewHTTPSink(cfg.URL, cfg.Headers), nil

	case "stdout":
		return NewStdoutSink(), nil

	case "file":
		if cfg.Path == "" {
			return nil, fmt.Errorf("file sink requires a path")
		}
		sink, err := NewFileSink(cfg.Path)
		if err != nil {
			return nil, fmt.Errorf("file sink: %w", err)
		}
		return sink, nil

	default:
		return nil, fmt.Errorf("unknown sink type %q", cfg.Type)
	}
}

// FromConfigs constructs multiple Sinks from a slice of sink configurations.
func FromConfigs(cfgs []config.SinkConfig) ([]Sink, error) {
	sinks := make([]Sink, 0, len(cfgs))
	for i, cfg := range cfgs {
		s, err := FromConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("sink[%d]: %w", i, err)
		}
		sinks = append(sinks, s)
	}
	return sinks, nil
}
