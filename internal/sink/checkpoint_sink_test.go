package sink_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"logpipe/internal/checkpoint"
	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func cpEntry() parser.Entry {
	return parser.Entry{
		Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Level:     "info",
		Message:   "checkpoint test",
		Fields:    map[string]any{"svc": "api"},
	}
}

func newTestCheckpoint(t *testing.T) (*checkpoint.Checkpoint, string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cp.json")
	cp, err := checkpoint.New(p)
	if err != nil {
		t.Fatalf("checkpoint.New: %v", err)
	}
	return cp, p
}

func TestNewCheckpointSink_NilInnerReturnsError(t *testing.T) {
	cp, _ := newTestCheckpoint(t)
	_, err := sink.NewCheckpointSink(nil, cp, "src")
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewCheckpointSink_NilCheckpointReturnsError(t *testing.T) {
	_, err := sink.NewCheckpointSink(sink.NewStdoutSinkWithWriter(os.Discard), nil, "src")
	if err == nil {
		t.Fatal("expected error for nil checkpoint")
	}
}

func TestNewCheckpointSink_EmptySourceKeyReturnsError(t *testing.T) {
	cp, _ := newTestCheckpoint(t)
	_, err := sink.NewCheckpointSink(sink.NewStdoutSinkWithWriter(os.Discard), cp, "")
	if err == nil {
		t.Fatal("expected error for empty source key")
	}
}

func TestCheckpointSink_Write_ForwardsToInner(t *testing.T) {
	cp, _ := newTestCheckpoint(t)
	inner := sink.NewStdoutSinkWithWriter(os.Discard)
	s, err := sink.NewCheckpointSink(inner, cp, "/var/log/app.log")
	if err != nil {
		t.Fatalf("NewCheckpointSink: %v", err)
	}
	if err := s.Write(cpEntry()); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestCheckpointSink_Write_PersistsOffset(t *testing.T) {
	cp, cpPath := newTestCheckpoint(t)
	inner := sink.NewStdoutSinkWithWriter(os.Discard)
	s, _ := sink.NewCheckpointSink(inner, cp, "/var/log/app.log")

	e := cpEntry()
	_ = s.Write(e)

	if err := cp.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	cp2, _ := checkpoint.New(cpPath)
	offset, ok := cp2.Get("/var/log/app.log")
	if !ok {
		t.Fatal("expected offset to be persisted")
	}
	want := "2024-06-01T12:00:00.000000000Z"
	if offset != want {
		t.Fatalf("offset = %q, want %q", offset, want)
	}
}

func TestCheckpointSink_Write_InnerErrorSkipsCheckpoint(t *testing.T) {
	cp, _ := newTestCheckpoint(t)
	failing := &alwaysFailSink{err: errors.New("downstream unavailable")}
	s, _ := sink.NewCheckpointSink(failing, cp, "src")

	_ = s.Write(cpEntry())

	_, ok := cp.Get("src")
	if ok {
		t.Fatal("checkpoint should not be updated when inner sink fails")
	}
}

// alwaysFailSink is a minimal Sink that always returns an error.
type alwaysFailSink struct{ err error }

func (f *alwaysFailSink) Write(_ parser.Entry) error { return f.err }
