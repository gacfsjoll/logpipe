package sink

import (
	"fmt"

	"logpipe/internal/parser"
	"logpipe/internal/schema"
)

// SchemaSink validates each log entry against a set of schema rules before
// forwarding to an inner sink. Entries that fail validation are dropped and
// an error is returned to the caller.
type schemaValidateSink struct {
	inner    Sink
	validator *schema.Validator
}

// NewSchemaSink constructs a sink that validates entries against the supplied
// rules. It returns an error if inner is nil or rules is empty.
func NewSchemaSink(inner Sink, rules []schema.Rule) (*schemaValidateSink, error) {
	if inner == nil {
		return nil, fmt.Errorf("schema sink: inner sink must not be nil")
	}
	v, err := schema.New(rules)
	if err != nil {
		return nil, fmt.Errorf("schema sink: %w", err)
	}
	return &schemaValidateSink{inner: inner, validator: v}, nil
}

// Write validates the entry. If validation passes the entry is forwarded to
// the inner sink. If validation fails the entry is dropped and the validation
// error is returned.
func (s *schemaValidateSink) Write(entry parser.Entry) error {
	if err := s.validator.Validate(entry); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return s.inner.Write(entry)
}
