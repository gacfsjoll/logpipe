package sink_test

import (
	"testing"

	"logpipe/internal/config"
	"logpipe/internal/sink"
)

func TestFromConfig_BufferedSink_DefaultCapacity(t *testing.T) {
	cfg := config.SinkConfig{
		Type:     "buffered",
		Capacity: 0, // should default to 512
		Inner:    &config.SinkConfig{Type: "stdout"},
	}
	s, err := sink.FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
}

func TestFromConfig_BufferedSink_ExplicitCapacity(t *testing.T) {
	cfg := config.SinkConfig{
		Type:     "buffered",
		Capacity: 64,
		Inner:    &config.SinkConfig{Type: "stdout"},
	}
	_, err := sink.FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFromConfig_BufferedSink_MissingInner(t *testing.T) {
	cfg := config.SinkConfig{
		Type:     "buffered",
		Capacity: 10,
		Inner:    nil,
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error when inner sink config is nil")
	}
}

func TestFromConfig_BufferedSink_InvalidInnerType(t *testing.T) {
	cfg := config.SinkConfig{
		Type:     "buffered",
		Capacity: 10,
		Inner:    &config.SinkConfig{Type: "unknown"},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unknown inner sink type")
	}
}

func TestFromConfig_BufferedSink_InnerHTTPMissingURL(t *testing.T) {
	cfg := config.SinkConfig{
		Type:     "buffered",
		Capacity: 10,
		Inner:    &config.SinkConfig{Type: "http", URL: ""},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error when inner http sink has no url")
	}
}
