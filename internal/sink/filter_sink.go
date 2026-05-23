package sink

import (
	"errors"
	"fmt"

	"logpipe/internal/filter"
	"logpipe/internal/parser"
)

// FilterSink wraps an inner Sink and only forwards log entries that pass
// all configured filter rules. Entries that do not match are silently dropped.
type FilterSink struct {
	inner  Sink
	filter *filter.Filter
}

// NewFilterSink constructs a FilterSink. Each rule in rules must specify a
// non-empty Field and at least one allowed Value.
func NewFilterSink(inner Sink, rules []filter.Rule) (*FilterSink, error) {
	if inner == nil {
		return nil, errors.New("filter sink: inner sink must not be nil")
	}
	f, err := filter.New(rules)
	if err != nil {
		return nil, fmt.Errorf("filter sink: %w", err)
	}
	return &FilterSink{inner: inner, filter: f}, nil
}

// Write forwards entry to the inner sink only when it passes all filter rules.
func (s *FilterSink) Write(entry parser.Entry) error {
	if !s.filter.Keep(entry) {
		return nil
	}
	return s.inner.Write(entry)
}
