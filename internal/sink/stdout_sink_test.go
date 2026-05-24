package sink

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"logpipe/internal/parser"
)

func TestStdoutSink_Write_ProducesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	s := NewStdoutSinkWithWriter(&buf)

	entry := parser.Entry{
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Level:     "info",
		Message:   "service started",
		Fields:    map[string]interface{}{"pid": float64(42)},
	}

	if err := s.Write(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected output, got empty string")
	}

	var got parser.Entry
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v — output: %s", err, line)
	}

	if got.Level != "info" {
		t.Errorf("expected level 'info', got %q", got.Level)
	}
	if got.Message != "service started" {
		t.Errorf("expected message 'service started', got %q", got.Message)
	}
}

func TestStdoutSink_Write_AppendsNewline(t *testing.T) {
	var buf bytes.Buffer
	s := NewStdoutSinkWithWriter(&buf)

	entry := sampleEntry()
	if err := s.Write(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestStdoutSink_Write_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	s := NewStdoutSinkWithWriter(&buf)

	for i := 0; i < 3; i++ {
		if err := s.Write(sampleEntry()); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestStdoutSink_Write_PreservesFields(t *testing.T) {
	var buf bytes.Buffer
	s := NewStdoutSinkWithWriter(&buf)

	entry := parser.Entry{
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Level:     "debug",
		Message:   "request handled",
		Fields:    map[string]interface{}{"method": "GET", "status": float64(200)},
	}

	if err := s.Write(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got parser.Entry
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if got.Fields["method"] != "GET" {
		t.Errorf("expected field method='GET', got %v", got.Fields["method"])
	}
	if got.Fields["status"] != float64(200) {
		t.Errorf("expected field status=200, got %v", got.Fields["status"])
	}
}
