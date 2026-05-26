package sink_test

import (
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func multilineEntry(msg string) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   msg,
		Fields:    map[string]interface{}{"message": msg},
	}
}

func TestNewMultilineSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewMultilineSink(nil, "message", []string{"\t"})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewMultilineSink_EmptyFieldReturnsError(t *testing.T) {
	inner := &captureSink{}
	_, err := sink.NewMultilineSink(inner, "", []string{"\t"})
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestNewMultilineSink_NoPrefixesReturnsError(t *testing.T) {
	inner := &captureSink{}
	_, err := sink.NewMultilineSink(inner, "message", nil)
	if err == nil {
		t.Fatal("expected error for empty prefixes")
	}
}

func TestMultilineSink_NonContinuationLinesForwardedSeparately(t *testing.T) {
	inner := &captureSink{}
	s, _ := sink.NewMultilineSink(inner, "message", []string{"\t"})

	_ = s.Write(multilineEntry("first"))
	_ = s.Write(multilineEntry("second"))
	_ = s.Flush()

	if len(inner.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(inner.entries))
	}
}

func TestMultilineSink_ContinuationLinesCoalesced(t *testing.T) {
	inner := &captureSink{}
	s, _ := sink.NewMultilineSink(inner, "message", []string{"\t", "at "})

	_ = s.Write(multilineEntry("exception occurred"))
	_ = s.Write(multilineEntry("\tat foo.Bar()"))
	_ = s.Write(multilineEntry("\tat foo.Baz()"))
	_ = s.Flush()

	if len(inner.entries) != 1 {
		t.Fatalf("expected 1 coalesced entry, got %d", len(inner.entries))
	}
	got, _ := inner.entries[0].Fields["message"].(string)
	const want = "exception occurred\n\tat foo.Bar()\n\tat foo.Baz()"
	if got != want {
		t.Errorf("coalesced message = %q; want %q", got, want)
	}
}

func TestMultilineSink_FlushOnNewNonContinuationLine(t *testing.T) {
	inner := &captureSink{}
	s, _ := sink.NewMultilineSink(inner, "message", []string{"\t"})

	_ = s.Write(multilineEntry("first"))
	_ = s.Write(multilineEntry("\tcontinuation"))
	// Writing a new non-continuation line should flush the previous group.
	_ = s.Write(multilineEntry("second"))

	if len(inner.entries) != 1 {
		t.Fatalf("expected 1 flushed entry before second, got %d", len(inner.entries))
	}
	_ = s.Flush()
	if len(inner.entries) != 2 {
		t.Fatalf("expected 2 total entries after flush, got %d", len(inner.entries))
	}
}

func TestMultilineSink_FlushEmptyIsNoop(t *testing.T) {
	inner := &captureSink{}
	s, _ := sink.NewMultilineSink(inner, "message", []string{"\t"})
	if err := s.Flush(); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if len(inner.entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(inner.entries))
	}
}
