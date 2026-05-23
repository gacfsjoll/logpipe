package sink_test

import (
	"errors"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func envelopeEntry() parser.Entry {
	return parser.Entry{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Message:   "hello envelope",
		Fields:    map[string]any{"app": "logpipe"},
	}
}

func TestNewEnvelopeSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewEnvelopeSink(nil, "svc")
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewEnvelopeSink_EmptySourceReturnsError(t *testing.T) {
	_, err := sink.NewEnvelopeSink(&sink.StdoutSink{}, "")
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestEnvelopeSink_Write_ForwardsToInner(t *testing.T) {
	cap := &captureSink{}
	es, err := sink.NewEnvelopeSink(cap, "test-source")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := envelopeEntry()
	if err := es.Write(entry); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	if len(cap.entries) != 1 {
		t.Fatalf("expected 1 entry forwarded, got %d", len(cap.entries))
	}
}

func TestEnvelopeSink_Write_SequenceIncreases(t *testing.T) {
	cap := &captureSink{}
	es, _ := sink.NewEnvelopeSink(cap, "svc")

	for i := 0; i < 3; i++ {
		_ = es.Write(envelopeEntry())
	}

	if len(cap.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(cap.entries))
	}

	seqs := make([]int64, 3)
	for i, e := range cap.entries {
		v, ok := e.Fields["seq"]
		if !ok {
			t.Fatalf("entry %d missing seq field", i)
		}
		seqs[i] = v.(int64)
	}
	for i := 1; i < len(seqs); i++ {
		if seqs[i] <= seqs[i-1] {
			t.Errorf("sequence not strictly increasing: %v", seqs)
		}
	}
}

func TestEnvelopeSink_Write_DoesNotMutateOriginal(t *testing.T) {
	cap := &captureSink{}
	es, _ := sink.NewEnvelopeSink(cap, "svc")

	entry := envelopeEntry()
	origFields := len(entry.Fields)
	_ = es.Write(entry)

	if len(entry.Fields) != origFields {
		t.Error("Write mutated the original entry's Fields map")
	}
}

func TestEnvelopeSink_Write_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	es, _ := sink.NewEnvelopeSink(&errorSink{err: sentinel}, "svc")

	if err := es.Write(envelopeEntry()); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
