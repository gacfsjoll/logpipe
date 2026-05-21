// Package retry provides a configurable exponential back-off retry helper
// used by sinks to recover from transient write failures.
package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// ErrExhausted is returned when all retry attempts have been consumed.
var ErrExhausted = errors.New("retry: all attempts exhausted")

// Config holds the parameters that control retry behaviour.
type Config struct {
	// MaxAttempts is the total number of tries (including the first).
	MaxAttempts int
	// BaseDelay is the wait time before the second attempt.
	BaseDelay time.Duration
	// MaxDelay caps the computed back-off interval.
	MaxDelay time.Duration
	// Multiplier scales the delay on each successive failure (default 2.0).
	Multiplier float64
}

// Retryer executes an operation with exponential back-off.
type Retryer struct {
	cfg Config
}

// New returns a Retryer with the supplied configuration.
// Multiplier is forced to at least 1.0 to avoid shrinking delays.
func New(cfg Config) *Retryer {
	if cfg.Multiplier < 1.0 {
		cfg.Multiplier = 2.0
	}
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	return &Retryer{cfg: cfg}
}

// Do calls fn up to MaxAttempts times. It stops early when ctx is cancelled
// or fn returns nil. If every attempt fails it returns ErrExhausted wrapping
// the last error.
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < r.cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if attempt == r.cfg.MaxAttempts-1 {
			break
		}
		delay := r.delay(attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return errors.Join(ErrExhausted, lastErr)
}

// delay returns the back-off duration for the given zero-based attempt index.
func (r *Retryer) delay(attempt int) time.Duration {
	d := float64(r.cfg.BaseDelay) * math.Pow(r.cfg.Multiplier, float64(attempt))
	if r.cfg.MaxDelay > 0 && time.Duration(d) > r.cfg.MaxDelay {
		return r.cfg.MaxDelay
	}
	return time.Duration(d)
}
