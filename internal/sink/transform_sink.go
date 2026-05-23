package sink

import (
	"errors"

	"logpipe/internal/parser"
	"logpipe/internal/transform"
)

// TransformSink wraps an inner Sink and applies a Transformer to each log
// entry before forwarding it downstream. The original entry is never mutated.
type TransformSink struct {
	inner     Sink
	transformer *transform.Transformer
}

// NewTransformSink returns a TransformSink that applies t to every entry
// before writing to inner. Both arguments must be non-nil.
func NewTransformSink(inner Sink, t *transform.Transformer) (*TransformSink, error) {
	if inner == nil {
		return nil, errors.New("transform sink: inner sink must not be nil")
	}
	if t == nil {
		return nil, errors.New("transform sink: transformer must not be nil")
	}
	return &TransformSink{inner: inner, transformer: t}, nil
}

// Write transforms entry and forwards the result to the inner sink.
func (s *TransformSink) Write(entry parser.Entry) error {
	transformed := s.transformer.Apply(entry)
	return s.inner.Write(transformed)
}
