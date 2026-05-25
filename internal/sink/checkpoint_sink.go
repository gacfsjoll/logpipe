package sink

import (
	"fmt"

	"logpipe/internal/checkpoint"
	"logpipe/internal/parser"
)

// CheckpointSink wraps an inner Sink and records the offset (line key) of
// every successfully forwarded entry into a Checkpoint store. On restart the
// pipeline can skip entries that were already delivered.
type CheckpointSink struct {
	inner      Sink
	cp         *checkpoint.Checkpoint
	sourceKey  string
}

// NewCheckpointSink creates a CheckpointSink.
// sourceKey is the logical name used as the checkpoint map key (e.g. file path).
func NewCheckpointSink(inner Sink, cp *checkpoint.Checkpoint, sourceKey string) (*CheckpointSink, error) {
	if inner == nil {
		return nil, fmt.Errorf("checkpoint sink: inner sink must not be nil")
	}
	if cp == nil {
		return nil, fmt.Errorf("checkpoint sink: checkpoint store must not be nil")
	}
	if sourceKey == "" {
		return nil, fmt.Errorf("checkpoint sink: sourceKey must not be empty")
	}
	return &CheckpointSink{inner: inner, cp: cp, sourceKey: sourceKey}, nil
}

// Write forwards the entry to the inner sink and, on success, persists the
// entry's timestamp (as a string offset) to the checkpoint store.
func (s *CheckpointSink) Write(entry parser.Entry) error {
	if err := s.inner.Write(entry); err != nil {
		return err
	}
	offset := entry.Timestamp.UTC().Format("2006-01-02T15:04:05.000000000Z")
	s.cp.Set(s.sourceKey, offset)
	return nil
}
