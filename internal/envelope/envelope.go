// Package envelope wraps a parsed log entry with routing metadata before
// it is forwarded to downstream sinks.  The envelope carries the source
// file path, a monotonically-increasing sequence number, and the wall-clock
// time at which the entry was received by the pipeline.
package envelope

import (
	"sync/atomic"
	"time"

	"logpipe/internal/parser"
)

// Envelope decorates a [parser.Entry] with pipeline-level metadata.
type Envelope struct {
	// Sequence is a process-wide monotonic counter assigned at wrap time.
	Sequence uint64
	// Source is the absolute path of the file the entry originated from.
	Source string
	// ReceivedAt is the wall-clock time the entry entered the pipeline.
	ReceivedAt time.Time
	// Entry is the underlying parsed log entry.
	Entry parser.Entry
}

// global sequence counter; starts at zero, incremented atomically.
var seq atomic.Uint64

// Wrap creates an Envelope for entry originating from source, assigning the
// next sequence number and stamping ReceivedAt with the current UTC time.
func Wrap(source string, entry parser.Entry) Envelope {
	return Envelope{
		Sequence:   seq.Add(1),
		Source:     source,
		ReceivedAt: time.Now().UTC(),
		Entry:      entry,
	}
}

// ResetSequence resets the global sequence counter to zero.  It is intended
// only for use in tests where deterministic sequence numbers are required.
func ResetSequence() {
	seq.Store(0)
}
