package sink_test

import (
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func truncateEntry() parser.Entry {
	return parser.Entry{
		"message": "this is a fairly long log message that may need truncating",
		"level":   "info",
		"service": "auth",
	}
}

func TestNewTruncateSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewTruncateSink(nil, map[string]int{"message": 20})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewTruncateSink_EmptyRulesReturnsError(t *testing.T) {
	inner := sink.NewStdoutSink()
	_, err := sink.NewTruncateSink(inner, map[string]int{})
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNewTruncateSink_EmptyFieldNameReturnsError(t *testing.T) {
	inner := sink.NewStdoutSink()
	_, err := sink.NewTruncateSink(inner, map[string]int{"": 20})
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
}

func TestNewTruncateSink_MaxBytesTooSmallReturnsError(t *testing.T) {
	inner := sink.NewStdoutSink()
	_, err := sink.NewTruncateSink(inner, map[string]int{"message": 3})
	if err == nil {
		t.Fatal("expected error when max bytes <= 3")
	}
}

func TestTruncateSink_Write_TruncatesLongField(t *testing.T) {
	cap := &captureSink{}
	ts, err := sink.NewTruncateSink(cap, map[string]int{"message": 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := ts.Write(truncateEntry()); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	got, ok := cap.last["message"].(string)
	if !ok {
		t.Fatal("message field missing or not a string")
	}
	if len(got) != 10 {
		t.Errorf("expected length 10, got %d (%q)", len(got), got)
	}
	if got[7:] != "..." {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
}

func TestTruncateSink_Write_ShortFieldUnchanged(t *testing.T) {
	cap := &captureSink{}
	ts, err := sink.NewTruncateSink(cap, map[string]int{"message": 200})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	orig := truncateEntry()
	if err := ts.Write(orig); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	if cap.last["message"] != orig["message"] {
		t.Errorf("expected message unchanged, got %q", cap.last["message"])
	}
}

func TestTruncateSink_Write_DoesNotMutateOriginal(t *testing.T) {
	cap := &captureSink{}
	ts, _ := sink.NewTruncateSink(cap, map[string]int{"message": 10})

	orig := truncateEntry()
	origMsg := orig["message"].(string)
	_ = ts.Write(orig)

	if orig["message"] != origMsg {
		t.Error("original entry was mutated")
	}
}

func TestTruncateSink_Write_NonStringFieldIgnored(t *testing.T) {
	cap := &captureSink{}
	ts, _ := sink.NewTruncateSink(cap, map[string]int{"count": 10})

	entry := parser.Entry{"count": 42, "level": "info"}
	if err := ts.Write(entry); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if cap.last["count"] != 42 {
		t.Errorf("expected count unchanged, got %v", cap.last["count"])
	}
}
