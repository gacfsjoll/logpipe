package sink_test

import (
	"errors"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func bufEntry(msg string) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   msg,
		Fields:    map[string]any{},
	}
}

func TestBufferedSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewBufferedSink(nil, 10)
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestBufferedSink_ZeroCapacityReturnsError(t *testing.T) {
	stdout := sink.NewStdoutSinkWithWriter(nil)
	_, err := sink.NewBufferedSink(stdout, 0)
	if err == nil {
		t.Fatal("expected error for zero capacity")
	}
}

func TestBufferedSink_WriteDoesNotForwardImmediately(t *testing.T) {
	var received []parser.Entry
	inner := &collectSink{collect: &received}

	bs, err := sink.NewBufferedSink(inner, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = bs.Write(bufEntry("hello"))
	if len(received) != 0 {
		t.Fatalf("expected 0 forwarded entries, got %d", len(received))
	}
	if bs.Len() != 1 {
		t.Fatalf("expected buffer len 1, got %d", bs.Len())
	}
}

func TestBufferedSink_FlushForwardsAll(t *testing.T) {
	var received []parser.Entry
	inner := &collectSink{collect: &received}

	bs, _ := sink.NewBufferedSink(inner, 5)
	for i := 0; i < 3; i++ {
		_ = bs.Write(bufEntry("msg"))
	}
	if err := bs.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}
	if len(received) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(received))
	}
	if bs.Len() != 0 {
		t.Fatalf("expected empty buffer after flush, got %d", bs.Len())
	}
}

func TestBufferedSink_OverflowDropsOldest(t *testing.T) {
	var received []parser.Entry
	inner := &collectSink{collect: &received}

	bs, _ := sink.NewBufferedSink(inner, 2)
	_ = bs.Write(bufEntry("first"))
	_ = bs.Write(bufEntry("second"))
	_ = bs.Write(bufEntry("third")) // should evict "first"

	_ = bs.Flush()
	if len(received) != 2 {
		t.Fatalf("expected 2 entries after overflow, got %d", len(received))
	}
	if received[0].Message != "second" {
		t.Errorf("expected oldest surviving entry to be 'second', got %q", received[0].Message)
	}
}

func TestBufferedSink_FlushPropagatesInnerError(t *testing.T) {
	inner := &errSink{err: errors.New("downstream failure")}
	bs, _ := sink.NewBufferedSink(inner, 5)
	_ = bs.Write(bufEntry("x"))
	if err := bs.Flush(); err == nil {
		t.Fatal("expected flush to return inner sink error")
	}
}

// collectSink accumulates written entries for assertion.
type collectSink struct{ collect *[]parser.Entry }

func (c *collectSink) Write(e parser.Entry) error {
	*c.collect = append(*c.collect, e)
	return nil
}

// errSink always returns an error on Write.
type errSink struct{ err error }

func (e *errSink) Write(_ parser.Entry) error { return e.err }
