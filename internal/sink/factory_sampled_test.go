package sink_test

import (
	"testing"

	"logpipe/internal/config"
	"logpipe/internal/sink"
)

func TestFromConfig_SampledSink_ForwardsToStdout(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "sampled",
		Rate: 1.0,
		Inner: &config.SinkConfig{Type: "stdout"},
	}
	s, err := sink.FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
}

func TestFromConfig_SampledSink_MissingInner(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "sampled",
		Rate: 0.5,
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error when inner sink is missing")
	}
}

func TestFromConfig_SampledSink_InvalidRate(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "sampled",
		Rate: 2.0,
		Inner: &config.SinkConfig{Type: "stdout"},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid rate")
	}
}

func TestFromConfig_SampledSink_InvalidInnerType(t *testing.T) {
	cfg := config.SinkConfig{
		Type: "sampled",
		Rate: 0.5,
		Inner: &config.SinkConfig{Type: "unknown"},
	}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unknown inner sink type")
	}
}
