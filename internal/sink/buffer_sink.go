package sink

import (
	"errors"
	"fmt"

	"logpipe/internal/buffer"
	"logpipe/internal/parser"
)

// BufferedSink wraps an inner Sink with an in-memory ring buffer. Entries are
// accumulated in the buffer and flushed to the inner sink in batches when
// Flush is called, or individually as the buffer fills.
type BufferedSink struct {
	inner  Sink
	buf    *buffer.Buffer
	cap    int
}

// NewBufferedSink creates a BufferedSink that buffers up to capacity entries
// before forwarding them to inner.
func NewBufferedSink(inner Sink, capacity int) (*BufferedSink, error) {
	if inner == nil {
		return nil, errors.New("buffer_sink: inner sink must not be nil")
	}
	if capacity <= 0 {
		return nil, fmt.Errorf("buffer_sink: capacity must be > 0, got %d", capacity)
	}
	return &BufferedSink{
		inner: inner,
		buf:   buffer.New(capacity),
		cap:   capacity,
	}, nil
}

// Write enqueues the entry into the ring buffer. If the buffer is full the
// oldest entry is dropped and the new one is appended.
func (s *BufferedSink) Write(entry parser.Entry) error {
	if err := s.buf.Push(entry); err != nil {
		// Buffer full — drop oldest to make room.
		s.buf.Pop()
		_ = s.buf.Push(entry)
	}
	return nil
}

// Flush drains all buffered entries to the inner sink. The first error
// encountered is returned; remaining entries stay in the buffer.
func (s *BufferedSink) Flush() error {
	for {
		e := s.buf.Pop()
		if e == nil {
			break
		}
		if err := s.inner.Write(*e); err != nil {
			return fmt.Errorf("buffer_sink: flush: %w", err)
		}
	}
	return nil
}

// Len returns the number of entries currently held in the buffer.
func (s *BufferedSink) Len() int {
	return s.buf.Len()
}
