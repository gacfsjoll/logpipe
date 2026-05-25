// Package sink provides a collection of Sink implementations and
// decorator sinks that compose around an inner Sink.
//
// # TruncateSink
//
// TruncateSink is a decorator that truncates string field values exceeding
// a per-field byte limit before forwarding the entry to the inner sink.
// This is useful when downstream systems impose payload size constraints
// (e.g. HTTP body limits or database column widths).
//
// Truncated values are suffixed with "..." so consumers can detect that
// data was elided. The max-bytes threshold must therefore be greater than
// three.
//
// Example:
//
//	inner := sink.NewStdoutSink()
//	ts, err := sink.NewTruncateSink(inner, map[string]int{
//		"message": 256,
//		"stack_trace": 1024,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
package sink
