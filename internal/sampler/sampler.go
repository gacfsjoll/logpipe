// Package sampler provides probabilistic log sampling to reduce
// the volume of log entries forwarded to downstream sinks.
//
// A Sampler accepts a rate in the range (0, 1] where 1.0 means
// "keep every entry" and 0.1 means "keep roughly 10% of entries".
package sampler

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
)

// ErrInvalidRate is returned when the sampling rate is outside (0, 1].
var ErrInvalidRate = errors.New("sampler: rate must be in the range (0, 1]")

// Sampler decides probabilistically whether a log line should be kept.
type Sampler struct {
	rate    float64
	mu      sync.Mutex
	rng     *rand.Rand
	kept    atomic.Int64
	dropped atomic.Int64
}

// New creates a Sampler with the given keep rate.
// rate must satisfy 0 < rate <= 1.
func New(rate float64, seed int64) (*Sampler, error) {
	if rate <= 0 || rate > 1 {
		return nil, ErrInvalidRate
	}
	return &Sampler{
		rate: rate,
		//nolint:gosec // non-cryptographic sampling is intentional
		rng: rand.New(rand.NewSource(seed)),
	}, nil
}

// Keep returns true if the entry should be forwarded.
func (s *Sampler) Keep() bool {
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()

	if v < s.rate {
		s.kept.Add(1)
		return true
	}
	s.dropped.Add(1)
	return false
}

// Stats returns the total number of entries kept and dropped since
// the Sampler was created or last reset.
func (s *Sampler) Stats() (kept, dropped int64) {
	return s.kept.Load(), s.dropped.Load()
}

// Reset zeroes the kept/dropped counters.
func (s *Sampler) Reset() {
	s.kept.Store(0)
	s.dropped.Store(0)
}
