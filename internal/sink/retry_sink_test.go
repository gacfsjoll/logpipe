package sink_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"logpipe/internal/parser"
	"logpipe/internal/retry"
	"logpipe/internal/sink"
)

// failNSink is a Sink that fails the first n calls then succeeds.
type failNSink struct {
	failures int32
	calls    atomic.Int32
}

func (f *failNSink) Write(_ context.Context, _ parser.Entry) error {
	if int(f.calls.Add(1)) <= int(f.failures) {
		return errors.New("transient error")
	}
	return nil
}

func retryCfg(attempts int) retry.Config {
	return retry.Config{
		MaxAttempts: attempts,
		BaseDelay:   time.Millisecond,
		Multiplier:  1.0,
	}
}

func TestNewRetrySink_NilInnerReturnsError(t *testing.T) {
	_, err := sink.NewRetrySink(nil, retryCfg(3))
	if err == nil {
		t.Fatal("expected error for nil inner sink")
	}
}

func TestNewRetrySink_InvalidConfigReturnsError(t *testing.T) {
	inner := &failNSink{}
	_, err := sink.NewRetrySink(inner, retry.Config{MaxAttempts: 0})
	if err == nil {
		t.Fatal("expected error for invalid retry config")
	}
}

func TestRetrySink_Write_SucceedsFirstAttempt(t *testing.T) {
	inner := &failNSink{failures: 0}
	s, err := sink.NewRetrySink(inner, retryCfg(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Write(context.Background(), parser.Entry{}); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestRetrySink_Write_RetriesAndSucceeds(t *testing.T) {
	inner := &failNSink{failures: 2}
	s, err := sink.NewRetrySink(inner, retryCfg(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Write(context.Background(), parser.Entry{}); err != nil {
		t.Fatalf("expected eventual success, got: %v", err)
	}
	if got := inner.calls.Load(); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestRetrySink_Write_ExhaustsRetries(t *testing.T) {
	inner := &failNSink{failures: 10}
	s, err := sink.NewRetrySink(inner, retryCfg(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Write(context.Background(), parser.Entry{}); err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if got := inner.calls.Load(); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestRetrySink_Write_ContextCancelled(t *testing.T) {
	inner := &failNSink{failures: 10}
	s, err := sink.NewRetrySink(inner, retryCfg(10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Write(ctx, parser.Entry{}); err == nil {
		t.Fatal("expected error when context is already cancelled")
	}
}
