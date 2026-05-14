package sink_test

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func sampleFileEntry() parser.Entry {
	return parser.Entry{
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Level:     "info",
		Message:   "file sink test",
		Fields:    map[string]any{"service": "test"},
	}
}

func TestFileSink_Write_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/out.log"

	s, err := sink.NewFileSink(path)
	if err != nil {
		t.Fatalf("NewFileSink: %v", err)
	}
	defer s.Close()

	if err := s.Write(sampleFileEntry()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist after write")
	}
}

func TestFileSink_Write_ValidJSON(t *testing.T) {
	tmp := t.TempDir()
	s, err := sink.NewFileSink(tmp + "/out.log")
	if err != nil {
		t.Fatalf("NewFileSink: %v", err)
	}
	defer s.Close()

	entry := sampleFileEntry()
	if err := s.Write(entry); err != nil {
		t.Fatalf("Write: %v", err)
	}
	s.Close()

	f, _ := os.Open(s.Path())
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatal("expected at least one line")
	}
	var got map[string]any
	if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
		t.Fatalf("line is not valid JSON: %v", err)
	}
	if got["message"] != "file sink test" {
		t.Errorf("unexpected message: %v", got["message"])
	}
}

func TestFileSink_Write_MultipleEntries(t *testing.T) {
	tmp := t.TempDir()
	s, err := sink.NewFileSink(tmp + "/multi.log")
	if err != nil {
		t.Fatalf("NewFileSink: %v", err)
	}
	defer s.Close()

	for i := 0; i < 5; i++ {
		if err := s.Write(sampleFileEntry()); err != nil {
			t.Fatalf("Write[%d]: %v", i, err)
		}
	}
	s.Close()

	f, _ := os.Open(s.Path())
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 5 {
		t.Errorf("expected 5 lines, got %d", count)
	}
}

func TestFileSink_InvalidPath(t *testing.T) {
	_, err := sink.NewFileSink("/nonexistent/dir/out.log")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}
