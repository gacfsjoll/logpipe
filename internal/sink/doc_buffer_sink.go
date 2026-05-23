// Package sink provides output adapters (sinks) for parsed log entries.
//
// BufferedSink
//
// BufferedSink wraps any Sink implementation with a fixed-capacity in-memory
// ring buffer. Writes are accepted immediately without blocking the caller;
// the buffer is drained to the inner sink by calling Flush.
//
// When the buffer reaches capacity the oldest entry is silently dropped to
// make room for the incoming one, preserving recency at the cost of
// completeness. This trade-off is intentional for high-throughput scenarios
// where backpressure from the downstream sink would otherwise stall the
// pipeline.
//
// Typical usage:
//
//	bs, err := sink.NewBufferedSink(httpSink, 1000)
//	// ... write entries during normal operation ...
//	_ = bs.Write(entry)
//	// ... periodically or on shutdown ...
//	_ = bs.Flush()
package sink
