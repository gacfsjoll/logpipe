package parser

import (
	"testing"
	"time"
)

func TestParse_ValidFullEntry(t *testing.T) {
	p := New("app")
	line := `{"timestamp":"2024-01-15T10:00:00Z","level":"info","message":"started","pid":42}`

	entry, err := p.Parse(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Source != "app" {
		t.Errorf("expected source 'app', got %q", entry.Source)
	}
	if entry.Level != "info" {
		t.Errorf("expected level 'info', got %q", entry.Level)
	}
	if entry.Message != "started" {
		t.Errorf("expected message 'started', got %q", entry.Message)
	}
	if _, ok := entry.Fields["pid"]; !ok {
		t.Error("expected 'pid' in fields")
	}
	expected := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	if !entry.Timestamp.Equal(expected) {
		t.Errorf("expected timestamp %v, got %v", expected, entry.Timestamp)
	}
}

func TestParse_AlternativeKeys(t *testing.T) {
	p := New("svc")
	line := `{"ts":"2024-06-01T12:00:00Z","severity":"error","msg":"oops"}`

	entry, err := p.Parse(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Level != "error" {
		t.Errorf("expected level 'error', got %q", entry.Level)
	}
	if entry.Message != "oops" {
		t.Errorf("expected message 'oops', got %q", entry.Message)
	}
}

func TestParse_EmptyLine(t *testing.T) {
	p := New("x")
	_, err := p.Parse("")
	if err == nil {
		t.Error("expected error for empty line")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	p := New("x")
	_, err := p.Parse("not json at all")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParse_MissingTimestampDefaultsToNow(t *testing.T) {
	p := New("x")
	before := time.Now().UTC()
	entry, err := p.Parse(`{"message":"hello"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now().UTC()
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Errorf("timestamp %v not within expected range", entry.Timestamp)
	}
}
