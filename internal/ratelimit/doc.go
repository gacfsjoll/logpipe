// Package ratelimit implements a thread-safe token-bucket rate limiter
// for use in the logpipe pipeline.
//
// A Limiter is created with a sustained rate (tokens per second) and a
// burst capacity. Callers invoke Wait to block until a token is available
// or the supplied context is cancelled.
//
// Typical usage inside a pipeline stage:
//
//	limiter := ratelimit.New(500, 50) // 500 lines/sec, burst of 50
//	for line := range lines {
//		if err := limiter.Wait(ctx); err != nil {
//			return err
//		}
//		process(line)
//	}
package ratelimit
