package sink_test

import (
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func levelEntry(level string) parser.Entry {
	return parser.Entry{"level": level, "msg": "hello"}
}

func TestNewLevelFilterSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewLevelFilterSink(nil, "info")
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewLevelFilterSink_EmptyLevelReturnsError(t *testing.T) {
	_, err := sink.NewLevelFilterSink(&sink.StdoutSink{}, "")
	if err == nil {
		t.Fatal("expected error for empty level")
	}
}

func TestNewLevelFilterSink_UnknownLevelReturnsError(t *testing.T) {
	_, err := sink.NewLevelFilterSink(&sink.StdoutSink{}, "trace")
	if err == nil {
		t.Fatal("expected error for unknown level")
	}
}

func TestLevelFilterSink_Write_PassesAtOrAboveMinLevel(t *testing.T) {
	cap := &captureSink{}
	s, err := sink.NewLevelFilterSink(cap, "warn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, lvl := range []string{"warn", "error", "fatal"} {
		if err := s.Write(levelEntry(lvl)); err != nil {
			t.Errorf("Write(%q) returned error: %v", lvl, err)
		}
	}
	if got := len(cap.entries); got != 3 {
		t.Fatalf("expected 3 forwarded entries, got %d", got)
	}
}

func TestLevelFilterSink_Write_DropsBelow(t *testing.T) {
	cap := &captureSink{}
	s, err := sink.NewLevelFilterSink(cap, "error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, lvl := range []string{"debug", "info", "warn"} {
		if err := s.Write(levelEntry(lvl)); err != nil {
			t.Errorf("Write(%q) returned unexpected error: %v", lvl, err)
		}
	}
	if got := len(cap.entries); got != 0 {
		t.Fatalf("expected 0 forwarded entries, got %d", got)
	}
}

func TestLevelFilterSink_Write_DropsEntryWithNoLevelField(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewLevelFilterSink(cap, "info")
	_ = s.Write(parser.Entry{"msg": "no level here"})
	if len(cap.entries) != 0 {
		t.Fatalf("expected entry to be dropped, got %d forwarded", len(cap.entries))
	}
}

func TestLevelFilterSink_Write_DropsEntryWithUnknownLevel(t *testing.T) {
	cap := &captureSink{}
	s, _ := sink.NewLevelFilterSink(cap, "info")
	_ = s.Write(parser.Entry{"level": "verbose", "msg": "unknown level"})
	if len(cap.entries) != 0 {
		t.Fatalf("expected entry to be dropped, got %d forwarded", len(cap.entries))
	}
}
