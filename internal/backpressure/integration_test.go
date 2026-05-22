package backpressure_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/logpipe/logpipe/internal/backpressure"
)

// TestValve_PipelineSimulation models a producer/consumer pair where the
// consumer is intentionally slower than the producer. The valve should cap
// concurrency and the dropped counter should reflect any entries that could
// not be acquired within the producer's deadline.
func TestValve_PipelineSimulation(t *testing.T) {
	t.Parallel()

	const limit = 5
	const produced = 20
	const consumerDelay = 15 * time.Millisecond
	const acquireTimeout = 10 * time.Millisecond

	v := backpressure.New(limit)

	var processed atomic.Int64
	var dropped atomic.Int64

	for i := 0; i < produced; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), acquireTimeout)
		err := v.Acquire(ctx)
		cancel()
		if err != nil {
			dropped.Add(1)
			continue
		}
		go func() {
			time.Sleep(consumerDelay)
			processed.Add(1)
			v.Release()
		}()
	}

	// Wait for all in-flight entries to drain.
	shutCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := v.WaitIdle(shutCtx); err != nil {
		t.Fatalf("WaitIdle: %v", err)
	}

	s := v.Stats()
	total := processed.Load() + dropped.Load()
	if int(total) != produced {
		t.Fatalf("want total=%d, got %d (processed=%d dropped=%d)",
			produced, total, processed.Load(), dropped.Load())
	}
	if s.InFlight != 0 {
		t.Fatalf("want InFlight=0, got %d", s.InFlight)
	}
	if s.Dropped != uint64(dropped.Load()) {
		t.Fatalf("stats.Dropped=%d does not match local dropped=%d",
			s.Dropped, dropped.Load())
	}
}
