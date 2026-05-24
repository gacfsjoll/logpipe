package sink

import (
	"context"
	"fmt"

	"logpipe/internal/batch"
	"logpipe/internal/parser"
)

// BatchedSink wraps an inner Sink and accumulates entries, flushing them
// as a group when the batch is full or the flush interval elapses.
//
// Flush is performed by writing each entry in the batch to the inner sink
// in order. The caller must invoke Close (or cancel the context passed to
// Start) to drain the final partial batch before shutdown.
type BatchedSink struct {
	inner  Sink
	batcher *batch.Batcher
}

// BatchedSinkConfig holds tuning parameters for BatchedSink.
type BatchedSinkConfig = batch.Config

// NewBatchedSink creates a BatchedSink that groups writes to inner.
// cfg.MaxSize and cfg.FlushInterval control when a batch is flushed.
func NewBatchedSink(inner Sink, cfg BatchedSinkConfig) (*BatchedSink, error) {
	if inner == nil {
		return nil, fmt.Errorf("batch sink: inner sink must not be nil")
	}

	flushFn := func(entries []parser.Entry) error {
		for _, e := range entries {
			if err := inner.Write(e); err != nil {
				return err
			}
		}
		return nil
	}

	b, err := batch.New(cfg, flushFn)
	if err != nil {
		return nil, fmt.Errorf("batch sink: %w", err)
	}

	return &BatchedSink{inner: inner, batcher: b}, nil
}

// Start begins the background flush timer. It blocks until ctx is cancelled.
func (s *BatchedSink) Start(ctx context.Context) {
	s.batcher.Start(ctx)
}

// Write adds entry to the current batch. If the batch reaches MaxSize it is
// flushed synchronously before Write returns.
func (s *BatchedSink) Write(e parser.Entry) error {
	return s.batcher.Add(e)
}
