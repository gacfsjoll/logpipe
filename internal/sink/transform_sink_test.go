package sink_test

import (
	"errors"
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
	"logpipe/internal/transform"
)

type captureTransformSink struct {
	wrote []parser.Entry
	errOn int // return error on this call (1-based); 0 = never
	calls int
}

func (c *captureTransformSink) Write(e parser.Entry) error {
	c.calls++
	if c.errOn > 0 && c.calls == c.errOn {
		return errors.New("injected error")
	}
	c.wrote = append(c.wrote, e)
	return nil
}

func makeTransformEntry() parser.Entry {
	return parser.Entry{
		"level":   "info",
		"message": "hello",
	}
}

func TestNewTransformSink_NilInnerReturnsError(t *testing.T) {
	t.Parallel()
	tr := transform.New()
	_, err := sink.NewTransformSink(nil, tr)
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewTransformSink_NilTransformerReturnsError(t *testing.T) {
	t.Parallel()
	cap := &captureTransformSink{}
	_, err := sink.NewTransformSink(cap, nil)
	if err == nil {
		t.Fatal("expected error for nil transformer")
	}
}

func TestTransformSink_Write_ForwardsTransformedEntry(t *testing.T) {
	t.Parallel()
	tr := transform.New(transform.AddField("env", "test"))
	cap := &captureTransformSink{}
	s, err := sink.NewTransformSink(cap, tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := makeTransformEntry()
	if err := s.Write(entry); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	if len(cap.wrote) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cap.wrote))
	}
	if cap.wrote[0]["env"] != "test" {
		t.Errorf("expected env=test, got %v", cap.wrote[0]["env"])
	}
}

func TestTransformSink_Write_DoesNotMutateOriginal(t *testing.T) {
	t.Parallel()
	tr := transform.New(transform.AddField("injected", "yes"))
	cap := &captureTransformSink{}
	s, _ := sink.NewTransformSink(cap, tr)

	entry := makeTransformEntry()
	_ = s.Write(entry)

	if _, ok := entry["injected"]; ok {
		t.Error("original entry was mutated by TransformSink")
	}
}

func TestTransformSink_Write_PropagatesInnerError(t *testing.T) {
	t.Parallel()
	tr := transform.New()
	cap := &captureTransformSink{errOn: 1}
	s, _ := sink.NewTransformSink(cap, tr)

	if err := s.Write(makeTransformEntry()); err == nil {
		t.Error("expected error from inner sink to propagate")
	}
}
