package sink

import (
	"context"
	"errors"
	"fmt"

	"logpipe/internal/parser"
	"logpipe/internal/retry"
)

// RetrySink wraps any Sink and retries failed writes according to the
// provided retry.Config. Each call to Write is attempted up to
// Config.MaxAttempts times before the error is surfaced to the caller.
type RetrySink struct {
	inner  Sink
	retrier *retry.Retrier
}

// NewRetrySink returns a RetrySink that delegates writes to inner and retries
// on transient failures. Returns an error when inner is nil or the retry
// configuration is invalid.
func NewRetrySink(inner Sink, cfg retry.Config) (*RetrySink, error) {
	if inner == nil {
		return nil, errors.New("retry_sink: inner sink must not be nil")
	}
	r, err := retry.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("retry_sink: %w", err)
	}
	return &RetrySink{inner: inner, retrier: r}, nil
}

// Write forwards entry to the inner sink, retrying on failure according to
// the configured policy. The context is forwarded so that in-flight retries
// are cancelled when the context is done.
func (s *RetrySink) Write(ctx context.Context, entry parser.Entry) error {
	return s.retrier.Do(ctx, func() error {
		return s.inner.Write(ctx, entry)
	})
}
