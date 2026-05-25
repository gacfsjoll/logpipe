// Package sink provides Sink implementations for logpipe.
//
// # CheckpointSink
//
// CheckpointSink is a decorating sink that records a delivery offset into a
// persistent checkpoint store after each successful write to the wrapped inner
// sink.
//
// On a clean restart the pipeline can read the stored offset and instruct the
// tailer to skip log lines that have already been delivered, providing
// at-least-once delivery semantics with minimal re-processing.
//
// Usage:
//
//	cp, _ := checkpoint.New("/var/lib/logpipe/state.json")
//	inner, _ := sink.NewHTTPSink(cfg)
//	s, _ := sink.NewCheckpointSink(inner, cp, "/var/log/app.log")
//
// The offset stored is the RFC3339Nano UTC representation of the entry
// timestamp, which is sufficient for most log sources that emit monotonically
// increasing timestamps.
package sink
