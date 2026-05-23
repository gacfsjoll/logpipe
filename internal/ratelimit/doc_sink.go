// Package ratelimit provides a token-bucket rate limiter and a sink wrapper
// that enforces per-sink write rate limits.
//
// # Limiter
//
// New constructs a Limiter from a Config that specifies a steady-state Rate
// (tokens per second) and an initial Burst allowance. Callers invoke Wait to
// consume a token, blocking until one is available or the context is
// cancelled.
//
// # RateLimitedSink
//
// NewRateLimitedSink wraps any Sink implementation and gates every Write call
// through the limiter. This is useful when a downstream HTTP or file sink
// must not be overwhelmed, or when a per-tenant quota must be enforced.
//
// Example:
//
//	s, _ := ratelimit.NewRateLimitedSink(httpSink, ratelimit.Config{
//		Rate:  100, // 100 writes per second
//		Burst: 20,  // allow short bursts of up to 20
//	})
//	_ = s.Write(ctx, entry)
package ratelimit
