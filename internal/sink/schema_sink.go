package sink

import (
	"fmt"

	"logpipe/internal/parser"
	"logpipe/internal/schema"
)

// SchemaSink validates each log entry against a schema before forwarding it
// to an inner sink. Entries that fail validation are dropped and an error is
// returned to the caller.
type SchemaSink struct {
	inner    Sink
	validator *schema.Validator
}

// NewSchemaSink constructs a SchemaSink that enforces the supplied rules on
// every entry written to it. An error is returned when inner is nil or when
// the schema.Validator cannot be constructed from rules.
func NewSchemaSink(inner Sink, rules []schema.Rule) (*SchemaSink, error) {
	if inner == nil {
		return nil, fmt.Errorf("schema sink: inner sink must not be nil")
	}
	v, err := schema.New(rules)
	if err != nil {
		return nil, fmt.Errorf("schema sink: %w", err)
	}
	return &SchemaSink{inner: inner, validator: v}, nil
}

// Write validates entry against the configured schema. If validation passes
// the entry is forwarded to the inner sink unchanged. If validation fails the
// entry is dropped and the validation error is returned.
func (s *SchemaSink) Write(entry parser.LogEntry) error {
	if err := s.validator.Validate(entry); err != nil {
		return fmt.Errorf("schema sink: validation failed: %w", err)
	}
	return s.inner.Write(entry)
}
