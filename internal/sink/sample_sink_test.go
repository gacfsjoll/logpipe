package sink_test

import (
	"testing"

	"logpipe/internal/parser"
	"logpipe/internal/sink"
)

func sampleSinkEntry() parser.Entry {
	return parser.Entry{
		"level":   "info",
		"message": "hello",
		"service": "test",
	}
}

func TestNewSampledSink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewSampledSink(nil, sink.SamplerConfig{Rate: 0.5})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewSampledSink_InvalidRateReturnsError(t *testing.T) {
	w := &stdoutWriter{}
	inner := sink.NewStdoutSinkWithWriter(w)
	_, err := sink.NewSampledSink(inner, sink.SamplerConfig{Rate: 1.5})
	if err == nil {
		t.Fatal("expected error for rate > 1")
	}
}

func TestSampledSink_RateOne_ForwardsAll(t *testing.T) {
	collector := &collectSink{}
	s, err := sink.NewSampledSink(collector, sink.SamplerConfig{Rate: 1.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const n = 20
	for i := 0; i < n; i++ {
		if err := s.Write(sampleSinkEntry()); err != nil {
			t.Fatalf("Write error: %v", err)
		}
	}
	if len(collector.entries) != n {
		t.Errorf("expected %d entries forwarded, got %d", n, len(collector.entries))
	}
}

func TestSampledSink_RateZeroPoint1_DropsEntries(t *testing.T) {
	collector := &collectSink{}
	// Use a fixed seed for determinism.
	s, err := sink.NewSampledSink(collector, sink.SamplerConfig{Rate: 0.1, Seed: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const n = 1000
	for i := 0; i < n; i++ {
		_ = s.Write(sampleSinkEntry())
	}
	if len(collector.entries) == n {
		t.Error("expected some entries to be dropped at rate 0.1")
	}
	if len(collector.entries) == 0 {
		t.Error("expected some entries to be forwarded at rate 0.1")
	}
}

func TestSampledSink_InnerErrorPropagated(t *testing.T) {
	inner := &errorSink{}
	s, err := sink.NewSampledSink(inner, sink.SamplerConfig{Rate: 1.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Write(sampleSinkEntry()); err == nil {
		t.Error("expected error from inner sink to propagate")
	}
}

// collectSink records every entry written to it.
type collectSink struct {
	entries []parser.Entry
}

func (c *collectSink) Write(e parser.Entry) error {
	c.entries = append(c.entries, e)
	return nil
}
