// Package backpressure implements a token-based flow-control valve for
// logpipe pipelines.
//
// # Overview
//
// When downstream sinks are slow (e.g. a remote HTTP endpoint with high
// latency), unbounded in-flight entries can exhaust memory. A Valve caps the
// number of entries that may be simultaneously in-flight between the tailer
// and the sink layer.
//
// # Usage
//
//	v := backpressure.New(100) // at most 100 entries in-flight
//
//	// producer side
//	if err := v.Acquire(ctx); err != nil {
//	    // context expired — entry is dropped and counter incremented
//	    return err
//	}
//
//	// consumer side (after sink.Write returns)
//	v.Release()
//
// # Graceful shutdown
//
// Call WaitIdle to drain all in-flight entries before exiting:
//
//	_ = v.WaitIdle(shutdownCtx)
package backpressure
