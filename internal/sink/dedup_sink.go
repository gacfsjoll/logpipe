package sink

import (
	"errors"
	"fmt"

	"logpipe/internal/dedup"
	"logpipe/internal/parser"
)

// DedupSink wraps an inner Sink and drops duplicate log entries within a
// configurable time window. Deduplication is keyed on the raw log line.
type DedupSink struct {
	inner  Sink
	filter *dedup.Dedup
}

// DedupConfig holds configuration for the deduplication sink wrapper.
type DedupConfig struct {
	// WindowSeconds is the duration in seconds during which identical lines are
	// considered duplicates. Must be greater than zero.
	WindowSeconds int
	// MaxEntries is the maximum number of unique entries tracked at once.
	// Must be greater than zero.
	MaxEntries int
}

// NewDedupSink returns a Sink that forwards unique entries to inner, silently
// dropping any entry whose raw line has already been seen within the window.
func NewDedupSink(inner Sink, cfg DedupConfig) (*DedupSink, error) {
	if inner == nil {
		return nil, errors.New("dedup sink: inner sink must not be nil")
	}
	if cfg.WindowSeconds <= 0 {
		return nil, fmt.Errorf("dedup sink: WindowSeconds must be > 0, got %d", cfg.WindowSeconds)
	}
	if cfg.MaxEntries <= 0 {
		return nil, fmt.Errorf("dedup sink: MaxEntries must be > 0, got %d", cfg.MaxEntries)
	}

	d, err := dedup.New(dedup.Config{
		WindowSeconds: cfg.WindowSeconds,
		MaxEntries:    cfg.MaxEntries,
	})
	if err != nil {
		return nil, fmt.Errorf("dedup sink: %w", err)
	}

	return &DedupSink{inner: inner, filter: d}, nil
}

// Write forwards entry to the inner sink only if it has not been seen within
// the configured deduplication window.
func (d *DedupSink) Write(entry parser.Entry) error {
	if d.filter.IsDuplicate(entry.Raw) {
		return nil
	}
	return d.inner.Write(entry)
}
