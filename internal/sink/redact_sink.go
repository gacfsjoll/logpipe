package sink

import (
	"errors"

	"logpipe/internal/parser"
	"logpipe/internal/redact"
)

// RedactSink wraps an inner Sink and applies field redaction rules to each
// log entry before forwarding it downstream. The original entry is never
// mutated; redact.Apply returns a shallow copy with the affected fields
// replaced by the configured mask string.
type RedactSink struct {
	inner   Sink
	redactor *redact.Redactor
}

// NewRedactSink constructs a RedactSink.
//
// rules is a map of field name → mask string (e.g. {"password": "[REDACTED]"})
// An empty rules map is rejected because a no-op wrapper adds overhead with no
// benefit; callers should omit the sink in that case.
func NewRedactSink(inner Sink, rules map[string]string) (*RedactSink, error) {
	if inner == nil {
		return nil, errors.New("redact sink: inner sink must not be nil")
	}
	r, err := redact.New(rules)
	if err != nil {
		return nil, err
	}
	return &RedactSink{inner: inner, redactor: r}, nil
}

// Write applies the redaction rules to entry and forwards the result to the
// inner sink. Errors from the inner sink are propagated unchanged.
func (s *RedactSink) Write(entry parser.Entry) error {
	return s.inner.Write(s.redactor.Apply(entry))
}
