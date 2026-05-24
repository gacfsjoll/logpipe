// Package sink provides composable Sink implementations for logpipe.
//
// # BackpressureSink
//
// BackpressureSink wraps any Sink and enforces a maximum number of concurrent
// in-flight Write calls. When the limit is reached, subsequent callers block
// until a slot is freed or their context is cancelled.
//
// This is useful when the inner sink communicates over the network and you want
// to bound the number of simultaneous open connections / goroutines, preventing
// unbounded resource growth during traffic spikes.
//
// Example usage:
//
//	inner, _ := sink.NewHTTPSink(sink.HTTPConfig{URL: "https://logs.example.com/ingest"})
//	s, err := sink.NewBackpressureSink(inner, sink.BackpressureConfig{Limit: 8})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Optionally wait for all in-flight writes to drain before shutdown.
//	_ = s.WaitIdle(ctx)
package sink
