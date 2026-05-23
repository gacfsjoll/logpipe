package sink

import (
	"errors"
	"fmt"

	"logpipe/internal/config"
)

// Sink is the common interface every output adapter must satisfy.
type Sink interface {
	Write(entry interface{ GetFields() map[string]any }) error
}

// FromConfig constructs a single Sink from a SinkConfig.
func FromConfig(cfg config.SinkConfig) (interface{ Write(e interface{ GetFields() map[string]any }) error }, error) {
	switch cfg.Type {
	case "stdout":
		return NewStdoutSink(), nil
	case "http":
		if cfg.URL == "" {
			return nil, errors.New("factory: http sink requires a url")
		}
		return NewHTTPSink(cfg)
	case "file":
		if cfg.Path == "" {
			return nil, errors.New("factory: file sink requires a path")
		}
		return NewFileSink(cfg.Path)
	case "buffered":
		if cfg.Inner == nil {
			return nil, errors.New("factory: buffered sink requires an inner sink config")
		}
		cap := cfg.Capacity
		if cap <= 0 {
			cap = 512
		}
		inner, err := FromConfig(*cfg.Inner)
		if err != nil {
			return nil, fmt.Errorf("factory: buffered inner: %w", err)
		}
		return NewBufferedSink(inner, cap)
	default:
		return nil, fmt.Errorf("factory: unknown sink type %q", cfg.Type)
	}
}

// FromConfigs constructs multiple sinks from a slice of SinkConfig.
func FromConfigs(cfgs []config.SinkConfig) ([]interface{ Write(e interface{ GetFields() map[string]any }) error }, error) {
	out := make([]interface{ Write(e interface{ GetFields() map[string]any }) error }, 0, len(cfgs))
	for _, c := range cfgs {
		s, err := FromConfig(c)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
