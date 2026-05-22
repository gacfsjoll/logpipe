package envelope_test

import (
	"testing"
	"time"

	"logpipe/internal/envelope"
	"logpipe/internal/parser"
)

func makeEntry(msg string) parser.Entry {
	return parser.Entry{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Message:   msg,
		Fields:    map[string]any{"app": "logpipe"},
	}
}

func TestWrap_AssignsMonotonicSequence(t *testing.T) {
	envelope.ResetSequence()

	e1 := envelope.Wrap("/var/log/a.log", makeEntry("first"))
	e2 := envelope.Wrap("/var/log/a.log", makeEntry("second"))

	if e1.Sequence != 1 {
		t.Fatalf("expected sequence 1, got %d", e1.Sequence)
	}
	if e2.Sequence != 2 {
		t.Fatalf("expected sequence 2, got %d", e2.Sequence)
	}
}

func TestWrap_SequencesAreStrictlyIncreasing(t *testing.T) {
	envelope.ResetSequence()

	const n = 50
	envs := make([]envelope.Envelope, n)
	for i := range envs {
		envs[i] = envelope.Wrap("/log", makeEntry("msg"))
	}
	for i := 1; i < n; i++ {
		if envs[i].Sequence <= envs[i-1].Sequence {
			t.Fatalf("sequence not strictly increasing at index %d: %d <= %d",
				i, envs[i].Sequence, envs[i-1].Sequence)
		}
	}
}

func TestWrap_SourcePreserved(t *testing.T) {
	envelope.ResetSequence()
	src := "/var/log/app.log"
	e := envelope.Wrap(src, makeEntry("hello"))
	if e.Source != src {
		t.Fatalf("expected source %q, got %q", src, e.Source)
	}
}

func TestWrap_EntryPreserved(t *testing.T) {
	envelope.ResetSequence()
	entry := makeEntry("preserved")
	e := envelope.Wrap("/log", entry)
	if e.Entry.Message != entry.Message {
		t.Fatalf("expected message %q, got %q", entry.Message, e.Entry.Message)
	}
}

func TestWrap_ReceivedAtIsUTC(t *testing.T) {
	envelope.ResetSequence()
	before := time.Now().UTC()
	e := envelope.Wrap("/log", makeEntry("ts check"))
	after := time.Now().UTC()

	if e.ReceivedAt.Before(before) || e.ReceivedAt.After(after) {
		t.Fatalf("ReceivedAt %v not in expected range [%v, %v]",
			e.ReceivedAt, before, after)
	}
	if e.ReceivedAt.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %v", e.ReceivedAt.Location())
	}
}
