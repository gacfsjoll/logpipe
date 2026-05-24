package sink_test

import (
	"testing"

	"logpipe/internal/config"
	"logpipe/internal/sink"
)

func stdoutSinkCfg() *config.SinkConfig {
	return &config.SinkConfig{Type: "stdout"}
}

func TestFromConfig_RouterSink_Valid(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "router",
		Routes: []config.RouteConfig{
			{
				Field:  "level",
				Values: []string{"error"},
				Sink:   stdoutSinkCfg(),
			},
		},
	}
	s, err := sink.FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
}

func TestFromConfig_RouterSink_NoRoutes(t *testing.T) {
	cfg := config.SinkConfig{Type: "router"}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing routes")
	}
}

func TestFromConfig_RouterSink_MissingField(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "router",
		Routes: []config.RouteConfig{
			{Field: "", Values: []string{"error"}, Sink: stdoutSinkCfg()},
		},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestFromConfig_RouterSink_MissingValues(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "router",
		Routes: []config.RouteConfig{
			{Field: "level", Values: nil, Sink: stdoutSinkCfg()},
		},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing values")
	}
}

func TestFromConfig_RouterSink_MissingSinkConfig(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "router",
		Routes: []config.RouteConfig{
			{Field: "level", Values: []string{"error"}, Sink: nil},
		},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil inner sink config")
	}
}

func TestFromConfig_RouterSink_InvalidInnerType(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "router",
		Routes: []config.RouteConfig{
			{
				Field:  "level",
				Values: []string{"error"},
				Sink:   &config.SinkConfig{Type: "unknown_type"},
			},
		},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unknown inner sink type")
	}
}
