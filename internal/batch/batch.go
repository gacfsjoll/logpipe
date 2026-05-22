// Package batch provides a size- and time-bounded accumulator that groups
// individual log entries into slices before forwarding them to a flush
// function. This reduces the number of outbound calls made by sinks that
// benefit from bulk delivery (e.g. HTTP endpoints that accept JSON arrays).
package batch

import (
	"context"
	"sync"
	"time"

	"logpipe/internal/parser"
)

// FlushFunc is called with a non-empty batch of entries whenever the batcher
// decides it is time to flush.
type FlushFunc func(entries []parser.Entry)

// Batcher accumulates entries and flushes them when either the maximum batch
// size is reached or the flush interval elapses, whichever comes first.
type Batcher struct {
	mu       sync.Mutex
	buf      []parser.Entry
	maxSize  int
	interval time.Duration
	flush    FlushFunc
	timer    *time.Timer
}

// New creates a Batcher. maxSize must be >= 1; interval must be > 0.
func New(maxSize int, interval time.Duration, flush FlushFunc) *Batcher {
	if maxSize < 1 {
		panic("batch: maxSize must be >= 1")
	}
	if interval <= 0 {
		panic("batch: interval must be > 0")
	}
	if flush == nil {
		panic("batch: flush func must not be nil")
	}
	b := &Batcher{
		buf:     make([]parser.Entry, 0, maxSize),
		maxSize: maxSize,
		interval: interval,
		flush:   flush,
	}
	b.resetTimer()
	return b
}

// Add appends an entry to the current batch. If the batch reaches maxSize it
// is flushed immediately.
func (b *Batcher) Add(e parser.Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, e)
	if len(b.buf) >= b.maxSize {
		b.flushLocked()
	}
}

// Run blocks until ctx is cancelled, driving the interval-based flush ticker.
// Call this in a goroutine alongside the rest of the pipeline.
func (b *Batcher) Run(ctx context.Context) {
	for {
		select {
		case <-b.timer.C:
			b.mu.Lock()
			b.flushLocked()
			b.mu.Unlock()
		case <-ctx.Done():
			b.mu.Lock()
			b.flushLocked()
			b.mu.Unlock()
			return
		}
	}
}

// flushLocked sends the current buffer to the flush function and resets state.
// Caller must hold b.mu.
func (b *Batcher) flushLocked() {
	if len(b.buf) == 0 {
		b.resetTimer()
		return
	}
	batch := make([]parser.Entry, len(b.buf))
	copy(batch, b.buf)
	b.buf = b.buf[:0]
	b.resetTimer()
	// Release lock while calling user code.
	b.mu.Unlock()
	b.flush(batch)
	b.mu.Lock()
}

func (b *Batcher) resetTimer() {
	if b.timer != nil {
		b.timer.Stop()
	}
	b.timer = time.NewTimer(b.interval)
}
