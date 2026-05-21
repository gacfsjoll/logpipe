// Package buffer provides an in-memory ring buffer for log entries that
// smooths out bursts between the pipeline and downstream sinks.
package buffer

import (
	"errors"
	"sync"

	"logpipe/internal/parser"
)

// ErrFull is returned by Push when the buffer has reached its capacity and
// the caller chose a non-blocking strategy.
var ErrFull = errors.New("buffer: ring buffer is full")

// RingBuffer holds up to Cap log entries in a fixed-size circular queue.
type RingBuffer struct {
	mu    sync.Mutex
	items []*parser.Entry
	head  int
	tail  int
	count int
	cap   int
}

// New creates a RingBuffer with the given capacity. Panics if cap < 1.
func New(cap int) *RingBuffer {
	if cap < 1 {
		panic("buffer: capacity must be at least 1")
	}
	return &RingBuffer{
		items: make([]*parser.Entry, cap),
		cap:   cap,
	}
}

// Push adds an entry to the tail of the buffer. Returns ErrFull if the
// buffer is at capacity.
func (r *RingBuffer) Push(e *parser.Entry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.count == r.cap {
		return ErrFull
	}
	r.items[r.tail] = e
	r.tail = (r.tail + 1) % r.cap
	r.count++
	return nil
}

// Pop removes and returns the oldest entry from the buffer. Returns nil
// when the buffer is empty.
func (r *RingBuffer) Pop() *parser.Entry {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.count == 0 {
		return nil
	}
	e := r.items[r.head]
	r.items[r.head] = nil
	r.head = (r.head + 1) % r.cap
	r.count--
	return e
}

// Len returns the number of entries currently held in the buffer.
func (r *RingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

// Cap returns the maximum number of entries the buffer can hold.
func (r *RingBuffer) Cap() int {
	return r.cap
}
