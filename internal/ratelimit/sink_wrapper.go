package ratelimit

import (
	"context"
	"fmt"

	"logpipe/internal/parser"
)

// Sink is the interface satisfied by any log sink.
type Sink interface {
	Write(ctx context.Context, entry parser.Entry) error
}

// RateLimitedSink wraps a Sink and enforces a token-bucket rate limit on
// every Write call. If the context is cancelled while waiting for a token
// the error is propagated to the caller.
type RateLimitedSink struct {
	inner   Sink
	limiter *Limiter
}

// NewRateLimitedSink returns a RateLimitedSink that allows at most cfg.Rate
// writes per second with an initial burst of cfg.Burst.
func NewRateLimitedSink(inner Sink, cfg Config) (*RateLimitedSink, error) {
	if inner == nil {
		return nil, fmt.Errorf("ratelimit: inner sink must not be nil")
	}
	lim, err := New(cfg)
	if err != nil {
		return nil, fmt.Errorf("ratelimit: %w", err)
	}
	return &RateLimitedSink{inner: inner, limiter: lim}, nil
}

// Write blocks until a token is available (or ctx is done) and then
// delegates to the wrapped sink.
func (s *RateLimitedSink) Write(ctx context.Context, entry parser.Entry) error {
	if err := s.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("ratelimit: %w", err)
	}
	return s.inner.Write(ctx, entry)
}

// Stats returns a snapshot of the underlying limiter counters.
func (s *RateLimitedSink) Stats() Snapshot {
	return s.limiter.Stats()
}
