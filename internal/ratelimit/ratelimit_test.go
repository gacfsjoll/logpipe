package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"logpipe/internal/ratelimit"
)

func TestNew_AllowsBurst(t *testing.T) {
	l := ratelimit.New(10, 5)
	ctx := context.Background()

	// All burst tokens should be consumed immediately.
	for i := 0; i < 5; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("burst token %d: unexpected error: %v", i, err)
		}
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	// Rate of 0.001 tokens/sec means we will never get a token in time.
	l := ratelimit.New(0.001, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

func TestWait_RefillsOverTime(t *testing.T) {
	// 100 tokens/sec, burst 1 — consume the burst, then wait for refill.
	l := ratelimit.New(100, 1)
	ctx := context.Background()

	// Consume the single burst token.
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// At 100 tok/s the next token arrives in ~10ms; give it 200ms.
	ctx2, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := l.Wait(ctx2); err != nil {
		t.Fatalf("expected token to be available after refill: %v", err)
	}
}

func TestWait_ZeroRateBlocksUntilCancel(t *testing.T) {
	l := ratelimit.New(0, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Fatal("expected context deadline exceeded, got nil")
	}
}

func TestWait_ConcurrentConsumers(t *testing.T) {
	const workers = 10
	l := ratelimit.New(1000, workers)
	ctx := context.Background()

	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			errs <- l.Wait(ctx)
		}()
	}
	for i := 0; i < workers; i++ {
		if err := <-errs; err != nil {
			t.Errorf("worker error: %v", err)
		}
	}
}
