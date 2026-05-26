package sink_test

import (
	"errors"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func timestampEntry(ts time.Time) parser.Entry {
	return parser.Entry{
		Timestamp: ts,
		Level:     "info",
		Message:   "hello",
		Fields:    map[string]any{"service": "test"},
	}
}

func TestNewTimestampSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewTimestampSink(nil, sink.TimestampConfig{})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestTimestampSink_Write_DefaultFieldName(t *testing.T) {
	cap := &captureSink{}
	s, err := sink.NewTimestampSink(cap, sink.TimestampConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if err := s.Write(timestampEntry(now)); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	got, ok := cap.last.Fields["timestamp"]
	if !ok {
		t.Fatal("expected 'timestamp' field to be set")
	}
	want := now.UTC().Format(time.RFC3339Nano)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTimestampSink_Write_CustomFieldName(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewTimestampSink(cap, sink.TimestampConfig{FieldName: "@ts"})

	now := time.Now().UTC()
	_ = s.Write(timestampEntry(now))

	if _, ok := cap.last.Fields["@ts"]; !ok {
		t.Error("expected '@ts' field to be set")
	}
	if _, ok := cap.last.Fields["timestamp"]; ok {
		t.Error("unexpected default 'timestamp' field")
	}
}

func TestTimestampSink_Write_RespectsLocation(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	cap := &captureSink{}
	s, _ := sink.NewTimestampSink(cap, sink.TimestampConfig{Location: loc})

	now := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	_ = s.Write(timestampEntry(now))

	got := cap.last.Fields["timestamp"].(string)
	want := now.In(loc).Format(time.RFC3339Nano)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTimestampSink_Write_DoesNotMutateOriginal(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewTimestampSink(cap, sink.TimestampConfig{})

	entry := timestampEntry(time.Now().UTC())
	delete(entry.Fields, "timestamp")
	_ = s.Write(entry)

	if _, ok := entry.Fields["timestamp"]; ok {
		t.Error("original entry should not have been mutated")
	}
}

func TestTimestampSink_Write_PropagatesInnerError(t *testing.T) {
	errSink := &errorSink{err: errors.New("downstream failure")}
	s, _ := sink.NewTimestampSink(errSink, sink.TimestampConfig{})

	err := s.Write(timestampEntry(time.Now().UTC()))
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
}
