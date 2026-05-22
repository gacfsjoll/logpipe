package backpressure_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/logpipe/logpipe/internal/backpressure"
)

func TestNew_PanicOnZeroLimit(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for limit=0")
		}
	}()
	backpressure.New(0)
}

func TestAcquireRelease_TokensRecycled(t *testing.T) {
	t.Parallel()
	v := backpressure.New(2)
	ctx := context.Background()

	if err := v.Acquire(ctx); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if err := v.Acquire(ctx); err != nil {
		t.Fatalf("second acquire: %v", err)
	}

	s := v.Stats()
	if s.InFlight != 2 {
		t.Fatalf("want InFlight=2, got %d", s.InFlight)
	}

	v.Release()
	v.Release()

	s = v.Stats()
	if s.InFlight != 0 {
		t.Fatalf("want InFlight=0, got %d", s.InFlight)
	}
}

func TestAcquire_BlocksWhenFull(t *testing.T) {
	t.Parallel()
	v := backpressure.New(1)
	ctx := context.Background()

	if err := v.Acquire(ctx); err != nil {
		t.Fatal(err)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()

	err := v.Acquire(ctxTimeout)
	if err == nil {
		t.Fatal("expected ErrCapacityExceeded")
	}
	if err != backpressure.ErrCapacityExceeded {
		t.Fatalf("unexpected error: %v", err)
	}

	s := v.Stats()
	if s.Dropped != 1 {
		t.Fatalf("want Dropped=1, got %d", s.Dropped)
	}
}

func TestWaitIdle_ReturnsWhenInFlightReachesZero(t *testing.T) {
	t.Parallel()
	v := backpressure.New(3)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := v.Acquire(ctx); err != nil {
			t.Fatal(err)
		}
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		v.Release()
		v.Release()
		v.Release()
	}()

	ctxWait, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	if err := v.WaitIdle(ctxWait); err != nil {
		t.Fatalf("WaitIdle: %v", err)
	}
}

func TestConcurrentAcquireRelease(t *testing.T) {
	t.Parallel()
	v := backpressure.New(10)
	ctx := context.Background()
	const workers = 50

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := v.Acquire(ctx); err != nil {
				return
			}
			time.Sleep(2 * time.Millisecond)
			v.Release()
		}()
	}
	wg.Wait()

	s := v.Stats()
	if s.InFlight != 0 {
		t.Fatalf("want InFlight=0 after all goroutines finish, got %d", s.InFlight)
	}
	if s.Acquired != s.Released {
		t.Fatalf("acquired=%d != released=%d", s.Acquired, s.Released)
	}
}
