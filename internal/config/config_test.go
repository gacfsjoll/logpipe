package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"logpipe/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatalf("createTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTemp(t, `
sources:
  - path: /var/log/app.log
sinks:
  - type: stdout
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Sources) != 1 || cfg.Sources[0].Path != "/var/log/app.log" {
		t.Errorf("unexpected sources: %+v", cfg.Sources)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidate_NoSources(t *testing.T) {
	path := writeTemp(t, `sinks:\n  - type: stdout\n`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidate_HTTPSinkMissingURL(t *testing.T) {
	path := writeTemp(t, `
sources:
  - path: /tmp/a.log
sinks:
  - type: http
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error: http sink missing url")
	}
}

func TestValidate_FileSinkMissingPath(t *testing.T) {
	path := writeTemp(t, `
sources:
  - path: /tmp/a.log
sinks:
  - type: file
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error: file sink missing file_path")
	}
}

func TestValidate_FileSinkValid(t *testing.T) {
	tmp := t.TempDir()
	path := writeTemp(t, "sources:\n  - path: /tmp/a.log\nsinks:\n  - type: file\n    file_path: "+tmp+"/out.log\n")
	_, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_UnknownSinkType(t *testing.T) {
	path := writeTemp(t, `
sources:
  - path: /tmp/a.log
sinks:
  - type: kafka
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for unknown sink type")
	}
}
