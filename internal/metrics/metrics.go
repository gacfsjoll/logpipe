// Package metrics provides lightweight counters for tracking pipeline activity.
package metrics

import (
	"sync/atomic"
	"time"
)

// Counters holds atomic counters for key pipeline events.
type Counters struct {
	LinesRead      atomic.Int64
	LinesParsed    atomic.Int64
	ParseErrors    atomic.Int64
	SinkWrites     atomic.Int64
	SinkErrors     atomic.Int64
	StartTime      time.Time
}

// New returns a new Counters instance with StartTime set to now.
func New() *Counters {
	return &Counters{
		StartTime: time.Now(),
	}
}

// Snapshot is a point-in-time copy of the counters.
type Snapshot struct {
	LinesRead   int64
	LinesParsed int64
	ParseErrors int64
	SinkWrites  int64
	SinkErrors  int64
	Uptime      time.Duration
}

// Snapshot returns a consistent read of all counters.
func (c *Counters) Snapshot() Snapshot {
	return Snapshot{
		LinesRead:   c.LinesRead.Load(),
		LinesParsed: c.LinesParsed.Load(),
		ParseErrors: c.ParseErrors.Load(),
		SinkWrites:  c.SinkWrites.Load(),
		SinkErrors:  c.SinkErrors.Load(),
		Uptime:      time.Since(c.StartTime),
	}
}

// Reset zeroes all counters but preserves StartTime.
func (c *Counters) Reset() {
	c.LinesRead.Store(0)
	c.LinesParsed.Store(0)
	c.ParseErrors.Store(0)
	c.SinkWrites.Store(0)
	c.SinkErrors.Store(0)
}
