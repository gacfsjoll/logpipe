package ratelimit_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/ratelimit"
)

// stubSink counts Write calls and optionally returns an error.
type stubSink struct {
	calls atomic.Int64
	err   error
}

func (s *stubSink) Write(_ context.Context, _ parser.Entry) error {
	s.calls.Add(1)
	return s.err
}

func TestNewRateLimitedSink_NilInnerReturnsError(t *testing.T) {
	_, err := ratelimit.NewRateLimitedSink(nil, ratelimit.Config{Rate: 10, Burst: 1})
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewRateLimitedSink_InvalidConfigReturnsError(t *testing.T) {
	_, err := ratelimit.NewRateLimitedSink(&stubSink{}, ratelimit.Config{Rate: -1, Burst: 1})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestRateLimitedSink_Write_ForwardsToInner(t *testing.T) {
	inner := &stubSink{}
	s, err := ratelimit.NewRateLimitedSink(inner, ratelimit.Config{Rate: 100, Burst: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	entry := parser.Entry{"message": "hello"}

	if err := s.Write(ctx, entry); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestRateLimitedSink_Write_RespectsRateLimit(t *testing.T) {
	inner := &stubSink{}
	// Rate of 5/s, burst of 1 — writing 3 entries should take at least ~400 ms.
	s, err := ratelimit.NewRateLimitedSink(inner, ratelimit.Config{Rate: 5, Burst: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	entry := parser.Entry{"message": "tick"}

	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := s.Write(ctx, entry); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	if elapsed < 300*time.Millisecond {
		t.Errorf("expected rate limiting delay, got %v", elapsed)
	}
}

func TestRateLimitedSink_Write_ContextCancelledReturnsError(t *testing.T) {
	inner := &stubSink{}
	// Zero rate — every Wait will block until context is cancelled.
	s, err := ratelimit.NewRateLimitedSink(inner, ratelimit.Config{Rate: 0, Burst: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := s.Write(ctx, parser.Entry{}); err == nil {
		t.Fatal("expected context cancellation error")
	}
	if inner.calls.Load() != 0 {
		t.Errorf("inner sink should not have been called")
	}
}
