package sink_test

import (
	"testing"

	"logpipe/internal/config"
	"logpipe/internal/sink"
)

func TestFromConfig_StdoutSink(t *testing.T) {
	cfg := config.SinkConfig{Type: "stdout"}
	s, err := sink.FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
}

func TestFromConfig_HTTPSink(t *testing.T) {
	cfg := config.SinkConfig{Type: "http", URL: "http://localhost:9200/logs"}
	s, err := sink.FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
}

func TestFromConfig_HTTPSink_MissingURL(t *testing.T) {
	cfg := config.SinkConfig{Type: "http"}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestFromConfig_FileSink(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.SinkConfig{Type: "file", Path: tmp + "/out.log"}
	s, err := sink.FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sink")
	}
}

func TestFromConfig_FileSink_MissingPath(t *testing.T) {
	cfg := config.SinkConfig{Type: "file"}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestFromConfig_UnknownType(t *testing.T) {
	cfg := config.SinkConfig{Type: "kafka"}
	_, err := sink.FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unknown sink type")
	}
}

func TestFromConfigs_Multiple(t *testing.T) {
	tmp := t.TempDir()
	cfgs := []config.SinkConfig{
		{Type: "stdout"},
		{Type: "file", Path: tmp + "/multi.log"},
	}
	sinks, err := sink.FromConfigs(cfgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sinks) != 2 {
		t.Fatalf("expected 2 sinks, got %d", len(sinks))
	}
}

func TestFromConfigs_ErrorPropagates(t *testing.T) {
	cfgs := []config.SinkConfig{
		{Type: "stdout"},
		{Type: "http"}, // missing URL
	}
	_, err := sink.FromConfigs(cfgs)
	if err == nil {
		t.Fatal("expected error to propagate from invalid sink config")
	}
}
