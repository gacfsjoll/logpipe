package sink

import (
	"fmt"

	"logpipe/internal/parser"
)

// TruncateSink wraps an inner Sink and truncates string field values that
// exceed a configured maximum byte length. Truncated values are suffixed
// with "..." to signal the loss of data.
type TruncateSink struct {
	inner  Sink
	rules  map[string]int // field -> max bytes
}

// NewTruncateSink returns a TruncateSink that applies the given field
// truncation rules before forwarding each entry to inner.
//
// rules must be non-empty; every field name must be non-empty and every
// max-bytes value must be greater than three (to leave room for the
// ellipsis suffix).
func NewTruncateSink(inner Sink, rules map[string]int) (*TruncateSink, error) {
	if inner == nil {
		return nil, fmt.Errorf("truncate sink: inner sink must not be nil")
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("truncate sink: rules must not be empty")
	}
	for field, max := range rules {
		if field == "" {
			return nil, fmt.Errorf("truncate sink: field name must not be empty")
		}
		if max <= 3 {
			return nil, fmt.Errorf("truncate sink: max bytes for field %q must be greater than 3", field)
		}
	}
	return &TruncateSink{inner: inner, rules: rules}, nil
}

// Write applies truncation rules to a copy of entry, then forwards the
// modified copy to the inner sink.
func (t *TruncateSink) Write(entry parser.Entry) error {
	copy := make(parser.Entry, len(entry))
	for k, v := range entry {
		copy[k] = v
	}
	for field, max := range t.rules {
		v, ok := copy[field]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		if len(s) > max {
			copy[field] = s[:max-3] + "..."
		}
	}
	return t.inner.Write(copy)
}
