package sink

import (
	"context"
	"errors"
	"fmt"

	"logpipe/internal/backpressure"
	"logpipe/internal/parser"
)

// BackpressureSink wraps an inner Sink and applies a concurrency limit so that
// at most Limit writes are in-flight simultaneously. Callers block until a
// token is available or the context is cancelled.
type BackpressureSink struct {
	inner  Sink
	valve  *backpressure.Valve
}

// BackpressureConfig controls the behaviour of NewBackpressureSink.
type BackpressureConfig struct {
	// Limit is the maximum number of concurrent in-flight writes. Must be >= 1.
	Limit int
}

// NewBackpressureSink creates a BackpressureSink that wraps inner.
func NewBackpressureSink(inner Sink, cfg BackpressureConfig) (*BackpressureSink, error) {
	if inner == nil {
		return nil, errors.New("backpressure_sink: inner sink must not be nil")
	}
	if cfg.Limit < 1 {
		return nil, fmt.Errorf("backpressure_sink: limit must be >= 1, got %d", cfg.Limit)
	}
	v := backpressure.New(cfg.Limit)
	return &BackpressureSink{inner: inner, valve: v}, nil
}

// Write acquires a token from the valve, delegates to the inner sink, then
// releases the token. If the context is cancelled while waiting, the error is
// returned immediately without calling the inner sink.
func (s *BackpressureSink) Write(ctx context.Context, entry parser.Entry) error {
	if err := s.valve.Acquire(ctx); err != nil {
		return fmt.Errorf("backpressure_sink: acquire: %w", err)
	}
	defer s.valve.Release()
	return s.inner.Write(ctx, entry)
}

// WaitIdle blocks until all in-flight writes have completed.
func (s *BackpressureSink) WaitIdle(ctx context.Context) error {
	return s.valve.WaitIdle(ctx)
}
