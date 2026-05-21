package retry_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"logpipe/internal/retry"
)

var errFake = errors.New("fake error")

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	r := retry.New(retry.Config{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, Multiplier: 2})
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesAndSucceeds(t *testing.T) {
	r := retry.New(retry.Config{MaxAttempts: 5, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, Multiplier: 2})
	var calls int32
	err := r.Do(context.Background(), func() error {
		if atomic.AddInt32(&calls, 1) < 3 {
			return errFake
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAllAttempts(t *testing.T) {
	r := retry.New(retry.Config{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, Multiplier: 2})
	var calls int32
	err := r.Do(context.Background(), func() error {
		atomic.AddInt32(&calls, 1)
		return errFake
	})
	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
	if !errors.Is(err, errFake) {
		t.Fatalf("expected wrapped errFake, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCancelledBetweenRetries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := retry.New(retry.Config{MaxAttempts: 10, BaseDelay: 50 * time.Millisecond, MaxDelay: 200 * time.Millisecond, Multiplier: 2})
	var calls int32
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	err := r.Do(ctx, func() error {
		atomic.AddInt32(&calls, 1)
		return errFake
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls == 0 {
		t.Fatal("expected at least one call before cancellation")
	}
}

func TestDo_MultiplierDefaultsWhenBelowOne(t *testing.T) {
	// Multiplier < 1 should be corrected to 2.0 — verify no panic and retries work.
	r := retry.New(retry.Config{MaxAttempts: 3, BaseDelay: time.Millisecond, Multiplier: 0})
	var calls int32
	_ = r.Do(context.Background(), func() error {
		atomic.AddInt32(&calls, 1)
		return errFake
	})
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}
