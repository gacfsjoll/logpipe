package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/logpipe/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "logpipe-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	yaml := `
sources:
  - path: /var/log/app.log
    label: app
sinks:
  - name: console
    type: stdout
`
	path := writeTemp(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Sources) != 1 || cfg.Sources[0].Path != "/var/log/app.log" {
		t.Errorf("unexpected sources: %+v", cfg.Sources)
	}
	if len(cfg.Sinks) != 1 || cfg.Sinks[0].Type != config.SinkTypeStdout {
		t.Errorf("unexpected sinks: %+v", cfg.Sinks)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidate_NoSources(t *testing.T) {
	cfg := &config.Config{
		Sinks: []config.SinkConfig{{Name: "out", Type: config.SinkTypeStdout}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing sources")
	}
}

func TestValidate_HTTPSinkMissingURL(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.SourceConfig{{Path: "/tmp/a.log"}},
		Sinks:   []config.SinkConfig{{Name: "remote", Type: config.SinkTypeHTTP}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for http sink without url")
	}
}

func TestValidate_UnknownSinkType(t *testing.T) {
	cfg := &config.Config{
		Sources: []config.SourceConfig{{Path: "/tmp/a.log"}},
		Sinks:   []config.SinkConfig{{Name: "x", Type: "kafka"}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unknown sink type")
	}
}
