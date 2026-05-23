package sink

import (
	"errors"

	"logpipe/internal/envelope"
	"logpipe/internal/parser"
)

// EnvelopeSink wraps each log entry in an envelope (adding sequence number
// and source metadata) before forwarding it to an inner Sink.
type EnvelopeSink struct {
	inner  Sink
	source string
}

// NewEnvelopeSink constructs an EnvelopeSink that stamps every entry with
// a monotonically increasing sequence number and the given source label
// before delegating to inner.
func NewEnvelopeSink(inner Sink, source string) (*EnvelopeSink, error) {
	if inner == nil {
		return nil, errors.New("envelope sink: inner sink must not be nil")
	}
	if source == "" {
		return nil, errors.New("envelope sink: source must not be empty")
	}
	return &EnvelopeSink{inner: inner, source: source}, nil
}

// Write wraps entry in an envelope then forwards the wrapped entry to the
// inner sink. The original entry is never mutated.
func (e *EnvelopeSink) Write(entry parser.Entry) error {
	wrapped := envelope.Wrap(entry, e.source)
	return e.inner.Write(wrapped)
}
