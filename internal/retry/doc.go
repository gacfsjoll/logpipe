// Package retry implements a generic exponential back-off retry helper.
//
// # Overview
//
// Sinks and other I/O components may encounter transient failures (network
// timeouts, temporary unavailability, etc.). The retry package provides a
// lightweight, context-aware mechanism to re-attempt an operation with
// increasing delays between tries.
//
// # Usage
//
//	r := retry.New(retry.Config{
//		MaxAttempts: 5,
//		BaseDelay:   100 * time.Millisecond,
//		MaxDelay:    10 * time.Second,
//		Multiplier:  2.0,
//	})
//
//	err := r.Do(ctx, func() error {
//		return sink.Write(entry)
//	})
//
// If all attempts fail, Do returns retry.ErrExhausted joined with the last
// underlying error so callers can use errors.Is for both.
package retry
