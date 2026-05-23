// Package sink provides output destinations for parsed log entries.
//
// # SampledSink
//
// SampledSink wraps any Sink implementation and performs probabilistic
// sampling before forwarding entries to the inner sink. This is useful for
// high-volume pipelines where only a representative fraction of log lines
// needs to reach an expensive downstream (e.g. an HTTP endpoint).
//
// Configuration
//
//	Rate  float64  // fraction of entries to keep, must be in (0, 1]
//	Seed  int64    // optional fixed PRNG seed; 0 = random
//
// Example YAML
//
//	sinks:
//	  - type: sampled
//	    rate: 0.1          # keep ~10 % of log lines
//	    inner:
//	      type: http
//	      url: https://logs.example.com/ingest
//
// The sampler is safe for concurrent use. Dropped entries are silently
// discarded; no error is returned for them. Errors returned by the inner
// sink are propagated to the caller unchanged.
package sink
