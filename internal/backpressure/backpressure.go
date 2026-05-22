// Package backpressure provides a token-based flow-control mechanism that
// signals upstream producers to slow down when downstream sinks are falling
// behind. A Valve exposes a single Acquire/Release pair; producers call
// Acquire before forwarding a log entry and Release once the entry has been
// handed off to a sink.
package backpressure

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ErrCapacityExceeded is returned by Acquire when the context deadline is
// reached before a token becomes available.
var ErrCapacityExceeded = errors.New("backpressure: capacity exceeded, entry dropped")

// Stats is a point-in-time snapshot of valve counters.
type Stats struct {
	Acquired uint64
	Released uint64
	Dropped  uint64
	InFlight int64
}

// Valve controls the number of log entries that may be in-flight at once.
type Valve struct {
	tokens   chan struct{}
	acquired atomic.Uint64
	released atomic.Uint64
	dropped  atomic.Uint64
}

// New creates a Valve with the given concurrency limit. New panics if limit
// is less than 1.
func New(limit int) *Valve {
	if limit < 1 {
		panic("backpressure: limit must be >= 1")
	}
	v := &Valve{tokens: make(chan struct{}, limit)}
	for i := 0; i < limit; i++ {
		v.tokens <- struct{}{}
	}
	return v
}

// Acquire blocks until a token is available or ctx is cancelled. If the
// context expires before a token can be obtained, ErrCapacityExceeded is
// returned and the dropped counter is incremented.
func (v *Valve) Acquire(ctx context.Context) error {
	select {
	case <-v.tokens:
		v.acquired.Add(1)
		return nil
	case <-ctx.Done():
		v.dropped.Add(1)
		return ErrCapacityExceeded
	}
}

// Release returns a token to the pool. It must be called exactly once per
// successful Acquire.
func (v *Valve) Release() {
	v.released.Add(1)
	v.tokens <- struct{}{}
}

// Stats returns a consistent snapshot of the valve counters.
func (v *Valve) Stats() Stats {
	acq := v.acquired.Load()
	rel := v.released.Load()
	return Stats{
		Acquired: acq,
		Released: rel,
		Dropped:  v.dropped.Load(),
		InFlight: int64(acq) - int64(rel),
	}
}

// WaitIdle blocks until all acquired tokens have been released or ctx is
// cancelled. It polls at 5 ms intervals and is intended for graceful shutdown.
func (v *Valve) WaitIdle(ctx context.Context) error {
	for {
		if v.Stats().InFlight == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
}
