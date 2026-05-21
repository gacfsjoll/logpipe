// Package buffer implements a thread-safe, fixed-capacity ring buffer for
// [parser.Entry] values.
//
// The ring buffer is used by the pipeline to absorb short bursts of log
// lines without blocking the tailer goroutines. When the buffer is full
// the caller receives [ErrFull] and can choose to drop the entry or apply
// back-pressure.
//
// # Usage
//
//	b := buffer.New(1024)
//
//	// producer
//	if err := b.Push(entry); err == buffer.ErrFull {
//	    // handle overflow
//	}
//
//	// consumer
//	if e := b.Pop(); e != nil {
//	    // process e
//	}
package buffer
